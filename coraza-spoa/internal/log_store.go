package internal

import (
	"context"
	"runtime"
	"sync"
	"sync/atomic"
	"time"

	"github.com/HUAHUAI23/simple-waf/pkg/model"
	"github.com/rs/zerolog"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

// LogStore 定义日志存储接口
type LogStore interface {
	Store(log model.WAFLog) error
	Start()
	Close()
}

// LogBatch 日志批次结构，用于对象池
type LogBatch struct {
	logs []interface{}
	size int
}

// 批次对象池，减少内存分配
var batchPool = sync.Pool{
	New: func() interface{} {
		return &LogBatch{
			logs: make([]interface{}, 0, 200), // 预分配更大的容量
		}
	},
}

// RingBuffer 高性能无锁环形缓冲区
type RingBuffer struct {
	buffer   []model.WAFLog
	capacity uint64
	mask     uint64
	_        [56]byte // 缓存行填充，避免false sharing
	head     atomic.Uint64
	_        [56]byte
	tail     atomic.Uint64
	_        [56]byte
}

// NewRingBuffer 创建环形缓冲区
func NewRingBuffer(size int) *RingBuffer {
	// 确保容量是2的幂
	capacity := uint64(1)
	for capacity < uint64(size) {
		capacity <<= 1
	}

	return &RingBuffer{
		buffer:   make([]model.WAFLog, capacity),
		capacity: capacity,
		mask:     capacity - 1,
	}
}

// Push 无锁推送日志
func (rb *RingBuffer) Push(log model.WAFLog) bool {
	for {
		tail := rb.tail.Load()
		next := (tail + 1) & rb.mask
		head := rb.head.Load()

		if next == head {
			return false // 缓冲区满
		}

		if rb.tail.CompareAndSwap(tail, next) {
			rb.buffer[tail] = log
			return true
		}
		// CAS失败，重试
		runtime.Gosched()
	}
}

// PopBatch 批量弹出日志
func (rb *RingBuffer) PopBatch(batch *LogBatch, maxSize int) int {
	head := rb.head.Load()
	tail := rb.tail.Load()

	// 计算可用数量
	available := int((tail - head) & rb.mask)
	if available == 0 {
		return 0
	}

	// 限制批量大小
	if available > maxSize {
		available = maxSize
	}

	// 批量读取
	for i := 0; i < available; i++ {
		idx := (head + uint64(i)) & rb.mask
		batch.logs = append(batch.logs, rb.buffer[idx])
	}

	// 更新head
	rb.head.Add(uint64(available))
	batch.size = available
	return available
}

// MongoLogStore MongoDB实现的高性能日志存储
type MongoLogStore struct {
	mongo       *mongo.Client
	mongoDB     string
	collection  *mongo.Collection
	ringBuffers []*RingBuffer // 多个环形缓冲区，减少竞争
	numBuffers  int
	logger      zerolog.Logger
	state       atomic.Uint32 // 0: stopped, 1: running, 2: closing
	wg          sync.WaitGroup

	// 性能优化参数
	writerCount    int           // 写入协程数
	batchSize      atomic.Int32  // 动态批大小
	batchInterval  time.Duration // 批处理间隔
	maxBatchSize   int
	minBatchSize   int
	bufferSelector atomic.Uint64 // 用于选择buffer
}

// 单例实例
var (
	mongoLogStoreOnce     sync.Once
	mongoLogStoreInstance *MongoLogStore
)

// StoreConfig 存储配置
type StoreConfig struct {
	BufferSize    int
	NumBuffers    int
	WriterCount   int
	BatchInterval time.Duration
	MaxBatchSize  int
	MinBatchSize  int
}

// DefaultConfig 默认配置
func DefaultConfig() StoreConfig {
	cores := runtime.NumCPU()
	return StoreConfig{
		BufferSize:    8192,                   // 每个buffer的大小
		NumBuffers:    cores,                  // buffer数量等于CPU核心数
		WriterCount:   cores / 2,              // 写入协程数为核心数的一半
		BatchInterval: 100 * time.Millisecond, // 100ms批处理间隔
		MaxBatchSize:  500,
		MinBatchSize:  50,
	}
}

// NewMongoLogStore 创建新的MongoDB日志存储器（单例模式）
func NewMongoLogStore(client *mongo.Client, database, collection string, logger zerolog.Logger) *MongoLogStore {
	config := DefaultConfig()
	return NewMongoLogStoreWithConfig(client, database, collection, config, logger)
}

// NewMongoLogStoreWithConfig 使用配置创建存储器
func NewMongoLogStoreWithConfig(client *mongo.Client, database, collection string, config StoreConfig, logger zerolog.Logger) *MongoLogStore {
	mongoLogStoreOnce.Do(func() {
		if config.NumBuffers <= 0 {
			config.NumBuffers = runtime.NumCPU()
		}
		if config.WriterCount <= 0 {
			config.WriterCount = config.NumBuffers / 2
			if config.WriterCount == 0 {
				config.WriterCount = 1
			}
		}

		store := &MongoLogStore{
			mongo:         client,
			mongoDB:       database,
			collection:    client.Database(database).Collection(collection),
			ringBuffers:   make([]*RingBuffer, config.NumBuffers),
			numBuffers:    config.NumBuffers,
			logger:        logger,
			writerCount:   config.WriterCount,
			batchInterval: config.BatchInterval,
			maxBatchSize:  config.MaxBatchSize,
			minBatchSize:  config.MinBatchSize,
		}

		// 初始化环形缓冲区
		bufferSizePerRing := config.BufferSize / config.NumBuffers
		for i := 0; i < config.NumBuffers; i++ {
			store.ringBuffers[i] = NewRingBuffer(bufferSizePerRing)
		}

		// 设置初始批大小
		store.batchSize.Store(int32(config.MinBatchSize))

		mongoLogStoreInstance = store
		logger.Info().
			Int("num_buffers", config.NumBuffers).
			Int("writer_count", config.WriterCount).
			Int("buffer_size", config.BufferSize).
			Msg("创建新的高性能MongoLogStore实例")
	})

	return mongoLogStoreInstance
}

// Store 高性能非阻塞存储
func (s *MongoLogStore) Store(log model.WAFLog) error {
	// 快速检查状态
	if s.state.Load() != 1 {
		return nil // 未运行状态，直接丢弃
	}

	// 选择buffer（轮询方式分散负载）
	idx := s.bufferSelector.Add(1) % uint64(s.numBuffers)
	buffer := s.ringBuffers[idx]

	// 尝试推送到环形缓冲区
	if !buffer.Push(log) {
		// 缓冲区满，直接丢弃（按要求可以接受日志丢失）
		return nil
	}

	return nil
}

// Start 启动日志存储处理
func (s *MongoLogStore) Start() {
	if !s.state.CompareAndSwap(0, 1) {
		s.logger.Debug().Msg("日志处理已在运行")
		return
	}

	s.logger.Info().Msg("启动高性能日志处理")

	// 启动多个写入协程
	for i := 0; i < s.writerCount; i++ {
		s.wg.Add(1)
		go s.writer(i)
	}

	// 启动动态调整协程
	s.wg.Add(1)
	go s.dynamicAdjuster()
}

// writer 写入协程
func (s *MongoLogStore) writer(id int) {
	defer s.wg.Done()

	// 每个writer负责特定的buffers
	buffersPerWriter := s.numBuffers / s.writerCount
	startIdx := id * buffersPerWriter
	endIdx := startIdx + buffersPerWriter
	if id == s.writerCount-1 {
		endIdx = s.numBuffers // 最后一个writer处理剩余的buffers
	}

	ticker := time.NewTicker(s.batchInterval)
	defer ticker.Stop()

	for {
		<-ticker.C
		if s.state.Load() == 2 {
			// 正在关闭，处理剩余数据
			s.flushAll(startIdx, endIdx)
			return
		}
		// 处理分配的buffers
		s.processBatches(startIdx, endIdx)

		// 检查是否需要退出
		if s.state.Load() == 2 {
			return
		}
	}
}

// processBatches 处理批次
func (s *MongoLogStore) processBatches(startIdx, endIdx int) {
	batch := batchPool.Get().(*LogBatch)
	defer func() {
		batch.logs = batch.logs[:0]
		batch.size = 0
		batchPool.Put(batch)
	}()

	currentBatchSize := int(s.batchSize.Load())
	totalCollected := 0

	// 从多个buffer收集日志
	for i := startIdx; i < endIdx && totalCollected < currentBatchSize; i++ {
		buffer := s.ringBuffers[i]
		remaining := currentBatchSize - totalCollected
		count := buffer.PopBatch(batch, remaining)
		totalCollected += count
	}

	// 如果收集到日志，执行批量写入
	if batch.size > 0 {
		s.batchInsert(batch)
	}
}

// batchInsert 批量插入（无重试，追求性能）
func (s *MongoLogStore) batchInsert(batch *LogBatch) {
	if batch.size == 0 {
		return
	}

	// 使用较短的超时，失败就丢弃
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	// 批量插入，忽略错误（可接受丢失）
	_, _ = s.collection.InsertMany(ctx, batch.logs[:batch.size])
}

// dynamicAdjuster 动态调整批大小
func (s *MongoLogStore) dynamicAdjuster() {
	defer s.wg.Done()

	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	lastCheck := time.Now()
	lastCount := uint64(0)

	for {
		<-ticker.C
		if s.state.Load() == 2 {
			return
		}

		// 计算当前负载
		currentCount := uint64(0)
		for _, buffer := range s.ringBuffers {
			head := buffer.head.Load()
			tail := buffer.tail.Load()
			currentCount += (tail - head) & buffer.mask
		}

		// 根据增长率调整批大小
		elapsed := time.Since(lastCheck).Seconds()
		rate := float64(currentCount-lastCount) / elapsed

		currentBatchSize := s.batchSize.Load()
		newBatchSize := currentBatchSize

		if rate > 1000 { // 高负载
			newBatchSize = min(currentBatchSize*2, int32(s.maxBatchSize))
		} else if rate < 100 { // 低负载
			newBatchSize = max(currentBatchSize/2, int32(s.minBatchSize))
		}

		if newBatchSize != currentBatchSize {
			s.batchSize.Store(newBatchSize)
		}

		lastCheck = time.Now()
		lastCount = currentCount
	}
}

// flushAll 刷新所有缓冲区
func (s *MongoLogStore) flushAll(startIdx, endIdx int) {
	batch := batchPool.Get().(*LogBatch)
	defer func() {
		batch.logs = batch.logs[:0]
		batch.size = 0
		batchPool.Put(batch)
	}()

	// 尽可能多地收集日志
	for i := startIdx; i < endIdx; i++ {
		buffer := s.ringBuffers[i]
		for {
			count := buffer.PopBatch(batch, s.maxBatchSize)
			if count == 0 {
				break
			}
			if batch.size > 0 {
				s.batchInsert(batch)
				batch.logs = batch.logs[:0]
				batch.size = 0
			}
		}
	}
}

// Close 关闭日志存储器
func (s *MongoLogStore) Close() {
	// 设置关闭状态
	if !s.state.CompareAndSwap(1, 2) {
		return
	}

	s.logger.Info().Msg("关闭高性能日志处理")

	// 等待所有writer完成
	s.wg.Wait()

	// 重置状态
	s.state.Store(0)
}

// 辅助函数
func min(a, b int32) int32 {
	if a < b {
		return a
	}
	return b
}

func max(a, b int32) int32 {
	if a > b {
		return a
	}
	return b
}
