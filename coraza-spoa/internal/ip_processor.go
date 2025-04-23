package internal

import (
	"context"
	"net"
	"sync"

	"github.com/HUAHUAI23/simple-waf/pkg/model"
	"github.com/oschwald/geoip2-golang"
	"github.com/rs/zerolog"
)

// IPProcessor 是IP地理位置信息处理器接口
type IPProcessor interface {
	// GetIPInfo 根据IP地址字符串获取地理位置信息
	GetIPInfo(ipStr string) *model.IPInfo

	// Close 关闭处理器并释放资源
	Close()
}

// GeoIP2Processor 是基于MaxMind GeoIP2数据库的IP处理器实现
type GeoIP2Processor struct {
	cityDB *geoip2.Reader  // 城市数据库
	asnDB  *geoip2.Reader  // ASN数据库
	logger zerolog.Logger  // 日志记录器
	mutex  sync.RWMutex    // 读写锁
	ctx    context.Context // 上下文
	closed bool            // 是否已关闭
}

// GeoIP2Options 包含GeoIP2处理器的配置选项
type GeoIP2Options struct {
	CityDBPath string // City数据库路径
	ASNDBPath  string // ASN数据库路径
}

// NewGeoIP2Processor 创建一个新的GeoIP2处理器实例
func NewGeoIP2Processor(ctx context.Context, options GeoIP2Options, logger zerolog.Logger) (IPProcessor, error) {
	processor := &GeoIP2Processor{
		logger: logger,
		ctx:    ctx,
	}

	// 尝试打开City数据库
	if options.CityDBPath != "" {
		cityDB, err := geoip2.Open(options.CityDBPath)
		if err != nil {
			logger.Error().Err(err).Str("path", options.CityDBPath).Msg("打开城市数据库失败")
		} else {
			processor.cityDB = cityDB
		}
	} else {
		logger.Warn().Msg("未提供城市数据库路径，地理位置功能将不可用")
	}

	// 尝试打开ASN数据库
	if options.ASNDBPath != "" {
		asnDB, err := geoip2.Open(options.ASNDBPath)
		if err != nil {
			logger.Warn().Err(err).Str("path", options.ASNDBPath).Msg("打开ASN数据库失败，ASN信息将不可用")
		} else {
			processor.asnDB = asnDB
		}
	} else {
		logger.Warn().Msg("未提供ASN数据库路径，ASN信息将不可用")
	}

	// 如果两个数据库都无法打开，返回空实现
	if processor.cityDB == nil && processor.asnDB == nil {
		return NewNullIPProcessor(), nil
	}

	// 启动后台监听上下文取消信号的goroutine
	go processor.watchContext()

	return processor, nil
}

// watchContext 监听上下文取消信号，当上下文取消时自动释放资源
func (p *GeoIP2Processor) watchContext() {
	<-p.ctx.Done()
	p.closeResources()
}

// closeResources 关闭数据库并标记处理器为已关闭
func (p *GeoIP2Processor) closeResources() {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	if p.closed {
		return
	}

	// 关闭数据库
	if p.cityDB != nil {
		p.cityDB.Close()
		p.cityDB = nil
	}

	if p.asnDB != nil {
		p.asnDB.Close()
		p.asnDB = nil
	}

	p.closed = true
	p.logger.Debug().Msg("IP处理器资源已释放")
}

// Close 手动关闭IP处理器
func (p *GeoIP2Processor) Close() {
	p.closeResources()
}

