package flowcontroller

import (
	"container/heap"
	"context"
	"hash/fnv"
	"sync"
	"sync/atomic"
	"time"

	"github.com/HUAHUAI23/simple-waf/pkg/model"
	"github.com/rs/zerolog"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

// RecorderConfig 记录器配置
type RecorderConfig struct {
	Capacity        int
	CleanupInterval time.Duration
	BatchSize       int
	WriteTimeout    time.Duration
	MaxRetries      int
	ShardCount      int // 分片数量，必须是2的幂
	MetricsEnabled  bool
	WriteQueueSize  int
}

// DefaultConfig 默认配置
func DefaultConfig() RecorderConfig {
	return RecorderConfig{
		Capacity:        10000,
		CleanupInterval: time.Minute,
		BatchSize:       100,
		WriteTimeout:    10 * time.Second,
		MaxRetries:      3,
		ShardCount:      16, // 默认16个分片
		MetricsEnabled:  true,
		WriteQueueSize:  10000,
	}
}

// Metrics 监控指标
type Metrics struct {
	TotalBlocked    atomic.Uint64
	TotalExpired    atomic.Uint64
	CurrentBlocked  atomic.Uint64
	CleanupDuration atomic.Value // time.Duration
	WriteQueueSize  atomic.Uint64
	CacheHits       atomic.Uint64
	CacheMisses     atomic.Uint64
}

// MemoryBlockedIP 内存中的简化IP记录，只保存查询必需的字段
type MemoryBlockedIP struct {
	IP           string    // IP地址
	BlockedUntil time.Time // 限制结束时间
}

// IPRecorder IP记录器接口
type IPRecorder interface {
	RecordBlockedIP(ip string, reason string, requestUri string, duration time.Duration) error
	IsIPBlocked(ip string) (bool, *model.BlockedIPRecord)
	GetBlockedIPs() ([]model.BlockedIPRecord, error)
	Close() error
	GetMetrics() *Metrics
}

// IPExpiryItem 用于过期优先队列的项目
type IPExpiryItem struct {
	ip        string
	expiresAt time.Time
	index     int
}

// 对象池，减少GC压力
var ipExpiryItemPool = sync.Pool{
	New: func() interface{} {
		return &IPExpiryItem{}
	},
}

// IPExpiryHeap 过期IP的优先队列实现
type IPExpiryHeap []*IPExpiryItem

func (h IPExpiryHeap) Len() int { return len(h) }

func (h IPExpiryHeap) Less(i, j int) bool {
	return h[i].expiresAt.Before(h[j].expiresAt)
}

func (h IPExpiryHeap) Swap(i, j int) {
	h[i], h[j] = h[j], h[i]
	h[i].index = i
	h[j].index = j
}

func (h *IPExpiryHeap) Push(x interface{}) {
	n := len(*h)
	item := x.(*IPExpiryItem)
	item.index = n
	*h = append(*h, item)
}

func (h *IPExpiryHeap) Pop() interface{} {
	old := *h
	n := len(old)
	item := old[n-1]
	old[n-1] = nil
	item.index = -1
	*h = old[0 : n-1]
	return item
}

func (h *IPExpiryHeap) Peek() *IPExpiryItem {
	if len(*h) == 0 {
		return nil
	}
	return (*h)[0]
}

func (h *IPExpiryHeap) Update(item *IPExpiryItem, expiresAt time.Time) {
	item.expiresAt = expiresAt
	heap.Fix(h, item.index)
}

// shard 分片结构 - 现在使用简化的内存记录
type shard struct {
	mu          sync.RWMutex
	blockedIPs  map[string]MemoryBlockedIP // 使用简化的内存记录
	expiryItems map[string]*IPExpiryItem
	expiryHeap  IPExpiryHeap
	toDelete    []string // 复用的删除缓存
}

// MemoryIPRecorder 基于内存的IP记录器实现（分片版本）
type MemoryIPRecorder struct {
	shards          []*shard
	shardMask       uint32
	capacity        int
	config          RecorderConfig
	logger          zerolog.Logger
	cleanupInterval atomic.Value // time.Duration
	stopCleaner     chan struct{}
	Metrics         *Metrics // 公开以便 MongoIPRecorder 共享
}

// 单例实例
var (
	memoryIPRecorderOnce     sync.Once
	memoryIPRecorderInstance *MemoryIPRecorder
)

// fnv32 计算字符串的FNV-32哈希
func fnv32(s string) uint32 {
	h := fnv.New32a()
	h.Write([]byte(s))
	return h.Sum32()
}

// NewMemoryIPRecorder 创建新的内存IP记录器（单例模式）
func NewMemoryIPRecorder(capacity int, logger zerolog.Logger) *MemoryIPRecorder {
	config := DefaultConfig()
	config.Capacity = capacity
	return NewMemoryIPRecorderWithConfig(config, logger)
}

// NewMemoryIPRecorderWithConfig 使用配置创建内存IP记录器
func NewMemoryIPRecorderWithConfig(config RecorderConfig, logger zerolog.Logger) *MemoryIPRecorder {
	memoryIPRecorderOnce.Do(func() {
		if config.Capacity <= 0 {
			config.Capacity = 10000
		}

		// 确保分片数是2的幂
		shardCount := config.ShardCount
		if shardCount <= 0 || (shardCount&(shardCount-1)) != 0 {
			shardCount = 16
		}

		recorder := &MemoryIPRecorder{
			shards:      make([]*shard, shardCount),
			shardMask:   uint32(shardCount - 1),
			capacity:    config.Capacity,
			config:      config,
			logger:      logger,
			stopCleaner: make(chan struct{}),
			Metrics:     &Metrics{},
		}

		recorder.cleanupInterval.Store(config.CleanupInterval)

		// 初始化每个分片 - 使用简化的内存记录
		capacityPerShard := config.Capacity / shardCount
		for i := 0; i < shardCount; i++ {
			s := &shard{
				blockedIPs:  make(map[string]MemoryBlockedIP, capacityPerShard), // 简化记录
				expiryItems: make(map[string]*IPExpiryItem, capacityPerShard),
				expiryHeap:  make(IPExpiryHeap, 0, capacityPerShard),
				toDelete:    make([]string, 0, 100),
			}
			heap.Init(&s.expiryHeap)
			recorder.shards[i] = s
		}

		// 启动自适应清理
		go recorder.adaptiveCleanupLoop()

		memoryIPRecorderInstance = recorder
		logger.Info().
			Int("capacity", config.Capacity).
			Int("shards", shardCount).
			Msg("创建新的MemoryIPRecorder实例")
	})

	return memoryIPRecorderInstance
}

// getShard 获取IP对应的分片
func (r *MemoryIPRecorder) getShard(ip string) *shard {
	hash := fnv32(ip)
	return r.shards[hash&r.shardMask]
}

// adaptiveCleanupInterval 自适应清理间隔
func (r *MemoryIPRecorder) adaptiveCleanupInterval() time.Duration {
	// 计算总使用率
	var totalUsed int
	for _, s := range r.shards {
		s.mu.RLock()
		totalUsed += len(s.blockedIPs)
		s.mu.RUnlock()
	}

	usage := float64(totalUsed) / float64(r.capacity)

	var interval time.Duration
	switch {
	case usage > 0.9:
		interval = 10 * time.Second
	case usage > 0.7:
		interval = 30 * time.Second
	case usage > 0.5:
		interval = 45 * time.Second
	default:
		interval = r.config.CleanupInterval
	}

	r.cleanupInterval.Store(interval)
	return interval
}

// adaptiveCleanupLoop 自适应清理循环
func (r *MemoryIPRecorder) adaptiveCleanupLoop() {
	ticker := time.NewTicker(r.config.CleanupInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			start := time.Now()
			r.cleanupExpired()
			duration := time.Since(start)

			if r.config.MetricsEnabled {
				r.Metrics.CleanupDuration.Store(duration)
			}

			// 调整下次清理间隔
			interval := r.adaptiveCleanupInterval()
			ticker.Reset(interval)

		case <-r.stopCleaner:
			return
		}
	}
}

