package internal

import (
	"context"
	"time"

	"github.com/HUAHUAI23/simple-waf/pkg/model"
	"github.com/rs/zerolog"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

// LogStore 定义日志存储接口
type LogStore interface {
	Store(log model.WAFLog) error
	Start(ctx context.Context)
	Close()
}

// MongoLogStore MongoDB实现的日志存储
type MongoLogStore struct {
	mongo           *mongo.Client
	mongoDB         string
	mongoCollection string
	logChan         chan model.WAFLog
	logger          zerolog.Logger
}

const (
	defaultChannelSize = 1000 // 默认通道缓冲大小
)

// NewMongoLogStore 创建新的MongoDB日志存储器
func NewMongoLogStore(client *mongo.Client, database, collection string, logger zerolog.Logger) *MongoLogStore {
	return &MongoLogStore{
		mongo:           client,
		mongoDB:         database,
		mongoCollection: collection,
		logChan:         make(chan model.WAFLog, defaultChannelSize),
		logger:          logger,
	}
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
func (s *MongoLogStore) Start(ctx context.Context) {
	go s.processLogs(ctx)
}

// Close 关闭日志存储器
func (s *MongoLogStore) Close() {
	close(s.logChan)
}

// processLogs 处理日志存储循环
// processLogs 处理日志存储循环，使用批处理提高效率
func (s *MongoLogStore) processLogs(ctx context.Context) {
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
