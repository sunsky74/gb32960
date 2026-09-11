package model

// KafkaRecord is a transport envelope for sending parsed GB32960 messages to Kafka.
// It is NOT a protocol type — no codec implementation needed. JSON serialization only.
type KafkaRecord struct {
	VIN     string `json:"vin"`
	Command string `json:"command"`
	Time    string `json:"time"`
	Payload any    `json:"payload"`
}
