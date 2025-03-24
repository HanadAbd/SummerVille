package connections

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/IBM/sarama"
)

type KafkaConn struct {
	Name       string
	Host       string
	Port       string
	Producer   sarama.SyncProducer
	Consumer   sarama.Consumer
	Admin      sarama.ClusterAdmin
	Metrics    ConnectionMetrics
	Connected  bool
	Credential *KafkaCredential
}

type KafkaCredential struct {
	Name   string
	Broker string
	Topic  string
	// Additional fields for different Kafka configurations
	SecurityProtocol string
	SaslMechanism    string
	SaslUsername     string
	SaslPassword     string
}

func (k *KafkaConn) GetSource() string {
	return fmt.Sprintf("%s:%s", k.Host, k.Port)
}

func (k *KafkaConn) InitKafka(kf *KafkaCredential, cm ConnectionMetrics) (*KafkaConn, error) {
	config := sarama.NewConfig()
	config.Producer.Return.Successes = true

	// Configure security if needed
	if kf.SecurityProtocol != "" {
		// Set security options based on the protocol
	}

	broker := kf.Broker
	if broker == "" {
		broker = fmt.Sprintf("%s:%s", k.Host, k.Port)
	}

	// Create producer
	producer, err := sarama.NewSyncProducer([]string{broker}, config)
	if err != nil {
		return nil, fmt.Errorf("failed to create Kafka producer: %w", err)
	}

	// Create consumer
	consumer, err := sarama.NewConsumer([]string{broker}, config)
	if err != nil {
		producer.Close()
		return nil, fmt.Errorf("failed to create Kafka consumer: %w", err)
	}

	// Create admin client
	admin, err := sarama.NewClusterAdmin([]string{broker}, config)
	if err != nil {
		producer.Close()
		consumer.Close()
		return nil, fmt.Errorf("failed to create Kafka admin client: %w", err)
	}

	k.Producer = producer
	k.Consumer = consumer
	k.Admin = admin
	k.Metrics = cm
	k.Connected = true
	k.Credential = kf

	log.Printf("Successfully connected to Kafka broker at %s", broker)
	return k, nil
}

func (k *KafkaConn) RetryConnection(maxAttempts int, delay time.Duration) error {
	if k.Connected {
		return nil
	}

	var lastErr error
	for attempt := 1; attempt <= maxAttempts; attempt++ {
		config := sarama.NewConfig()
		config.Producer.Return.Successes = true

		broker := k.Credential.Broker
		if broker == "" {
			broker = fmt.Sprintf("%s:%s", k.Host, k.Port)
		}

		producer, err := sarama.NewSyncProducer([]string{broker}, config)
		if err != nil {
			lastErr = err
			log.Printf("Kafka connection attempt %d/%d failed: %v", attempt, maxAttempts, err)
			time.Sleep(delay)
			continue
		}

		consumer, err := sarama.NewConsumer([]string{broker}, config)
		if err != nil {
			producer.Close()
			lastErr = err
			log.Printf("Kafka consumer connection attempt %d/%d failed: %v", attempt, maxAttempts, err)
			time.Sleep(delay)
			continue
		}

		admin, err := sarama.NewClusterAdmin([]string{broker}, config)
		if err != nil {
			producer.Close()
			consumer.Close()
			lastErr = err
			log.Printf("Kafka admin connection attempt %d/%d failed: %v", attempt, maxAttempts, err)
			time.Sleep(delay)
			continue
		}

		k.Producer = producer
		k.Consumer = consumer
		k.Admin = admin
		k.Connected = true

		log.Printf("Successfully reconnected to Kafka broker at %s", broker)
		return nil
	}

	return fmt.Errorf("failed to reconnect to Kafka after %d attempts: %v", maxAttempts, lastErr)
}

func (k *KafkaConn) MonitorConnection(maxAttempts int, delay time.Duration) error {
	if k.Producer == nil {
		return k.RetryConnection(maxAttempts, delay)
	}

	// Check if we can communicate with the broker
	topics, err := k.Admin.ListTopics()
	if err != nil {
		k.Connected = false
		k.Metrics.Status = "disconnected"
		k.Metrics.LastError = err
		k.Metrics.LastErrorTime = time.Now()

		log.Printf("Kafka connection issue detected: %v", err)
		return k.RetryConnection(maxAttempts, delay)
	}

	k.Connected = true
	k.Metrics.Status = "connected"
	log.Printf("Kafka connection healthy, found %d topics", len(topics))

	return nil
}

