package sarama

import (
	"context"
	"errors"
	"fmt"
	"os"

	streamdal "github.com/streamdal/streamdal/sdks/go"
)

const (
	StreamdalEnvAddress     = "STREAMDAL_ADDRESS"
	StreamdalEnvAuthToken   = "STREAMDAL_AUTH_TOKEN"
	StreamdalEnvServiceName = "STREAMDAL_SERVICE_NAME"

	StreamdalDefaultComponentName = "kafka"
	StreamdalDefaultOperationName = "unknown"
	StreamdalContextValueKey      = "streamdal-runtime-config"
)

// StreamdalRuntimeConfig is an optional configuration structure that can be
// passed to kafka.FetchMessage() and kafka.WriteMessage() methods to influence
// streamdal shim behavior.
//
// NOTE: This struct is intended to be passed as a value in a context.Context.
// This is done this way to avoid having to change FetchMessage() and WriteMessages()
// signatures.
type StreamdalRuntimeConfig struct {
	// StrictErrors will cause the shim to return a kafka.Error if Streamdal.Process()
	// runs into an unrecoverable error. Default: swallow error and return original value.
	StrictErrors bool

	// Audience is used to specify a custom audience when the shim calls on
	// streamdal.Process(); if nil, a default ComponentName and OperationName
	// will be used. Only non-blank values will be used to override audience defaults.
	Audience *streamdal.Audience
}

func streamdalSetup() (*streamdal.Streamdal, error) {
	address := os.Getenv(StreamdalEnvAddress)
	if address == "" {
		return nil, errors.New(StreamdalEnvAddress + " env var is not set")
	}

	authToken := os.Getenv(StreamdalEnvAuthToken)
	if authToken == "" {
		return nil, errors.New(StreamdalEnvAuthToken + " env var is not set")
	}

	serviceName := os.Getenv(StreamdalEnvServiceName)
	if serviceName == "" {
		return nil, errors.New(StreamdalEnvServiceName + " env var is not set")
	}

	sc, err := streamdal.New(&streamdal.Config{
		ServerURL:   address,
		ServerToken: authToken,
		ServiceName: serviceName,
		ClientType:  streamdal.ClientTypeShim,
	})

	if err != nil {
		return nil, errors.New("unable to create streamdal client: " + err.Error())
	}

	return sc, nil
}

func streamdalProcessForConsumer(sc *streamdal.Streamdal, msg *ConsumerMessage, src *StreamdalRuntimeConfig) (*ConsumerMessage, error) {
	if sc == nil {
		return msg, nil
	}

	updatedData, err := streamdalProcess(sc, streamdal.OperationTypeConsumer, src, msg.Topic, msg.Value)
	if err != nil {
		return nil, fmt.Errorf("streamdalProcess error in streamdalProcessForConsumer: %w", err)
	}

	msg.Value = updatedData

	return msg, nil
}

func streamdalProcessForProducer(sc *streamdal.Streamdal, msg *ProducerMessage) (*ProducerMessage, error) {
	if sc == nil {
		return msg, nil
	}

	data, err := msg.Value.Encode()
	if err != nil {
		return nil, fmt.Errorf("unable to encode msg value in streamdalProcessForProducer: %w", err)
	}

	updatedData, err := streamdalProcess(sc, streamdal.OperationTypeProducer, msg.StreamdalRuntimeConfig, msg.Topic, data)
	if err != nil {
		return nil, fmt.Errorf("streamdalProcess error in streamdalProcessForProducer: %w", err)
	}

	msg.Value = ByteEncoder(updatedData)

	return msg, nil
}

func streamdalProcess(sc *streamdal.Streamdal, ot streamdal.OperationType, src *StreamdalRuntimeConfig, topic string, data []byte) ([]byte, error) {
	// Nothing to do if streamdal client is nil
	if sc == nil {
		return data, nil
	}

	// Generate an audience from the provided parameters
	aud := streamdalGenerateAudience(ot, topic, src)

	// Process msg payload via Streamdal
	resp := sc.Process(context.Background(), &streamdal.ProcessRequest{
		ComponentName: aud.ComponentName,
		OperationType: ot,
		OperationName: aud.OperationName,
		Data:          data,
	})

	switch resp.Status {
	case streamdal.ExecStatusTrue, streamdal.ExecStatusFalse:
		// Process() did not error - replace kafka.Value
		data = resp.Data
	case streamdal.ExecStatusError:
		// Process() errored - return message as-is; if strict errors are NOT
		// set, return error instead of message
		if src != nil && src.StrictErrors {
			fmt.Printf("streamdal.Process() error (strict-errors=true): %v\n", ptrStr(resp.StatusMessage))
			return nil, errors.New("streamdal.Process() error: " + ptrStr(resp.StatusMessage))
		} else {
			fmt.Printf("streamdal.Process() error (strict-errors=false): %v\n", ptrStr(resp.StatusMessage))
		}
	}

	return data, nil
}

// Helper func for generating an "audience" that can be passed to streamdal's .Process() method.
//
// Topic is only used if the provided runtime config is nil or the underlying
// audience does not have an OperationName set.
func streamdalGenerateAudience(ot streamdal.OperationType, topic string, src *StreamdalRuntimeConfig) *streamdal.Audience {
	var (
		componentName = StreamdalDefaultComponentName
		operationName = StreamdalDefaultOperationName
	)

	if topic != "" {
		operationName = topic
	}

	if src != nil && src.Audience != nil {
		if src.Audience.OperationName != "" {
			operationName = src.Audience.OperationName
		}

		if src.Audience.ComponentName != "" {
			componentName = src.Audience.ComponentName
		}
	}

	return &streamdal.Audience{
		OperationType: ot,
		OperationName: operationName,
		ComponentName: componentName,
	}
}

// Helper func to deref string ptrs
func ptrStr(s *string) string {
	if s == nil {
		return ""
	}

	return *s
}
