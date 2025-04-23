// model/ip_info.go
package model

// IPInfo 表示IP地址的地理位置信息
// @Description IP地址地理位置详细信息，包含城市、区域、国家和ASN等数据
type IPInfo struct {
	City struct {
		NameZH string `json:"nameZh" bson:"nameZh" example:"杭州"`       // 城市中文名称
		NameEN string `json:"nameEn" bson:"nameEn" example:"Hangzhou"` // 城市英文名称
	} `json:"city" bson:"city"`

	Subdivision struct {
		NameZH  string `json:"nameZh" bson:"nameZh" example:"浙江"`       // 省/州中文名称
		NameEN  string `json:"nameEn" bson:"nameEn" example:"Zhejiang"` // 省/州英文名称
		IsoCode string `json:"isoCode" bson:"isoCode" example:"ZJ"`     // 省/州代码
	} `json:"subdivision" bson:"subdivision"`

	Country struct {
		NameZH  string `json:"nameZh" bson:"nameZh" example:"中国"`    // 国家中文名称
		NameEN  string `json:"nameEn" bson:"nameEn" example:"China"` // 国家英文名称
		IsoCode string `json:"isoCode" bson:"isoCode" example:"CN"`  // 国家ISO代码
	} `json:"country" bson:"country"`

	Continent struct {
		NameZH string `json:"nameZh" bson:"nameZh" example:"亚洲"`   // 大洲中文名称
		NameEN string `json:"nameEn" bson:"nameEn" example:"Asia"` // 大洲英文名称
	} `json:"continent" bson:"continent"`

	Location struct {
		Longitude float64 `json:"longitude" bson:"longitude" example:"120.1663"`    // 经度
		Latitude  float64 `json:"latitude" bson:"latitude" example:"30.2943"`       // 纬度
		TimeZone  string  `json:"timeZone" bson:"timeZone" example:"Asia/Shanghai"` // 时区
	} `json:"location" bson:"location"`

	ASN struct {
		Number       uint   `json:"number" bson:"number" example:"4134"`                      // ASN号码
		Organization string `json:"organization" bson:"organization" example:"China Telecom"` // 组织名称
	} `json:"asn" bson:"asn"`
}