// cleanupExpired 清理过期的IP记录
func (r *MemoryIPRecorder) cleanupExpired() {
	now := time.Now()
	var totalRemoved int

	// 并行清理各个分片
	var wg sync.WaitGroup
	removedChan := make(chan int, len(r.shards))

	for _, s := range r.shards {
		wg.Add(1)
		go func(s *shard) {
			defer wg.Done()
			removed := r.cleanupShardExpired(s, now)
			removedChan <- removed
		}(s)
	}

	wg.Wait()
	close(removedChan)

	for removed := range removedChan {
		totalRemoved += removed
	}

	if totalRemoved > 0 {
		r.Metrics.TotalExpired.Add(uint64(totalRemoved))
		r.logger.Debug().
			Int("removed", totalRemoved).
			Msg("已清理过期IP记录")
	}
}

// cleanupShardExpired 清理单个分片的过期记录
func (r *MemoryIPRecorder) cleanupShardExpired(s *shard, now time.Time) int {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.toDelete = s.toDelete[:0]

	// 从堆顶开始检查
	for s.expiryHeap.Len() > 0 {
		item := s.expiryHeap.Peek()
		if item.expiresAt.After(now) {
			break
		}

		heap.Pop(&s.expiryHeap)
		s.toDelete = append(s.toDelete, item.ip)

		// 返回对象到池
		item.ip = ""
		item.expiresAt = time.Time{}
		item.index = -1
		ipExpiryItemPool.Put(item)
	}

	// 批量删除
	removed := len(s.toDelete)
	for _, ip := range s.toDelete {
		delete(s.blockedIPs, ip)
		delete(s.expiryItems, ip)
	}

	return removed
}

