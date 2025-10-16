// Package mock provides mock implementations of the mq package interfaces for testing.
package mock

import (
	"context"
	"sync"

	amqp "github.com/rabbitmq/amqp091-go"

	"procodus.dev/demo-app/pkg/mq"
)

// MockClient is a mock implementation of ClientInterface for testing.
// It tracks method calls and allows configuring return values and behavior.
type MockClient struct {
	mu sync.Mutex

	// PushFunc is called when Push is invoked. If nil, returns PushError.
	PushFunc func(ctx context.Context, data []byte) error
	// PushError is returned by Push if PushFunc is nil.
	PushError error
	// PushCalls tracks all calls to Push with their arguments.
	PushCalls []PushCall

	// UnsafePushFunc is called when UnsafePush is invoked. If nil, returns UnsafePushError.
	UnsafePushFunc func(ctx context.Context, data []byte) error
	// UnsafePushError is returned by UnsafePush if UnsafePushFunc is nil.
	UnsafePushError error
	// UnsafePushCalls tracks all calls to UnsafePush with their arguments.
	UnsafePushCalls []UnsafePushCall

	// ConsumeFunc is called when Consume is invoked. If nil, returns ConsumeChannel and ConsumeError.
	ConsumeFunc func() (<-chan amqp.Delivery, error)
	// ConsumeChannel is returned by Consume if ConsumeFunc is nil.
	ConsumeChannel <-chan amqp.Delivery
	// ConsumeError is returned by Consume if ConsumeFunc is nil.
	ConsumeError error
	// ConsumeCalls tracks the number of times Consume was called.
	ConsumeCalls int

	// CloseFunc is called when Close is invoked. If nil, returns CloseError.
	CloseFunc func() error
	// CloseError is returned by Close if CloseFunc is nil.
	CloseError error
	// CloseCalls tracks the number of times Close was called.
	CloseCalls int
}

// PushCall records the arguments to a Push call.
type PushCall struct {
	Ctx  context.Context
	Data []byte
}

// UnsafePushCall records the arguments to an UnsafePush call.
type UnsafePushCall struct {
	Ctx  context.Context
	Data []byte
}

// NewMockClient creates a new MockClient with default behavior (no errors).
func NewMockClient() *MockClient {
	return &MockClient{
		PushCalls:       make([]PushCall, 0),
		UnsafePushCalls: make([]UnsafePushCall, 0),
		ConsumeChannel:  make(chan amqp.Delivery),
	}
}

// Push implements ClientInterface.
func (m *MockClient) Push(ctx context.Context, data []byte) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.PushCalls = append(m.PushCalls, PushCall{
		Ctx:  ctx,
		Data: data,
	})

	if m.PushFunc != nil {
		return m.PushFunc(ctx, data)
	}
	return m.PushError
}

// UnsafePush implements ClientInterface.
func (m *MockClient) UnsafePush(ctx context.Context, data []byte) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.UnsafePushCalls = append(m.UnsafePushCalls, UnsafePushCall{
		Ctx:  ctx,
		Data: data,
	})

	if m.UnsafePushFunc != nil {
		return m.UnsafePushFunc(ctx, data)
	}
	return m.UnsafePushError
}

// Consume implements ClientInterface.
func (m *MockClient) Consume() (<-chan amqp.Delivery, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.ConsumeCalls++

	if m.ConsumeFunc != nil {
		return m.ConsumeFunc()
	}
	return m.ConsumeChannel, m.ConsumeError
}

// Close implements ClientInterface.
func (m *MockClient) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.CloseCalls++

	if m.CloseFunc != nil {
		return m.CloseFunc()
	}
	return m.CloseError
}

// Reset clears all tracked calls and resets the mock to its initial state.
func (m *MockClient) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.PushCalls = make([]PushCall, 0)
	m.UnsafePushCalls = make([]UnsafePushCall, 0)
	m.ConsumeCalls = 0
	m.CloseCalls = 0
}

// Ensure MockClient implements mq.ClientInterface.
var _ mq.ClientInterface = (*MockClient)(nil)
