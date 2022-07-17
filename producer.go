package main

import (
	"bufio"
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"sync"
	"time"

	"gopkg.in/yaml.v3"

	"gopkg.in/confluentinc/confluent-kafka-go.v1/kafka"
)

var TOPIC = "the-meetup-events"

type config struct {
	BootstrapServer  string `yaml:"bootstrap.servers"`
	SecurityProtocol string `yaml:"security.protocol"`
	SASLMechanism    string `yaml:"sasl.mechanisms"`
	SASLUsername     string `yaml:"sasl.username"`
	SASLPassword     string `yaml:"sasl.password"`
}

// RecordValue represents the struct of the value in a Kafka message
type RecordValue struct {
	Count int
	sync.Mutex
}

func getConfig(filename string) (*config, error) {
	buf, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	c := &config{}
	err = yaml.Unmarshal(buf, c)
	if err != nil {
		return nil, fmt.Errorf("in file %q: %v", filename, err)
	}
	fmt.Println(c)

	return c, nil
}

func CreateTopic(p *kafka.Producer, topic string) {
	cl, err := kafka.NewAdminClientFromProducer(p)
	if err != nil {
		panic(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	maxDur, err := time.ParseDuration("60s")
	if err != nil {
		panic(err)
	}
	results, err := cl.CreateTopics(
		ctx,
		[]kafka.TopicSpecification{{
			Topic:             topic,
			NumPartitions:     1,
			ReplicationFactor: 3,
		}},
		kafka.SetAdminOperationTimeout(maxDur),
	)
	if err != nil {
		panic(err)
	}
	for _, result := range results {
		if result.Error.Code() != kafka.ErrNoError && result.Error.Code() != kafka.ErrTopicAlreadyExists {
			fmt.Printf("Failed to create topic: %v\n", result.Error)
			os.Exit(1)
		}
		fmt.Printf("%v\n", result)
	}
	cl.Close()

}

func main() {
	cfg, err := getConfig("config.yaml")
	if err != nil {
		log.Fatal(err)
	}
	//count := RecordValue{
	//	Count: 0,
	//}
	p, err := kafka.NewProducer(&kafka.ConfigMap{
		"bootstrap.servers": cfg.BootstrapServer,
		"sasl.mechanisms":   cfg.SASLMechanism,
		"security.protocol": cfg.SecurityProtocol,
		"sasl.username":     cfg.SASLUsername,
		"sasl.password":     cfg.SASLPassword,
	})
	if err != nil {
		panic(err)
	}

	CreateTopic(p, TOPIC)

	//// Go-routine to handle message delivery reports and
	//// possibly other event types (errors, stats, etc)
	//go func() {
	//	for e := range p.Events() {
	//		switch ev := e.(type) {
	//		case *kafka.Message:
	//			if ev.TopicPartition.Error != nil {
	//				fmt.Printf("Failed to deliver message: %v\n", ev.TopicPartition)
	//			} else {
	//				count.Lock()
	//				count.Count++
	//				fmt.Printf("%d Successfully produced record to topic %s partition [%d] @ offset %v\n",
	//					count.Count, *ev.TopicPartition.Topic, ev.TopicPartition.Partition, ev.TopicPartition.Offset)
	//				count.Unlock()
	//			}
	//		}
	//	}
	//}()

	f, err := os.Open("./data/events.json") // file.json has the json content
	if err != nil {
		log.Fatal(err)
	}

	//client, err := schemaregistry.NewClient(schemaregistry.NewConfigWithAuthentication(
	//	config.,
	//	schemaRegistryAPIKey,
	//	schemaRegistryAPISecret))
	//
	//if err != nil {
	//	fmt.Printf("Failed to create schema registry client: %s\n", err)
	//	os.Exit(1)
	//}
	//
	//ser, err := avro.NewGenericSerializer(client, serde.ValueSerde, avro.NewSerializerConfig())
	//
	//if err != nil {
	//	fmt.Printf("Failed to create serializer: %s\n", err)
	//	os.Exit(1)
	//}
	//deser, err := avro.NewGenericDeserializer(client, serde.ValueSerde, avro.NewDeserializerConfig())
	//
	//if err != nil {
	//	fmt.Printf("Failed to create deserializer: %s\n", err)
	//	os.Exit(1)
	//}
	defer f.Close()
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		p.Produce(&kafka.Message{
			TopicPartition: kafka.TopicPartition{Topic: &TOPIC, Partition: kafka.PartitionAny},
			Value:          scanner.Bytes(),
			Key:            []byte("events"),
		}, nil)
		e := <-p.Events()
		message := e.(*kafka.Message)
		if message.TopicPartition.Error != nil {
			fmt.Printf("failed to deliver message: %v\n",
				message.TopicPartition)
		} else {
			fmt.Printf("delivered to topic %s [%d] at offset %v\n",
				*message.TopicPartition.Topic,
				message.TopicPartition.Partition,
				message.TopicPartition.Offset)
		}
	}
	p.Close()

	//// Wait for all messages to be delivered
	//p.Flush(80 * 1000)

	fmt.Printf("15 messages were produced to topic %s\n!", TOPIC)

	//p.Close()

}
