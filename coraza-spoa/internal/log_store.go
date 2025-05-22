package internal

import (
	"context"
	"sync"
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

// MongoLogStore MongoDB实现的日志存储
type MongoLogStore struct {
	mongo           *mongo.Client
	mongoDB         string
	mongoCollection string
	logChan         chan model.WAFLog
	logger          zerolog.Logger
	processMutex    sync.Mutex // 控制processLogs启动的互斥锁
	processStarted  bool       // 标记是否已启动日志处理
}

const (
	defaultChannelSize = 1000 // 默认通道缓冲大小
)

// 添加单例实例和锁
var (
	mongoLogStoreInstance *MongoLogStore
	mongoLogStoreMutex    sync.Mutex
)

// NewMongoLogStore 创建新的MongoDB日志存储器（单例模式）
func NewMongoLogStore(client *mongo.Client, database, collection string, logger zerolog.Logger) *MongoLogStore {
	mongoLogStoreMutex.Lock()
	defer mongoLogStoreMutex.Unlock()

	// 如果实例已存在，直接返回
	if mongoLogStoreInstance != nil {
		logger.Info().Msg("复用现有的MongoLogStore实例")
		return mongoLogStoreInstance
	}

	store := &MongoLogStore{
		mongo:           client,
		mongoDB:         database,
		mongoCollection: collection,
		logChan:         make(chan model.WAFLog, defaultChannelSize),
		logger:          logger,
		processMutex:    sync.Mutex{},
		processStarted:  false,
	}

	// 保存单例实例
	mongoLogStoreInstance = store
	logger.Info().Msg("创建新的MongoLogStore实例")

	return store
}

// Store 非阻塞地发送日志到存储通道
func (s *MongoLogStore) Store(log model.WAFLog) error {
	select {
	case s.logChan <- log:
		return nil
	default:
		// 通道已满，丢弃日志
		s.logger.Warn().Msg("log channel is full, dropping log entry")
		return nil
	}
}

// Start 启动日志存储处理循环
func (s *MongoLogStore) Start() {
	ctx := context.Background()
	// 加锁确保线程安全
	s.processMutex.Lock()
	defer s.processMutex.Unlock()

	// 检查是否已启动，避免重复启动
	if s.processStarted {
		s.logger.Debug().Msg("日志处理循环已启动，跳过")
		return
	}

	s.logger.Info().Msg("启动日志处理循环")
	s.processStarted = true
	go s.processLogs(ctx)
}

// Close 关闭日志存储器
func (s *MongoLogStore) Close() {
	s.processMutex.Lock()
	defer s.processMutex.Unlock()

	// 如果已经启动，则关闭通道并重置状态
	if s.processStarted {
		s.logger.Info().Msg("关闭日志处理循环")
		close(s.logChan)
		s.processStarted = false
	}
}

// processLogs 处理日志存储循环，使用批处理提高效率
func (s *MongoLogStore) processLogs(ctx context.Context) {
	// 确保在处理结束时更新状态
	defer func() {
		s.processMutex.Lock()
		s.processStarted = false
		s.processMutex.Unlock()
		s.logger.Info().Msg("日志处理循环已退出")
	}()

	collection := s.mongo.Database(s.mongoDB).Collection(s.mongoCollection)

	const (
		batchSize     = 100
		batchInterval = 3 * time.Second
	)

	batch := make([]interface{}, 0, batchSize)
	ticker := time.NewTicker(batchInterval)
	defer ticker.Stop()

	// 刷新批次函数
	flushBatch := func() {
		if len(batch) == 0 {
			return
		}

		// 使用带超时的上下文进行存储操作
		storeCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
		_, err := collection.InsertMany(storeCtx, batch)
		cancel()

		if err != nil {
			s.logger.Error().Err(err).Int("batch_size", len(batch)).
				Msg("failed to save firewall logs to MongoDB")
		}

		// 清空批次
		batch = batch[:0]
	}

	for {
		select {
		case log, ok := <-s.logChan:
			if !ok {
				// 通道已关闭，刷新剩余的日志
				flushBatch()
				return // 通道已关闭
			}

			// 添加到批次
			batch = append(batch, log)

			// 如果批次已满，立即刷新
			if len(batch) >= batchSize {
				flushBatch()
			}

		case <-ticker.C:
			// 定时刷新，确保低流量情况下日志也能及时写入
			flushBatch()

		case <-ctx.Done():
			// 上下文取消，刷新剩余的日志
			flushBatch()
			return
		}
	}
}