// ensureShardCapacity 确保分片容量不超限
func (r *MemoryIPRecorder) ensureShardCapacity(s *shard) {
	capacityPerShard := r.capacity / len(r.shards)
	if len(s.blockedIPs) < capacityPerShard {
		return
	}

	now := time.Now()
	s.toDelete = s.toDelete[:0]

	// 先清理过期的
	for s.expiryHeap.Len() > 0 && len(s.blockedIPs) >= capacityPerShard {
		item := s.expiryHeap.Peek()
		if item.expiresAt.After(now) {
			break
		}

		heap.Pop(&s.expiryHeap)
		s.toDelete = append(s.toDelete, item.ip)
		ipExpiryItemPool.Put(item)
	}

	// 批量删除过期记录
	for _, ip := range s.toDelete {
		delete(s.blockedIPs, ip)
		delete(s.expiryItems, ip)
	}

	// 如果还是满的，删除最早过期的
	for len(s.blockedIPs) >= capacityPerShard && s.expiryHeap.Len() > 0 {
		item := heap.Pop(&s.expiryHeap).(*IPExpiryItem)
		delete(s.blockedIPs, item.ip)
		delete(s.expiryItems, item.ip)

		r.logger.Debug().
			Str("ip", item.ip).
			Time("expires_at", item.expiresAt).
			Msg("容量已满，移除最早过期IP记录")

		ipExpiryItemPool.Put(item)
	}
}

// RecordBlockedIP 记录被限制的IP - 内存中只保存必要字段
func (r *MemoryIPRecorder) RecordBlockedIP(ip string, reason string, requestUri string, duration time.Duration) error {
	s := r.getShard(ip)

	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()
	expiresAt := now.Add(duration)

	// 内存中只保存必要字段
	memoryRecord := MemoryBlockedIP{
		IP:           ip,
		BlockedUntil: expiresAt,
	}

	// 检查IP是否已存在
	if item, exists := s.expiryItems[ip]; exists {
		s.blockedIPs[ip] = memoryRecord
		s.expiryHeap.Update(item, expiresAt)

		r.logger.Info().
			Str("ip", ip).
			Str("reason", reason).
			Time("until", expiresAt).
			Msg("更新IP限制记录")
		return nil
	}

	// 确保容量
	r.ensureShardCapacity(s)

	// 添加新记录
	s.blockedIPs[ip] = memoryRecord

	item := ipExpiryItemPool.Get().(*IPExpiryItem)
	item.ip = ip
	item.expiresAt = expiresAt

	s.expiryItems[ip] = item
	heap.Push(&s.expiryHeap, item)

	r.Metrics.TotalBlocked.Add(1)
	r.Metrics.CurrentBlocked.Add(1)

	r.logger.Info().
		Str("ip", ip).
		Str("reason", reason).
		Time("until", expiresAt).
		Msg("IP已被限制")

	return nil
}

// IsIPBlocked 检查IP是否被限制 - 返回简化的结果
func (r *MemoryIPRecorder) IsIPBlocked(ip string) (bool, *model.BlockedIPRecord) {
	// 使用defer recover防止任何可能的panic
	defer func() {
		if r := recover(); r != nil {
			// 发生panic时直接放行
		}
	}()

	s := r.getShard(ip)

	// 无锁访问，接受并发风险
	memoryRecord, exists := s.blockedIPs[ip]
	if !exists {
		r.Metrics.CacheMisses.Add(1)
		return false, nil
	}

	// 安全检查时间
	now := time.Now()
	if memoryRecord.BlockedUntil.IsZero() || now.After(memoryRecord.BlockedUntil) {
		// 过期了，直接返回false，不删除（避免加锁）
		r.Metrics.CacheMisses.Add(1)
		return false, nil
	}

	r.Metrics.CacheHits.Add(1)

	// 转换为完整的BlockedIPRecord用于返回（只包含内存中有的字段）
	fullRecord := &model.BlockedIPRecord{
		IP:           memoryRecord.IP,
		BlockedUntil: memoryRecord.BlockedUntil,
		// Reason、RequestUri等字段在内存中不保存，保持零值
	}

	return true, fullRecord
}

