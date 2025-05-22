package flowcontroller

import (
	"container/heap"
	"context"
	"sync"
	"time"

	"github.com/HUAHUAI23/simple-waf/pkg/model"
	"github.com/rs/zerolog"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

// IPRecorder IP记录器接口
type IPRecorder interface {
	// RecordBlockedIP 记录被限制的IP
	RecordBlockedIP(ip string, reason string, duration time.Duration) error

	// IsIPBlocked 检查IP是否被限制
	IsIPBlocked(ip string) (bool, *model.BlockedIPRecord)

	// GetBlockedIPs 获取所有被限制的IP
	GetBlockedIPs() ([]model.BlockedIPRecord, error)

	// Close 关闭记录器并释放资源
	Close() error
}

// IPExpiryItem 用于过期优先队列的项目
type IPExpiryItem struct {
	ip        string    // IP地址
	expiresAt time.Time // 过期时间
	index     int       // 在堆中的索引
}

// IPExpiryHeap 过期IP的优先队列实现
type IPExpiryHeap []*IPExpiryItem

func (h IPExpiryHeap) Len() int { return len(h) }

func (h IPExpiryHeap) Less(i, j int) bool {
	// 过期时间早的优先
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
	old[n-1] = nil  // 避免内存泄漏
	item.index = -1 // 标记为已移除
	*h = old[0 : n-1]
	return item
}

func (h *IPExpiryHeap) Peek() *IPExpiryItem {
	if len(*h) == 0 {
		return nil
	}
	return (*h)[0]
}

// Update 更新堆中项目的过期时间
func (h *IPExpiryHeap) Update(item *IPExpiryItem, expiresAt time.Time) {
	item.expiresAt = expiresAt
	heap.Fix(h, item.index)
}

// MemoryIPRecorder 基于内存的IP记录器实现
type MemoryIPRecorder struct {
	blockedIPs      map[string]model.BlockedIPRecord // IP记录映射
	expiryItems     map[string]*IPExpiryItem         // IP到过期项的映射
	expiryHeap      IPExpiryHeap                     // 过期优先队列
	capacity        int                              // 最大容量
	mu              sync.RWMutex                     // 读写锁
	logger          zerolog.Logger                   // 日志记录器
	cleanupInterval time.Duration                    // 清理间隔
	stopCleaner     chan struct{}                    // 停止清理的信号

	// 优化：预分配批量删除缓存
	toDelete []string // 复用的删除缓存，减少内存分配
}

// 添加内存记录器单例实例和锁
var (
	memoryIPRecorderInstance *MemoryIPRecorder
	memoryIPRecorderMutex    sync.Mutex
)

// NewMemoryIPRecorder 创建新的内存IP记录器（单例模式）
// @Summary 创建内存IP记录器
// @Description 创建一个基于内存的IP记录器，用于记录被限制的IP，采用单例模式
// @Param capacity int - 记录器的最大容量，如果小于等于0则使用默认值10000
// @Param logger zerolog.Logger - 日志记录器
// @Return *MemoryIPRecorder - 创建的内存IP记录器
func NewMemoryIPRecorder(capacity int, logger zerolog.Logger) *MemoryIPRecorder {
	memoryIPRecorderMutex.Lock()
	defer memoryIPRecorderMutex.Unlock()

	// 如果实例已存在，直接返回
	if memoryIPRecorderInstance != nil {
		logger.Debug().Msg("复用现有的MemoryIPRecorder实例")
		return memoryIPRecorderInstance
	}

	if capacity <= 0 {
		capacity = 10000 // 默认容量
	}

	recorder := &MemoryIPRecorder{
		blockedIPs:      make(map[string]model.BlockedIPRecord, capacity),
		expiryItems:     make(map[string]*IPExpiryItem, capacity),
		expiryHeap:      make(IPExpiryHeap, 0, capacity),
		capacity:        capacity,
		logger:          logger,
		cleanupInterval: time.Minute, // 默认每分钟清理一次
		stopCleaner:     make(chan struct{}),
		toDelete:        make([]string, 0, 100), // 预分配批量删除缓存
	}

	// 初始化优先队列
	heap.Init(&recorder.expiryHeap)

	// 启动清理过期记录的goroutine
	go recorder.cleanupLoop()

	// 保存单例实例
	memoryIPRecorderInstance = recorder
	logger.Info().Msg("创建新的MemoryIPRecorder实例")

	return recorder
}

// cleanupLoop 循环清理过期记录
func (r *MemoryIPRecorder) cleanupLoop() {
	ticker := time.NewTicker(r.cleanupInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			r.cleanupExpired()
		case <-r.stopCleaner:
			return
		}
	}
}