// GetIPInfo 根据IP地址字符串获取地理位置信息
func (p *GeoIP2Processor) GetIPInfo(ipStr string) *model.IPInfo {
	// 快速检查IP格式有效性
	ip := net.ParseIP(ipStr)
	if ip == nil {
		p.logger.Warn().Str("ip", ipStr).Msg("无效的IP地址")
		return nil
	}

	// 检查上下文是否已取消
	select {
	case <-p.ctx.Done():
		return nil
	default:
		// 继续处理
	}

	// 如果处理器已关闭，直接返回nil
	if p.closed || (p.cityDB == nil && p.asnDB == nil) {
		return nil
	}

	ipInfo := &model.IPInfo{}
	hasInfo := false

	// 查询城市信息
	if p.cityDB != nil {
		cityRecord, err := p.cityDB.City(ip)
		if err != nil {
			p.logger.Warn().Str("ip", ipStr).Err(err).Msg("查询城市信息失败")
		} else {
			hasInfo = true

			// 填充城市信息
			if len(cityRecord.City.Names) > 0 {
				if name, exists := cityRecord.City.Names["zh-CN"]; exists {
					ipInfo.City.NameZH = name
				}
				if name, exists := cityRecord.City.Names["en"]; exists {
					ipInfo.City.NameEN = name
				}
			}

			// 填充省/州信息
			if len(cityRecord.Subdivisions) > 0 {
				subdivision := cityRecord.Subdivisions[0]
				if name, exists := subdivision.Names["zh-CN"]; exists {
					ipInfo.Subdivision.NameZH = name
				}
				if name, exists := subdivision.Names["en"]; exists {
					ipInfo.Subdivision.NameEN = name
				}
				ipInfo.Subdivision.IsoCode = subdivision.IsoCode
			}

			// 填充国家信息
			if name, exists := cityRecord.Country.Names["zh-CN"]; exists {
				ipInfo.Country.NameZH = name
			}
			if name, exists := cityRecord.Country.Names["en"]; exists {
				ipInfo.Country.NameEN = name
			}
			ipInfo.Country.IsoCode = cityRecord.Country.IsoCode

			// 填充大洲信息
			if len(cityRecord.Continent.Names) > 0 {
				if name, exists := cityRecord.Continent.Names["zh-CN"]; exists {
					ipInfo.Continent.NameZH = name
				}
				if name, exists := cityRecord.Continent.Names["en"]; exists {
					ipInfo.Continent.NameEN = name
				}
			}

			// 填充位置信息
			ipInfo.Location.Longitude = cityRecord.Location.Longitude
			ipInfo.Location.Latitude = cityRecord.Location.Latitude
			ipInfo.Location.TimeZone = cityRecord.Location.TimeZone
		}
	}

	// 查询ASN信息
	if p.asnDB != nil {
		asnRecord, err := p.asnDB.ASN(ip)
		if err != nil {
			p.logger.Warn().Str("ip", ipStr).Err(err).Msg("查询ASN信息失败")
		} else {
			hasInfo = true
			ipInfo.ASN.Number = asnRecord.AutonomousSystemNumber
			ipInfo.ASN.Organization = asnRecord.AutonomousSystemOrganization
		}
	}

	// 如果没有获取到任何信息，则返回nil
	if !hasInfo {
		return nil
	}

	return ipInfo
}

// NewIPProcessor 创建新的IP处理器实例的工厂方法
func NewIPProcessor(ctx context.Context, cityDBPath, asnDBPath string, logger zerolog.Logger) (IPProcessor, error) {
	if cityDBPath == "" && asnDBPath == "" {
		// 如果没有提供任何数据库路径，则返回空实现
		return NewNullIPProcessor(), nil
	}

	// 创建处理器
	return NewGeoIP2Processor(ctx, GeoIP2Options{
		CityDBPath: cityDBPath,
		ASNDBPath:  asnDBPath,
	}, logger)
}

// NullIPProcessor 是IP处理器接口的空实现
type NullIPProcessor struct{}

// NewNullIPProcessor 创建一个空实现的IP处理器
func NewNullIPProcessor() IPProcessor {
	return &NullIPProcessor{}
}

// GetIPInfo 返回nil，空实现始终返回nil
func (p *NullIPProcessor) GetIPInfo(ipStr string) *model.IPInfo {
	return nil
}

// Close 空实现不需要做任何事
func (p *NullIPProcessor) Close() {
	// 无需任何操作
}
