package tiga

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/segmentio/kafka-go"
)

type KafkaDao struct {
	client     *kafka.Client
	config     *Configuration
	partitions map[string][]int
}

func NewKafkaDao(config *Configuration) *KafkaDao {
	env := config.GetEnv()
	addr := config.GetConfigByEnv(env, "kafka.addr").(string)
	client := &kafka.Client{
		Addr:    kafka.TCP(addr),
		Timeout: 10 * time.Second,
	}
	return &KafkaDao{
		client:     client,
		config:     config,
		partitions: make(map[string][]int),
	}

}
func (k *KafkaDao) RecvMessage(topic string, groupID string, messages chan<- kafka.Message) error {
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers: []string{k.client.Addr.String()},
		Topic:   topic,
		GroupID: groupID,
	})
	for {
		m, err := reader.ReadMessage(context.Background())
		if err != nil {
			log.Errorf("消费队列：%s,出现错误:%s", topic, err.Error())
			return err
		}
		log.Infof("message at topic/partition/offset %v/%v/%v: %s = %s\n", m.Topic, m.Partition, m.Offset, string(m.Key), string(m.Value))

		messages <- m

	}
}
func (k *KafkaDao) readPartitions(topic string) error {
	conn, err := kafka.Dial("tcp", k.client.Addr.String())
	if err != nil {
		return fmt.Errorf("kafka连接失败:%w", err)
	}
	partitions, err := conn.ReadPartitions(topic) // your-topic替换为你的kafka主题
	if err != nil {
		return fmt.Errorf("Failed to get partitions:%w", err)
	}

	// 输出分区信息
	for _, p := range partitions {
		log.Infof("Topic: %s, Partition: %d\n", p.Topic, p.ID)
		k.partitions[topic] = append(k.partitions[topic], p.ID)
	}
	return nil
}
func (k *KafkaDao) SendMessage(topic string, record string) (*kafka.ProduceResponse, error) {
	now := time.Now()
	partitions := k.partitions[topic]
	if partitions == nil {
		err := k.readPartitions(topic)
		if err != nil {
			return nil, err
		}
	}
	r := rand.New(rand.NewSource(time.Now().UnixNano()))

	partitions = k.partitions[topic]
	partition := k.partitions[topic][r.Intn(len(partitions))]
	rsp, err := k.client.Produce(context.Background(), &kafka.ProduceRequest{
		Topic:        string(topic),
		Partition:    partition,
		RequiredAcks: -1,
		Records: kafka.NewRecordReader(
			kafka.Record{Time: now, Value: kafka.NewBytes([]byte(record)), Key: nil},
		),
	})
	if err != nil {
		return nil, err
	}
	if rsp.Error != nil {
		return nil, rsp.Error
	}
	return rsp, err
}