// cleanupExpired 清理过期的IP记录 - 优化版本
func (r *MemoryIPRecorder) cleanupExpired() {
	now := time.Now()

	r.mu.Lock()
	defer r.mu.Unlock()

	// 重置删除缓存，复用slice避免重复分配
	r.toDelete = r.toDelete[:0]

	// 从堆顶开始检查，收集所有已过期的记录
	for r.expiryHeap.Len() > 0 {
		item := r.expiryHeap.Peek()
		if item.expiresAt.After(now) {
			// 堆顶没过期，后面的也都没过期
			break
		}

		// 收集要删除的IP
		heap.Pop(&r.expiryHeap)
		r.toDelete = append(r.toDelete, item.ip)
	}

	// 批量删除，减少map操作次数
	removed := len(r.toDelete)
	for _, ip := range r.toDelete {
		delete(r.blockedIPs, ip)
		delete(r.expiryItems, ip)
	}

	if removed > 0 {
		r.logger.Debug().
			Int("removed", removed).
			Msg("已清理过期IP记录")
	}
}

// ensureCapacity 确保容量不超限 - 优化版本
func (r *MemoryIPRecorder) ensureCapacity() {
	// 如果没有达到容量限制，直接返回
	if len(r.blockedIPs) < r.capacity {
		return
	}

	// 先尝试清理已过期的记录
	now := time.Now()
	r.toDelete = r.toDelete[:0] // 复用slice

	// 快速检查堆顶的一些过期记录
	for r.expiryHeap.Len() > 0 && len(r.blockedIPs) >= r.capacity {
		item := r.expiryHeap.Peek()
		if item.expiresAt.After(now) {
			break // 没过期的停止
		}

		heap.Pop(&r.expiryHeap)
		r.toDelete = append(r.toDelete, item.ip)
	}

	// 批量删除过期记录
	for _, ip := range r.toDelete {
		delete(r.blockedIPs, ip)
		delete(r.expiryItems, ip)
	}

	// 如果清理过期记录后还是容量不够，移除最早过期的记录
	for len(r.blockedIPs) >= r.capacity && r.expiryHeap.Len() > 0 {
		item := heap.Pop(&r.expiryHeap).(*IPExpiryItem)
		delete(r.blockedIPs, item.ip)
		delete(r.expiryItems, item.ip)

		r.logger.Debug().
			Str("ip", item.ip).
			Time("expires_at", item.expiresAt).
			Msg("容量已满，移除最早过期IP记录")
	}
}

// removeExpiredRecord 立即删除过期记录的内部方法
// 可以在 IsIPBlocked 调用时，添加删除该条过期记录
func (r *MemoryIPRecorder) removeExpiredRecord(ip string) {
	if item, exists := r.expiryItems[ip]; exists {
		// 从堆中移除
		if item.index >= 0 && item.index < r.expiryHeap.Len() {
			heap.Remove(&r.expiryHeap, item.index)
		}
		delete(r.expiryItems, ip)
	}
	delete(r.blockedIPs, ip)
}

