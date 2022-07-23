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
			ReplicationFactor: 1,
		}},
		kafka.SetAdminOperationTimeout(maxDur),
	)
	if err != nil {
		panic(err)
	}
	for _, result := range results {
		if result.Error.Code() != kafka.ErrNoError &&
			result.Error.Code() != kafka.ErrTopicAlreadyExists {
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

	p, err := kafka.NewProducer(&kafka.ConfigMap{
		"bootstrap.servers": cfg.BootstrapServer,
		// "sasl.mechanisms":   cfg.SASLMechanism,
		// "security.protocol": cfg.SecurityProtocol,
		// "sasl.username":     cfg.SASLUsername,
		// "sasl.password":     cfg.SASLPassword,
	})
	if err != nil {
		panic(err)
	}

	CreateTopic(p, TOPIC)

	f, err := os.Open("./data/events.json")
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		p.Produce(&kafka.Message{
			TopicPartition: kafka.TopicPartition{
				Topic:     &TOPIC,
				Partition: kafka.PartitionAny,
			},
			Value: scanner.Bytes(),
			Key:   []byte("events"),
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

	fmt.Printf("Events are produced to %s topic \n", TOPIC)

}
