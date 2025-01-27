package connections

func InitKafka(broker, topic string) *kafkaConn {
	return &kafkaConn{
		broker: broker,
		topic:  topic,
	}
}
