package connections

func InitKafka(broker, topic string) *KafkaConn {
	return &KafkaConn{
		broker: broker,
		topic:  topic,
	}
}