// GetBlockedIPs 获取所有被限制的IP - 返回简化的记录
func (r *MemoryIPRecorder) GetBlockedIPs() ([]model.BlockedIPRecord, error) {
	now := time.Now()
	records := make([]model.BlockedIPRecord, 0, r.capacity/10)

	// 收集所有分片的记录
	for _, s := range r.shards {
		s.mu.RLock()
		for _, memoryRecord := range s.blockedIPs {
			if now.Before(memoryRecord.BlockedUntil) {
				// 转换为完整的BlockedIPRecord（只包含内存中有的字段）
				fullRecord := model.BlockedIPRecord{
					IP:           memoryRecord.IP,
					BlockedUntil: memoryRecord.BlockedUntil,
					// Reason、RequestUri等字段在内存中不保存，保持零值
				}
				records = append(records, fullRecord)
			}
		}
		s.mu.RUnlock()
	}

	return records, nil
}

// GetMetrics 获取监控指标
func (r *MemoryIPRecorder) GetMetrics() *Metrics {
	// 更新当前阻塞数
	var current uint64
	for _, s := range r.shards {
		s.mu.RLock()
		current += uint64(len(s.blockedIPs))
		s.mu.RUnlock()
	}
	r.Metrics.CurrentBlocked.Store(current)

	return r.Metrics
}

// Close 关闭记录器并释放资源
func (r *MemoryIPRecorder) Close() error {
	close(r.stopCleaner)
	return nil
}

// RingBuffer 环形缓冲区
type RingBuffer struct {
	buffer []model.BlockedIPRecord
	head   uint64
	tail   uint64
	size   uint64
	mask   uint64
	mu     sync.Mutex
}

// NewRingBuffer 创建环形缓冲区
func NewRingBuffer(size int) *RingBuffer {
	// 确保size是2的幂
	actualSize := 1
	for actualSize < size {
		actualSize <<= 1
	}

	return &RingBuffer{
		buffer: make([]model.BlockedIPRecord, actualSize),
		size:   uint64(actualSize),
		mask:   uint64(actualSize - 1),
	}
}

// Push 添加元素
func (rb *RingBuffer) Push(record model.BlockedIPRecord) bool {
	rb.mu.Lock()
	defer rb.mu.Unlock()

	next := (rb.tail + 1) & rb.mask
	if next == rb.head {
		return false // 缓冲区满
	}

	rb.buffer[rb.tail] = record
	rb.tail = next
	return true
}

// PopBatch 批量弹出元素
func (rb *RingBuffer) PopBatch(maxBatch int) []model.BlockedIPRecord {
	rb.mu.Lock()
	defer rb.mu.Unlock()

	if rb.head == rb.tail {
		return nil
	}

	batch := make([]model.BlockedIPRecord, 0, maxBatch)
	for i := 0; i < maxBatch && rb.head != rb.tail; i++ {
		batch = append(batch, rb.buffer[rb.head])
		rb.head = (rb.head + 1) & rb.mask
	}

	return batch
}

// Len 获取当前元素数量
func (rb *RingBuffer) Len() int {
	rb.mu.Lock()
	defer rb.mu.Unlock()

	if rb.tail >= rb.head {
		return int(rb.tail - rb.head)
	}
	return int(rb.size - rb.head + rb.tail)
}

// CircuitBreaker 熔断器
type CircuitBreaker struct {
	failures      atomic.Uint32
	successCount  atomic.Uint32
	lastFailure   atomic.Value  // time.Time
	state         atomic.Uint32 // 0: closed, 1: open, 2: half-open
	threshold     uint32
	timeout       time.Duration
	recoveryCount uint32
}

// NewCircuitBreaker 创建熔断器
func NewCircuitBreaker(threshold uint32, timeout time.Duration, recoveryCount uint32) *CircuitBreaker {
	return &CircuitBreaker{
		threshold:     threshold,
		timeout:       timeout,
		recoveryCount: recoveryCount,
	}
}

