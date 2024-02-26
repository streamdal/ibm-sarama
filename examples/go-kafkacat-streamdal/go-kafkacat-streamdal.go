/**
 * Copyright 2016 Confluent Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

/**
 * This is a modified & slimmed-down version of the go-kafkacat-streamdal
 * example from the Confluent Kafka Go client library that has been modified to
 * work with ibm-sarama library that is instrumented with Streamdal.
 *
 * See https://github.com/streamdal/confluent-kafka-go/blob/master/examples/go-kafkacat-streamdal/go-kafkacat-streamdal.go
 */

// Example kafkacat clone written in Golang
package main

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/alecthomas/kingpin"
	kafka "github.com/streamdal/ibm-sarama"
	streamdal "github.com/streamdal/streamdal/sdks/go"
)

var (
	keyDelim               = ""
	sigs                   chan os.Signal
	streamdalComponentName string
	streamdalOperationName string
	streamdalStrictErrors  bool
)

func runProducer(config *kafka.Config, brokers []string, topic string, partition int32) {
	// Create a producer that has Streamdal enabled
	p, err := kafka.NewAsyncProducer(brokers, config)
	if err != nil {
		fmt.Fprintf(os.Stderr, ">> Failed to create producer: %s\n", err)
		os.Exit(1)
	}

	fmt.Fprintf(os.Stderr, ">> Created Producer, topic %s [partition %d]\n", topic, partition)

	if streamdalComponentName != kafka.StreamdalDefaultComponentName || streamdalOperationName != kafka.StreamdalDefaultOperationName || streamdalStrictErrors {
		fmt.Fprintf(os.Stderr, ">> Streamdal runtime-config enabled: component=%s, operation=%s, strict-errors=%v\n", streamdalComponentName, streamdalOperationName, streamdalStrictErrors)
	}

	reader := bufio.NewReader(os.Stdin)
	stdinChan := make(chan string)

	// Read input from stdin and send to stdinChan
	go func() {
		for true {
			line, err := reader.ReadString('\n')
			if err != nil {
				break
			}

			line = strings.TrimSuffix(line, "\n")
			if len(line) == 0 {
				continue
			}

			stdinChan <- line
		}
		close(stdinChan)
	}()

	run := true

	for run == true {
		select {
		case sig := <-sigs:
			fmt.Fprintf(os.Stderr, ">> Terminating on signal %v\n", sig)
			run = false

		case line, ok := <-stdinChan:
			if !ok {
				run = false
				break
			}

			msg := &kafka.ProducerMessage{
				Topic:     topic,
				Partition: partition,
			}

			// If keyDelim is set, split the line into key and value
			if keyDelim != "" {
				vec := strings.SplitN(line, keyDelim, 2)
				if len(vec[0]) > 0 {
					msg.Key = kafka.ByteEncoder(vec[0])
				}
				if len(vec) == 2 && len(vec[1]) > 0 {
					msg.Value = kafka.ByteEncoder(vec[1])
				}
			} else {
				msg.Value = kafka.ByteEncoder(line)
			}

			// Inject Streamdal runtime-config into the msg (if provided)
			injectRuntimeConfig(msg, streamdalComponentName, streamdalOperationName, streamdalStrictErrors)

			// Write message to producer
			p.Input() <- msg
		}
	}

	fmt.Fprintf(os.Stderr, ">> Closing\n")
	p.Close()
}

func runReader(config *kafka.Config, brokers []string, groupID string, topics []string) {
	cg, err := kafka.NewConsumerGroup(brokers, groupID, config)
	if err != nil {
		fmt.Fprintf(os.Stderr, ">> Failed to create consumer group: %s\n", err)
		os.Exit(1)
	}

	fmt.Fprintf(os.Stderr, ">> Created Consumer (w/ groupID %s) for topic(s) %+v, using OffsetNewest\n", groupID, topics)

	run := true
	ctx, cancel := context.WithCancel(context.Background())

	consumer := Consumer{}

	// Watch for termination signal
	go func() {
		sig := <-sigs
		fmt.Fprintf(os.Stderr, ">> Terminating on signal %v\n", sig)
		run = false
		cancel()
	}()

	for run == true {
		// Consume handler will auto-commits offsets
		if err := cg.Consume(ctx, topics, &consumer); err != nil {
			if errors.Is(err, context.Canceled) {
				fmt.Fprint(os.Stderr, ">> Detected interrupt, exiting consumer\n")
				run = false
				break
			}

			fmt.Fprintf(os.Stderr, ">> Fetch error: %v\n", err)
			continue
		}
	}

	fmt.Fprintf(os.Stderr, ">> Closing consumer\n")
	cg.Close()
}

