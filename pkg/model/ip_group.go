package model

import "go.mongodb.org/mongo-driver/v2/bson"

// IPGroup 表示IP地址组信息
// @Description IP地址组信息，包含组名和IP地址列表
type IPGroup struct {
	ID    bson.ObjectID `bson:"_id,omitempty" json:"id,omitempty" example:"60d21b4367d0d8992e89e964"` // 组唯一标识符
	Name  string        `bson:"name" json:"name" example:"内部服务器"`                                     // 组名称
	Items []string      `bson:"items" json:"items" example:"['192.168.1.1', '10.0.0.1/24']"`          // IP地址或CIDR列表
}

func (i *IPGroup) GetCollectionName() string {
	return "ip_group"
}
