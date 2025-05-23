package benchmarks

import (
	"fmt"
	"sync/atomic"
	"testing"
	"time"

	flowcontroller "github.com/HUAHUAI23/simple-waf/coraza-spoa/internal/flow-controller"
	"github.com/rs/zerolog"
)

// BenchmarkMetricsOverhead 测试 Metrics 对系统的性能开销
func BenchmarkMetricsOverhead(b *testing.B) {
	logger := zerolog.Nop()

	b.Run("WithMetrics", func(b *testing.B) {
		config := flowcontroller.DefaultConfig()
		config.MetricsEnabled = true
		config.Capacity = 1000

		recorder := flowcontroller.NewMemoryIPRecorderWithConfig(config, logger)
		defer recorder.Close()

		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			ip := "192.168.1.1"
			for pb.Next() {
				recorder.IsIPBlocked(ip)
			}
		})
	})

	b.Run("WithoutMetrics", func(b *testing.B) {
		config := flowcontroller.DefaultConfig()
		config.MetricsEnabled = false
		config.Capacity = 1000

		recorder := flowcontroller.NewMemoryIPRecorderWithConfig(config, logger)
		defer recorder.Close()

		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			ip := "192.168.1.1"
			for pb.Next() {
				recorder.IsIPBlocked(ip)
			}
		})
	})
}

// BenchmarkIPRecorderRealWorld 测试真实场景下的性能影响
func BenchmarkIPRecorderRealWorld(b *testing.B) {
	logger := zerolog.Nop()
	config := flowcontroller.DefaultConfig()
	config.Capacity = 10000

	recorder := flowcontroller.NewMemoryIPRecorderWithConfig(config, logger)
	defer recorder.Close()

	// 预先添加一些IP记录
	for i := 0; i < 1000; i++ {
		ip := fmt.Sprintf("192.168.1.%d", i%255)
		recorder.RecordBlockedIP(ip, "test", "/", time.Hour)
	}

	b.Run("IsIPBlocked_Miss", func(b *testing.B) {
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				// 测试缓存未命中的场景
				recorder.IsIPBlocked("10.0.0.1")
			}
		})
	})

	b.Run("IsIPBlocked_Hit", func(b *testing.B) {
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				// 测试缓存命中的场景
				recorder.IsIPBlocked("192.168.1.1")
			}
		})
	})

	b.Run("RecordBlockedIP", func(b *testing.B) {
		b.RunParallel(func(pb *testing.PB) {
			i := 0
			for pb.Next() {
				ip := fmt.Sprintf("10.1.1.%d", i%255)
				recorder.RecordBlockedIP(ip, "test", "/", time.Hour)
				i++
			}
		})
	})
}

// BenchmarkAtomicOperations 测试原子操作的性能
func BenchmarkAtomicOperations(b *testing.B) {
	var counter atomic.Uint64

	b.Run("AtomicAdd", func(b *testing.B) {
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				counter.Add(1)
			}
		})
	})

	b.Run("AtomicStore", func(b *testing.B) {
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				counter.Store(uint64(time.Now().UnixNano()))
			}
		})
	})

	b.Run("AtomicLoad", func(b *testing.B) {
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				_ = counter.Load()
			}
		})
	})
}

// BenchmarkMetricsUpdate 模拟真实的指标更新场景
func BenchmarkMetricsUpdate(b *testing.B) {
	metrics := &flowcontroller.Metrics{}

	b.Run("SingleMetricUpdate", func(b *testing.B) {
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				metrics.CacheHits.Add(1)
			}
		})
	})

	b.Run("MultipleMetricsUpdate", func(b *testing.B) {
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				// 模拟 IsIPBlocked 中的指标更新
				metrics.CacheHits.Add(1)
				metrics.CurrentBlocked.Store(100)
			}
		})
	})

	b.Run("AllMetricsUpdate", func(b *testing.B) {
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				// 模拟完整的指标更新场景
				metrics.TotalBlocked.Add(1)
				metrics.TotalExpired.Add(1)
				metrics.CurrentBlocked.Store(100)
				metrics.WriteQueueSize.Store(50)
				metrics.CacheHits.Add(1)
				metrics.CacheMisses.Add(1)
				metrics.CleanupDuration.Store(time.Millisecond)
			}
		})
	})
}

// BenchmarkMemoryFootprint 测试内存占用
func BenchmarkMemoryFootprint(b *testing.B) {
	b.Run("WithMetrics", func(b *testing.B) {
		var metrics []*flowcontroller.Metrics
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			m := &flowcontroller.Metrics{}
			metrics = append(metrics, m)
		}

		b.StopTimer()
		// 防止编译器优化
		_ = metrics
	})

	b.Run("WithoutMetrics", func(b *testing.B) {
		var counters []int
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			counters = append(counters, i)
		}

		b.StopTimer()
		// 防止编译器优化
		_ = counters
	})
}

// BenchmarkCacheSimulation 模拟缓存命中率计算的性能开销
func BenchmarkCacheSimulation(b *testing.B) {
	metrics := &flowcontroller.Metrics{}

	b.Run("CacheHitRateCalculation", func(b *testing.B) {
		// 预先设置一些基础数据
		metrics.CacheHits.Store(1000)
		metrics.CacheMisses.Store(100)

		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				hits := metrics.CacheHits.Load()
				misses := metrics.CacheMisses.Load()
				total := hits + misses
				if total > 0 {
					hitRate := float64(hits) / float64(total) * 100
					_ = hitRate // 防止编译器优化
				}
			}
		})
	})
}