// IsOpen 检查熔断器是否打开
func (cb *CircuitBreaker) IsOpen() bool {
	state := cb.state.Load()
	if state == 0 { // closed
		return false
	}

	if state == 1 { // open
		lastFailure, ok := cb.lastFailure.Load().(time.Time)
		if ok && time.Since(lastFailure) > cb.timeout {
			// 尝试进入半开状态
			cb.state.CompareAndSwap(1, 2)
			cb.successCount.Store(0)
			return false
		}
		return true
	}

	// half-open
	return false
}

// RecordSuccess 记录成功
func (cb *CircuitBreaker) RecordSuccess() {
	state := cb.state.Load()
	if state == 2 { // half-open
		count := cb.successCount.Add(1)
		if count >= cb.recoveryCount {
			cb.state.Store(0) // closed
			cb.failures.Store(0)
		}
	} else if state == 0 { // closed
		cb.failures.Store(0)
	}
}

// RecordFailure 记录失败
func (cb *CircuitBreaker) RecordFailure() {
	failures := cb.failures.Add(1)
	cb.lastFailure.Store(time.Now())

	if failures >= cb.threshold {
		cb.state.Store(1) // open
	}
}

// MongoIPRecorder MongoDB实现的IP记录器
type MongoIPRecorder struct {
	client         *mongo.Client
	database       string
	collection     string
	memory         *MemoryIPRecorder
	logger         zerolog.Logger
	config         RecorderConfig
	metrics        *Metrics
	circuitBreaker *CircuitBreaker

	// 使用环形缓冲区替代channel
	writeBuffer     *RingBuffer
	stopWriter      chan struct{}
	avgWriteLatency atomic.Value // time.Duration
}

// 单例实例
var (
	mongoIPRecorderOnce     sync.Once
	mongoIPRecorderInstance *MongoIPRecorder
)

// NewMongoIPRecorder 创建新的MongoDB IP记录器
func NewMongoIPRecorder(client *mongo.Client, database string, capacity int, logger zerolog.Logger) *MongoIPRecorder {
	config := DefaultConfig()
	config.Capacity = capacity
	return NewMongoIPRecorderWithConfig(client, database, config, logger)
}

// NewMongoIPRecorderWithConfig 使用配置创建MongoDB IP记录器
func NewMongoIPRecorderWithConfig(client *mongo.Client, database string, config RecorderConfig, logger zerolog.Logger) *MongoIPRecorder {
	mongoIPRecorderOnce.Do(func() {
		var blockedIPs model.BlockedIPRecord

		// 先创建内存记录器
		memoryRecorder := NewMemoryIPRecorderWithConfig(config, logger)

		recorder := &MongoIPRecorder{
			client:         client,
			database:       database,
			collection:     blockedIPs.GetCollectionName(),
			memory:         memoryRecorder,
			logger:         logger,
			config:         config,
			metrics:        memoryRecorder.Metrics, // 共享metrics
			circuitBreaker: NewCircuitBreaker(5, 30*time.Second, 3),
			writeBuffer:    NewRingBuffer(config.WriteQueueSize),
			stopWriter:     make(chan struct{}),
		}

		// 启动批量写入
		go recorder.adaptiveBatchWriteLoop()

		mongoIPRecorderInstance = recorder
		logger.Info().Msg("创建新的MongoIPRecorder实例")
	})

	return mongoIPRecorderInstance
}

// adaptiveBatchSize 动态调整批量大小
func (r *MongoIPRecorder) adaptiveBatchSize() int {
	latency, ok := r.avgWriteLatency.Load().(time.Duration)
	if !ok {
		return r.config.BatchSize
	}

	switch {
	case latency > 200*time.Millisecond:
		return r.config.BatchSize / 2
	case latency > 100*time.Millisecond:
		return r.config.BatchSize * 3 / 4
	case latency < 50*time.Millisecond:
		return r.config.BatchSize * 2
	default:
		return r.config.BatchSize
	}
}

// adaptiveBatchWriteLoop 自适应批量写入循环
func (r *MongoIPRecorder) adaptiveBatchWriteLoop() {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			batchSize := r.adaptiveBatchSize()
			batch := r.writeBuffer.PopBatch(batchSize)
			if len(batch) > 0 {
				r.metrics.WriteQueueSize.Store(uint64(r.writeBuffer.Len()))
				go r.flushBatchWithRetry(batch)
			}

		case <-r.stopWriter:
			// 关闭前写入剩余数据
			batch := r.writeBuffer.PopBatch(r.config.BatchSize * 10)
			if len(batch) > 0 {
				r.flushBatchWithRetry(batch)
			}
			return
		}
	}
}