// RecordBlockedIP 记录被限制的IP - 优化版本
func (r *MemoryIPRecorder) RecordBlockedIP(ip string, reason string, duration time.Duration) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	now := time.Now()
	expiresAt := now.Add(duration)

	record := model.BlockedIPRecord{
		IP:           ip,
		Reason:       reason,
		BlockedAt:    now,
		BlockedUntil: expiresAt,
	}

	// 检查IP是否已存在
	if item, exists := r.expiryItems[ip]; exists {
		// 更新现有记录
		r.blockedIPs[ip] = record
		r.expiryHeap.Update(item, expiresAt)

		r.logger.Info().
			Str("ip", ip).
			Str("reason", reason).
			Time("until", expiresAt).
			Msg("更新IP限制记录")
		return nil
	}

	// 确保容量不超限
	r.ensureCapacity()

	// 添加新记录
	r.blockedIPs[ip] = record
	item := &IPExpiryItem{
		ip:        ip,
		expiresAt: expiresAt,
	}
	r.expiryItems[ip] = item
	heap.Push(&r.expiryHeap, item)

	r.logger.Info().
		Str("ip", ip).
		Str("reason", reason).
		Time("until", expiresAt).
		Msg("IP已被限制")

	return nil
}

// IsIPBlocked 检查IP是否被限制 - 无锁版本，接受并发风险
func (r *MemoryIPRecorder) IsIPBlocked(ip string) (bool, *model.BlockedIPRecord) {
	// 使用defer recover防止panic
	defer func() {
		if r := recover(); r != nil {
			// 发生panic时直接放行
		}
	}()

	record, exists := r.blockedIPs[ip]
	if !exists {
		return false, nil
	}

	// 安全检查时间，避免损坏的时间值
	now := time.Now()
	if record.BlockedUntil.IsZero() || now.After(record.BlockedUntil) {
		return false, nil
	}

	return true, &record
}

// GetBlockedIPs 获取所有被限制的IP - 优化版本
func (r *MemoryIPRecorder) GetBlockedIPs() ([]model.BlockedIPRecord, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	now := time.Now()
	// 预分配合适的容量，避免slice扩容
	records := make([]model.BlockedIPRecord, 0, len(r.blockedIPs))

	for _, record := range r.blockedIPs {
		if now.Before(record.BlockedUntil) {
			records = append(records, record)
		}
	}

	return records, nil
}

// Close 关闭记录器并释放资源
func (r *MemoryIPRecorder) Close() error {
	close(r.stopCleaner)
	return nil
}

// MongoIPRecorder MongoDB实现的IP记录器 - 优化版本
type MongoIPRecorder struct {
	client     *mongo.Client
	database   string
	collection string
	memory     *MemoryIPRecorder // 内存记录器
	logger     zerolog.Logger

	// 优化：批量写入通道
	writeQueue chan model.BlockedIPRecord
	stopWriter chan struct{}
}

// 添加单例实例和锁
var (
	mongoIPRecorderInstance *MongoIPRecorder
	mongoIPRecorderMutex    sync.Mutex
)

// NewMongoIPRecorder 创建新的MongoDB IP记录器 - 优化版本 (单例模式)
// @Summary 创建MongoDB IP记录器
// @Description 创建一个基于MongoDB的IP记录器，用于持久化存储被限制的IP，采用单例模式
// @Param client *mongo.Client - MongoDB客户端
// @Param database string - 数据库名称
// @Param capacity int - 记录器的最大容量，如果小于等于0则使用默认值10000
// @Param logger zerolog.Logger - 日志记录器
// @Return *MongoIPRecorder - 创建的MongoDB IP记录器
func NewMongoIPRecorder(client *mongo.Client, database string, capacity int, logger zerolog.Logger) *MongoIPRecorder {
	mongoIPRecorderMutex.Lock()
	defer mongoIPRecorderMutex.Unlock()

	// 如果实例已存在，直接返回
	if mongoIPRecorderInstance != nil {
		logger.Info().Msg("复用现有的MongoIPRecorder实例")
		return mongoIPRecorderInstance
	}

	var blockedIPs model.BlockedIPRecord

	recorder := &MongoIPRecorder{
		client:     client,
		database:   database,
		collection: blockedIPs.GetCollectionName(),
		memory:     NewMemoryIPRecorder(capacity, logger),
		logger:     logger,
		writeQueue: make(chan model.BlockedIPRecord, 1000), // 异步写入队列
		stopWriter: make(chan struct{}),
	}

	// 启动批量写入goroutine
	go recorder.batchWriteLoop()

	// 保存单例实例
	mongoIPRecorderInstance = recorder
	logger.Info().Msg("创建新的MongoIPRecorder实例")

	return recorder
}