func injectRuntimeConfig(msg *kafka.ProducerMessage, cn, on string, se bool) {
	if cn != kafka.StreamdalDefaultComponentName || on != kafka.StreamdalDefaultOperationName || se {
		src := &kafka.StreamdalRuntimeConfig{
			Audience: &streamdal.Audience{},
		}

		if cn != "" {
			src.Audience.ComponentName = cn
		}

		if on != "" {
			src.Audience.OperationName = on
		}

		if se {
			src.StrictErrors = true
		}

		msg.StreamdalRuntimeConfig = src
	}
}

// Consumer represents a Sarama consumer group consumer
type Consumer struct{}

// Setup is run at the beginning of a new session, before ConsumeClaim
func (consumer *Consumer) Setup(kafka.ConsumerGroupSession) error {
	return nil
}

// Cleanup is run at the end of a session, once all ConsumeClaim goroutines have exited
func (consumer *Consumer) Cleanup(kafka.ConsumerGroupSession) error {
	return nil
}

// ConsumeClaim must start a consumer loop of ConsumerGroupClaim's Messages().
// Once the Messages() channel is closed, the Handler must finish its processing
// loop and exit.
func (consumer *Consumer) ConsumeClaim(session kafka.ConsumerGroupSession, claim kafka.ConsumerGroupClaim) error {
	// NOTE:
	// Do not move the code below to a goroutine.
	// The `ConsumeClaim` itself is called within a goroutine, see:
	// https://github.com/IBM/sarama/blob/main/consumer_group.go#L27-L29
	for {
		select {
		case message, ok := <-claim.Messages():
			if !ok {
				log.Printf("message channel was closed")
				return nil
			}
			log.Printf("Message claimed: value = %s, timestamp = %v, topic = %s", string(message.Value), message.Timestamp, message.Topic)
			session.MarkMessage(message, "")
		// Should return when `session.Context()` is done.
		// If not, will raise `ErrRebalanceInProgress` or `read tcp <ip>:<port>: i/o timeout` when kafka rebalance. see:
		// https://github.com/IBM/sarama/issues/1192
		case <-session.Context().Done():
			return nil
		}
	}
}

func main() {
	sigs = make(chan os.Signal)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	/* General options */
	brokers := kingpin.Flag("broker", "Bootstrap broker(s)").Required().String()
	keyDelimArg := kingpin.Flag("key-delim", "Key and value delimiter (empty string=dont print/parse key)").Default("").String()
	kingpin.Flag("streamdal-component-name", "Streamdal component name").Default("kafka").StringVar(&streamdalComponentName)
	kingpin.Flag("streamdal-operation-name", "Streamdal operation name").Default(kafka.StreamdalDefaultOperationName).StringVar(&streamdalOperationName)
	kingpin.Flag("streamdal-strict-errors", "Streamdal enable strict errors").Default("false").BoolVar(&streamdalStrictErrors)

	/* Producer mode options */
	modeP := kingpin.Command("produce", "Produce messages")
	topic := modeP.Flag("topic", "Topic to produce to").Required().String()
	partition := modeP.Flag("partition", "Partition to produce to").Default("-1").Int32()

	/* Consumer mode options */
	modeC := kingpin.Command("consume", "Consume messages").Default()
	group := modeC.Flag("group", "Consumer group").Required().String()
	topics := modeC.Arg("topic", "Topic(s) to subscribe to").Required().Strings()

	mode := kingpin.Parse()
	keyDelim = *keyDelimArg
	splitBrokers := strings.Split(*brokers, ",")

	switch mode {
	case "produce":
		cfg := kafka.NewConfig()
		cfg.EnableStreamdal = true

		runProducer(cfg, splitBrokers, *topic, *partition)

	case "consume":
		cfg := kafka.NewConfig()
		cfg.EnableStreamdal = true
		cfg.Consumer.Offsets.Initial = kafka.OffsetNewest

		runReader(cfg, splitBrokers, *group, *topics)
	}
}