// flushBatchWithRetry 带重试的批量写入
func (r *MongoIPRecorder) flushBatchWithRetry(batch []model.BlockedIPRecord) {
	for retry := 0; retry < r.config.MaxRetries; retry++ {
		start := time.Now()
		err := r.flushBatch(batch)
		latency := time.Since(start)

		// 更新平均延迟
		r.avgWriteLatency.Store(latency)

		if err == nil {
			r.circuitBreaker.RecordSuccess()
			return
		}

		r.circuitBreaker.RecordFailure()

		if retry < r.config.MaxRetries-1 {
			// 指数退避
			time.Sleep(time.Duration(1<<retry) * time.Second)
		}
	}

	r.logger.Error().
		Int("batch_size", len(batch)).
		Msg("批量写入MongoDB失败，已达最大重试次数")
}

// flushBatch 批量写入到MongoDB
func (r *MongoIPRecorder) flushBatch(batch []model.BlockedIPRecord) error {
	if len(batch) == 0 {
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), r.config.WriteTimeout)
	defer cancel()

	collection := r.client.Database(r.database).Collection(r.collection)

	// 批量插入操作 - 直接插入新记录
	// 注意：这将保留完整的历史记录，建议考虑：
	// 1. 创建TTL索引自动清理过期记录: db.collection.createIndex({"blockedUntil": 1}, {expireAfterSeconds: 0})
	// 2. 为查询性能创建索引: db.collection.createIndex({"ip": 1, "blockedUntil": -1})
	// 3. 定期清理过期记录，避免数据过度累积
	models := make([]mongo.WriteModel, len(batch))
	for i, record := range batch {
		models[i] = mongo.NewInsertOneModel().SetDocument(record)
	}

	opts := options.BulkWrite().SetOrdered(false)
	_, err := collection.BulkWrite(ctx, models, opts)
	if err != nil {
		r.logger.Error().
			Err(err).
			Int("batch_size", len(batch)).
			Msg("批量插入IP限制记录到MongoDB失败")
		return err
	}

	r.logger.Debug().
		Int("batch_size", len(batch)).
		Msg("成功批量插入IP限制记录到MongoDB")

	return nil
}

// RecordBlockedIP 记录被限制的IP
func (r *MongoIPRecorder) RecordBlockedIP(ip string, reason string, requestUri string, duration time.Duration) error {
	// 先记录到内存
	err := r.memory.RecordBlockedIP(ip, reason, requestUri, duration)
	if err != nil {
		return err
	}

	// 如果熔断器打开，直接返回
	if r.circuitBreaker.IsOpen() {
		r.logger.Warn().
			Str("ip", ip).
			Msg("MongoDB熔断器已打开，跳过持久化")
		return nil
	}

	// 异步写入到环形缓冲区
	now := time.Now()
	record := model.BlockedIPRecord{
		IP:           ip,
		Reason:       reason,
		RequestUri:   requestUri,
		BlockedAt:    now,
		BlockedUntil: now.Add(duration),
	}

	if !r.writeBuffer.Push(record) {
		r.logger.Warn().
			Str("ip", ip).
			Msg("MongoDB写入缓冲区已满，丢弃记录")
	}

	return nil
}

// IsIPBlocked 检查IP是否被限制
func (r *MongoIPRecorder) IsIPBlocked(ip string) (bool, *model.BlockedIPRecord) {
	return r.memory.IsIPBlocked(ip)
}

// GetBlockedIPs 获取所有被限制的IP
func (r *MongoIPRecorder) GetBlockedIPs() ([]model.BlockedIPRecord, error) {
	return r.memory.GetBlockedIPs()
}

// GetMetrics 获取监控指标 - 确保内存指标是最新的
func (r *MongoIPRecorder) GetMetrics() *Metrics {
	// 首先获取内存记录器的最新指标
	r.memory.GetMetrics()

	// 然后更新MongoDB相关的指标
	r.metrics.WriteQueueSize.Store(uint64(r.writeBuffer.Len()))

	return r.metrics
}

// Close 关闭记录器并释放资源
func (r *MongoIPRecorder) Close() error {
	close(r.stopWriter)
	return r.memory.Close()
}
