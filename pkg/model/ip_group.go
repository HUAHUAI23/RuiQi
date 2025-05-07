package model

import "go.mongodb.org/mongo-driver/v2/bson"

type IPGroup struct {
	ID    bson.ObjectID `bson:"_id,omitempty" json:"id,omitempty"` // 日志唯一标识符
	Name  string        `bson:"name" json:"name"`
	Items []string      `bson:"items" json:"items"`
}

func (i *IPGroup) GetCollectionName() string {
	return "ip_group"
}