func (k *KafkaConn) AddData(table TableDefinition, data []interface{}) error {
	if k == nil {
		return fmt.Errorf("kafka connection is nil, use NewKafkaConn first")
	}

	if !k.Connected || k.Producer == nil {
		if k.Credential == nil {
			k.Credential = &KafkaCredential{
				Name:   table.Name,
				Broker: getEnvWithDefault("KAFKA_BROKER", "localhost:9092"),
				Topic:  table.Name,
			}
		}

		if _, err := k.InitKafka(k.Credential, k.Metrics); err != nil {
			return fmt.Errorf("failed to initialize Kafka connection: %w", err)
		}
	}

	if err := k.InitialiseData(table); err != nil {
		return fmt.Errorf("failed to initialize Kafka topic: %w", err)
	}

	topic := table.Name
	startTime := time.Now()

	for _, item := range data {
		jsonData, err := json.Marshal(item)
		if err != nil {
			return fmt.Errorf("failed to marshal data to JSON: %w", err)
		}

		msg := &sarama.ProducerMessage{
			Topic: topic,
			Value: sarama.ByteEncoder(jsonData),
		}

		_, _, err = k.Producer.SendMessage(msg)
		if err != nil {
			k.Metrics.LastError = err
			k.Metrics.LastErrorTime = time.Now()
			return fmt.Errorf("failed to send message to Kafka: %w", err)
		}
	}

	k.Metrics.LastQueryTime = time.Since(startTime)
	k.Metrics.QueryCount++

	return nil
}

func NewKafkaConn(name string) *KafkaConn {
	broker := getEnvWithDefault("KAFKA_BROKER", "localhost:9092")
	hostPort := strings.Split(broker, ":")
	host := hostPort[0]
	port := "9092"
	if len(hostPort) > 1 {
		port = hostPort[1]
	}

	return &KafkaConn{
		Name: name,
		Host: host,
		Port: port,
		Metrics: ConnectionMetrics{
			Status: "initializing",
		},
		Connected: false,
		Credential: &KafkaCredential{
			Name:   name,
			Broker: broker,
			Topic:  name,
		},
	}
}

func getEnvWithDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func (k *KafkaConn) InitialiseData(table TableDefinition) error {
	if !k.Connected || k.Admin == nil {
		return fmt.Errorf("kafka connection not established")
	}

	topic := table.Name

	// Check if topic exists
	topics, err := k.Admin.ListTopics()
	if err != nil {
		return fmt.Errorf("failed to list topics: %w", err)
	}

	if _, exists := topics[topic]; !exists {
		// Create topic if it doesn't exist
		topicDetail := &sarama.TopicDetail{
			NumPartitions:     1,
			ReplicationFactor: 1,
		}

		err = k.Admin.CreateTopic(topic, topicDetail, false)
		if err != nil {
			return fmt.Errorf("failed to create topic %s: %w", topic, err)
		}

		log.Printf("Created Kafka topic: %s", topic)
	}

	return nil
}

func (k *KafkaConn) PurgeData(table TableDefinition) error {
	if !k.Connected || k.Admin == nil {
		return fmt.Errorf("kafka connection not established")
	}

	topic := table.Name

	err := k.Admin.DeleteTopic(topic)
	if err != nil {
		return fmt.Errorf("failed to delete topic %s: %w", topic, err)
	}

	time.Sleep(5 * time.Second)

	topicDetail := &sarama.TopicDetail{
		NumPartitions:     1,
		ReplicationFactor: 1,
	}

	err = k.Admin.CreateTopic(topic, topicDetail, false)
	if err != nil {
		return fmt.Errorf("failed to recreate topic %s: %w", topic, err)
	}

	log.Printf("Purged Kafka topic: %s", topic)
	return nil
}

func (k *KafkaConn) PurgeAllData() error {
	if !k.Connected || k.Admin == nil {
		return fmt.Errorf("kafka connection not established")
	}

	topics, err := k.Admin.ListTopics()
	if err != nil {
		return fmt.Errorf("failed to list topics: %w", err)
	}

	for topic := range topics {
		if len(topic) > 0 && topic[0:2] == "__" {
			continue
		}

		err := k.Admin.DeleteTopic(topic)
		if err != nil {
			log.Printf("Failed to delete topic %s: %v", topic, err)
			continue
		}
	}

	return nil
}

func (k *KafkaConn) CloseConnection() error {
	var result error

	if k.Producer != nil {
		if err := k.Producer.Close(); err != nil {
			result = fmt.Errorf("failed to close producer: %w", err)
		}
	}

	if k.Consumer != nil {
		if err := k.Consumer.Close(); err != nil {
			if result != nil {
				result = fmt.Errorf("%v; failed to close consumer: %w", result, err)
			} else {
				result = fmt.Errorf("failed to close consumer: %w", err)
			}
		}
	}

	if k.Admin != nil {
		if err := k.Admin.Close(); err != nil {
			if result != nil {
				result = fmt.Errorf("%v; failed to close admin: %w", result, err)
			} else {
				result = fmt.Errorf("failed to close admin: %w", err)
			}
		}
	}

	k.Connected = false
	return result
}