// batchWriteLoop 批量写入到MongoDB
func (r *MongoIPRecorder) batchWriteLoop() {
	ticker := time.NewTicker(5 * time.Second) // 每5秒批量写入一次
	defer ticker.Stop()

	var batch []model.BlockedIPRecord

	for {
		select {
		case record := <-r.writeQueue:
			batch = append(batch, record)
			// 如果批次满了，立即写入
			if len(batch) >= 100 {
				r.flushBatch(batch)
				batch = batch[:0] // 重置slice
			}

		case <-ticker.C:
			// 定时写入
			if len(batch) > 0 {
				r.flushBatch(batch)
				batch = batch[:0]
			}

		case <-r.stopWriter:
			// 关闭前写入剩余数据
			if len(batch) > 0 {
				r.flushBatch(batch)
			}
			return
		}
	}
}

// flushBatch 批量写入到MongoDB
func (r *MongoIPRecorder) flushBatch(batch []model.BlockedIPRecord) {
	if len(batch) == 0 {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	collection := r.client.Database(r.database).Collection(r.collection)

	// 批量upsert操作
	models := make([]mongo.WriteModel, len(batch))
	for i, record := range batch {
		filter := map[string]interface{}{"ip": record.IP}
		update := map[string]interface{}{"$set": record}
		models[i] = mongo.NewUpdateOneModel().
			SetFilter(filter).
			SetUpdate(update).
			SetUpsert(true)
	}

	opts := options.BulkWrite().SetOrdered(false) // 无序写入提高性能
	_, err := collection.BulkWrite(ctx, models, opts)
	if err != nil {
		r.logger.Error().
			Err(err).
			Int("batch_size", len(batch)).
			Msg("批量保存IP限制记录到MongoDB失败")
	}
}

// RecordBlockedIP 记录被限制的IP - 优化版本
func (r *MongoIPRecorder) RecordBlockedIP(ip string, reason string, duration time.Duration) error {
	// 先记录到内存
	err := r.memory.RecordBlockedIP(ip, reason, duration)
	if err != nil {
		return err
	}

	// 异步写入到队列
	now := time.Now()
	record := model.BlockedIPRecord{
		IP:           ip,
		Reason:       reason,
		BlockedAt:    now,
		BlockedUntil: now.Add(duration),
	}

	// 非阻塞写入，如果队列满了就丢弃（优先保证内存记录的性能）
	select {
	case r.writeQueue <- record:
	default:
		r.logger.Warn().
			Str("ip", ip).
			Msg("MongoDB写入队列已满，丢弃记录")
	}

	return nil
}

// IsIPBlocked 检查IP是否被限制
func (r *MongoIPRecorder) IsIPBlocked(ip string) (bool, *model.BlockedIPRecord) {
	// 直接从内存查询
	return r.memory.IsIPBlocked(ip)
}

// GetBlockedIPs 获取所有被限制的IP
func (r *MongoIPRecorder) GetBlockedIPs() ([]model.BlockedIPRecord, error) {
	// 从内存返回所有记录
	return r.memory.GetBlockedIPs()
}

// Close 关闭记录器并释放资源
func (r *MongoIPRecorder) Close() error {
	close(r.stopWriter) // 先停止写入
	close(r.writeQueue) // 再关闭队列
	return r.memory.Close()
}
