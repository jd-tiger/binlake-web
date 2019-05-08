package models

// MQ 创建topic 的MQ 接口
type MQ interface {
	CreateTopic(topic, desc, orderType, url string) error
}

const (
	// MessageTypeBusinessKeyOrder  业务主键 消息的顺序类型
	MessageTypeBusinessKeyOrder = "BUSINESS_KEY_ORDER"

	// MessageTypeTableOrder 表顺序
	MessageTypeTableOrder = "TABLE_ORDER"

	// MessageTypeDbOrder 库顺序
	MessageTypeDbOrder = "DB_ORDER"

	// MessageTypeInstanceOrder 实例顺序
	MessageTypeInstanceOrder = "INSTANCE_ORDER"

	// MessageTypeNoOrder 完全乱序
	MessageTypeNoOrder = "NO_ORDER"
)
