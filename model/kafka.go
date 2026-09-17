package model

// KafkaRecord 是把解析后的 GB32960 消息发送到 Kafka 的传输信封。
// 它不是协议类型,不需要编解码器实现;仅做 JSON 序列化。
type KafkaRecord struct {
	VIN     string `json:"vin"`
	Command string `json:"command"`
	Time    string `json:"time"`
	Payload any    `json:"payload"`
}
