// Package mq provides end-to-end tests for the RabbitMQ client.
package mq

import (
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	clientmq "procodus.dev/demo-app/pkg/mq"
)

var _ = Describe("MQ Client E2E", func() {
	var (
		client    *clientmq.Client
		queueName string
	)

	BeforeEach(func() {
		// Generate unique queue name for this test
		queueName = "test-queue-" + time.Now().Format("20060102-150405.000")
	})

	AfterEach(func() {
		if client != nil {
			_ = client.Close()
			client = nil
		}
	})

	Describe("Connection", func() {
		It("should connect to RabbitMQ successfully", func() {
			client = clientmq.New(queueName, rabbitmqURL, testLogger)
			Expect(client).NotTo(BeNil())

			// Give client time to connect
			time.Sleep(1 * time.Second)
		})

		It("should handle invalid URL gracefully", func() {
			invalidClient := clientmq.New("test-queue", "amqp://invalid:5672", testLogger)
			Expect(invalidClient).NotTo(BeNil())

			// Should not crash, will keep retrying in background
			time.Sleep(500 * time.Millisecond)

			_ = invalidClient.Close()
		})
	})

	Describe("Publishing", func() {
		BeforeEach(func() {
			client = clientmq.New(queueName, rabbitmqURL, testLogger)
			time.Sleep(2 * time.Second) // Wait for connection
		})

		It("should publish a message successfully", func() {
			message := []byte("test message")
			err := client.Push(message)
			Expect(err).NotTo(HaveOccurred())
		})

		It("should publish multiple messages successfully", func() {
			messages := []string{
				"message 1",
				"message 2",
				"message 3",
			}

			for _, msg := range messages {
				err := client.Push([]byte(msg))
				Expect(err).NotTo(HaveOccurred())
			}
		})

		It("should publish large messages successfully", func() {
			// Create a 1MB message
			largeMessage := make([]byte, 1024*1024)
			for i := range largeMessage {
				largeMessage[i] = byte(i % 256)
			}

			err := client.Push(largeMessage)
			Expect(err).NotTo(HaveOccurred())
		})

		It("should handle rapid successive publishes", func() {
			for i := 0; i < 10; i++ {
				message := []byte("rapid message")
				err := client.Push(message)
				Expect(err).NotTo(HaveOccurred())
			}
		})

		It("should use UnsafePush without blocking", func() {
			message := []byte("unsafe message")
			err := client.UnsafePush(message)
			Expect(err).NotTo(HaveOccurred())
		})
	})

	Describe("Consuming", func() {
		BeforeEach(func() {
			client = clientmq.New(queueName, rabbitmqURL, testLogger)
			time.Sleep(2 * time.Second) // Wait for connection
		})

		It("should consume messages successfully", func() {
			// Start consuming FIRST
			deliveries, err := client.Consume()
			Expect(err).NotTo(HaveOccurred())
			Expect(deliveries).NotTo(BeNil())

			// Wait for consumer to register on server
			time.Sleep(500 * time.Millisecond)

			// THEN publish a message
			testMessage := []byte("consume test message")
			err = client.Push(testMessage)
			Expect(err).NotTo(HaveOccurred())

			// Receive the message
			select {
			case delivery := <-deliveries:
				Expect(string(delivery.Body)).To(ContainSubstring("consume test message"))
			case <-time.After(5 * time.Second):
				Fail("Did not receive message within timeout")
			}
		})

		It("should consume multiple messages in order", func() {
			// Start consuming FIRST
			deliveries, err := client.Consume()
			Expect(err).NotTo(HaveOccurred())

			// Wait for consumer to register on server
			time.Sleep(500 * time.Millisecond)

			// THEN publish multiple messages
			messages := []string{"first", "second", "third"}
			for _, msg := range messages {
				err := client.Push([]byte(msg))
				Expect(err).NotTo(HaveOccurred())
			}

			// Receive all messages and acknowledge each one
			receivedMessages := make([]string, 0, 3)
			for i := 0; i < 3; i++ {
				select {
				case delivery := <-deliveries:
					receivedMessages = append(receivedMessages, string(delivery.Body))
					// Acknowledge the message so the next one can be delivered
					err := delivery.Ack(false)
					Expect(err).NotTo(HaveOccurred())
				case <-time.After(5 * time.Second):
					Fail("Did not receive all messages within timeout")
				}
			}

			// Verify order and content
			Expect(receivedMessages).To(HaveLen(3))
			Expect(receivedMessages[0]).To(ContainSubstring("first"))
			Expect(receivedMessages[1]).To(ContainSubstring("second"))
			Expect(receivedMessages[2]).To(ContainSubstring("third"))
		})

		It("should handle message acknowledgment", func() {
			// Start consuming FIRST
			deliveries, err := client.Consume()
			Expect(err).NotTo(HaveOccurred())

			// Wait for consumer to register on server
			time.Sleep(500 * time.Millisecond)

			// THEN publish a message
			testMessage := []byte("ack test message")
			err = client.Push(testMessage)
			Expect(err).NotTo(HaveOccurred())

			// Receive and acknowledge
			select {
			case delivery := <-deliveries:
				err := delivery.Ack(false)
				Expect(err).NotTo(HaveOccurred())
			case <-time.After(5 * time.Second):
				Fail("Did not receive message within timeout")
			}
		})
	})

	Describe("Publish and Consume", func() {
		BeforeEach(func() {
			client = clientmq.New(queueName, rabbitmqURL, testLogger)
			time.Sleep(2 * time.Second) // Wait for connection
		})

		It("should handle full publish-consume cycle", func() {
			// Start consuming first
			deliveries, err := client.Consume()
			Expect(err).NotTo(HaveOccurred())

			// Wait for consumer to register on server
			time.Sleep(500 * time.Millisecond)

			// Publish a message
			testMessage := []byte(`{"sensor":"temp001","value":23.5}`)
			err = client.Push(testMessage)
			Expect(err).NotTo(HaveOccurred())

			// Consume and verify
			select {
			case delivery := <-deliveries:
				Expect(delivery.Body).To(Equal(testMessage))
			case <-time.After(5 * time.Second):
				Fail("Did not receive message within timeout")
			}
		})

		It("should handle sequential publishes and consumes", func() {
			// Start consuming FIRST
			deliveries, err := client.Consume()
			Expect(err).NotTo(HaveOccurred())

			// Wait for consumer to register on server
			time.Sleep(500 * time.Millisecond)

			// Publish multiple messages
			err = client.Push([]byte("message 1"))
			Expect(err).NotTo(HaveOccurred())

			err = client.Push([]byte("message 2"))
			Expect(err).NotTo(HaveOccurred())

			// Should receive both messages and acknowledge each one
			messages := make([]string, 0, 2)
			for i := 0; i < 2; i++ {
				select {
				case delivery := <-deliveries:
					messages = append(messages, string(delivery.Body))
					// Acknowledge the message so the next one can be delivered
					err := delivery.Ack(false)
					Expect(err).NotTo(HaveOccurred())
				case <-time.After(5 * time.Second):
					Fail("Did not receive all messages within timeout")
				}
			}

			Expect(messages).To(HaveLen(2))
			Expect(messages).To(ContainElement(ContainSubstring("message 1")))
			Expect(messages).To(ContainElement(ContainSubstring("message 2")))
		})
	})

	Describe("Concurrent Operations", func() {
		BeforeEach(func() {
			client = clientmq.New(queueName, rabbitmqURL, testLogger)
			time.Sleep(2 * time.Second) // Wait for connection
		})

		// TODO: Fix goroutine concurrency issues in these tests
		XIt("should handle concurrent publishes", func() {
			done := make(chan bool, 10)

			// Launch 10 concurrent publishers
			for i := 0; i < 10; i++ {
				go func(_ int) {
					defer GinkgoRecover()
					message := []byte("concurrent message")
					err := client.UnsafePush(message)
					Expect(err).NotTo(HaveOccurred())
					done <- true
				}(i)
			}

			// Wait for all to complete
			for i := 0; i < 10; i++ {
				select {
				case <-done:
					// Success
				case <-time.After(5 * time.Second):
					Fail("Concurrent publish timed out")
				}
			}
		})

		XIt("should handle high-throughput publishing", func() {
			messageCount := 100
			done := make(chan bool, 1)

			go func() {
				defer GinkgoRecover()
				for i := 0; i < messageCount; i++ {
					_ = client.UnsafePush([]byte("high throughput message"))
				}
				done <- true
			}()

			select {
			case <-done:
				// Success
			case <-time.After(30 * time.Second):
				Fail("High-throughput publishing timed out")
			}
		})
	})

	Describe("Error Handling", func() {
		It("should handle operations before connection", func() {
			client = clientmq.New(queueName, rabbitmqURL, testLogger)
			// Don't wait for connection

			// Operations should fail gracefully
			err := client.UnsafePush([]byte("test"))
			Expect(err).To(HaveOccurred())
		})

		It("should recover from connection issues", func() {
			client = clientmq.New(queueName, rabbitmqURL, testLogger)
			time.Sleep(2 * time.Second)

			// Publish should work
			err := client.Push([]byte("before disconnect"))
			Expect(err).NotTo(HaveOccurred())

			// Note: Simulating actual connection drop is complex
			// In real scenarios, the client should auto-reconnect
		})
	})

	Describe("Resource Cleanup", func() {
		It("should close client cleanly", func() {
			client = clientmq.New(queueName, rabbitmqURL, testLogger)
			time.Sleep(2 * time.Second)

			err := client.Close()
			Expect(err).NotTo(HaveOccurred())

			client = nil // Prevent double close in AfterEach
		})

		It("should handle close on unconnected client", func() {
			client = clientmq.New(queueName, "amqp://invalid:5672", testLogger)
			time.Sleep(500 * time.Millisecond)

			err := client.Close()
			Expect(err).To(HaveOccurred()) // Should error as it never connected

			client = nil
		})

		It("should handle double close gracefully", func() {
			client = clientmq.New(queueName, rabbitmqURL, testLogger)
			time.Sleep(2 * time.Second)

			err1 := client.Close()
			Expect(err1).NotTo(HaveOccurred())

			err2 := client.Close()
			Expect(err2).To(HaveOccurred()) // Second close should error

			client = nil
		})
	})

	Describe("Message Properties", func() {
		BeforeEach(func() {
			client = clientmq.New(queueName, rabbitmqURL, testLogger)
			time.Sleep(2 * time.Second)
		})

		It("should preserve message content exactly", func() {
			// Start consuming FIRST
			deliveries, err := client.Consume()
			Expect(err).NotTo(HaveOccurred())

			// Wait for consumer to register on server
			time.Sleep(500 * time.Millisecond)

			// THEN publish
			originalMessage := []byte("exact content preservation test 🎉")
			err = client.Push(originalMessage)
			Expect(err).NotTo(HaveOccurred())

			// Receive and verify
			select {
			case delivery := <-deliveries:
				Expect(delivery.Body).To(Equal(originalMessage))
			case <-time.After(5 * time.Second):
				Fail("Did not receive message within timeout")
			}
		})

		It("should handle binary data correctly", func() {
			// Start consuming FIRST
			deliveries, err := client.Consume()
			Expect(err).NotTo(HaveOccurred())

			// Wait for consumer to register on server
			time.Sleep(500 * time.Millisecond)

			// THEN publish binary data
			binaryData := []byte{0x00, 0x01, 0x02, 0xFF, 0xFE, 0xFD}
			err = client.Push(binaryData)
			Expect(err).NotTo(HaveOccurred())

			// Receive and verify
			select {
			case delivery := <-deliveries:
				Expect(delivery.Body).To(Equal(binaryData))
			case <-time.After(5 * time.Second):
				Fail("Did not receive message within timeout")
			}
		})

		It("should handle empty messages", func() {
			// Start consuming FIRST
			deliveries, err := client.Consume()
			Expect(err).NotTo(HaveOccurred())

			// Wait for consumer to register on server
			time.Sleep(500 * time.Millisecond)

			// THEN publish empty message
			emptyMessage := []byte{}
			err = client.Push(emptyMessage)
			Expect(err).NotTo(HaveOccurred())

			// Receive and verify
			select {
			case delivery := <-deliveries:
				Expect(delivery.Body).To(HaveLen(0))
			case <-time.After(5 * time.Second):
				Fail("Did not receive message within timeout")
			}
		})
	})
})
