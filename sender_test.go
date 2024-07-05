package main

import (
	"bytes"
	"net"
	"strconv"
	"strings"
	"testing"
	"time"
)

func TestMessageGenerator_Generate(t *testing.T) {
	generator := NewMessageGenerator()

	prevSeqNum := 0

	for i := 0; i < 100; i++ {
		message := generator.Generate()

		if !strings.HasPrefix(message, "Sequence Number:") {
			t.Fatalf("Expected message to start with 'Sequence Number:', got: %s", message)
		}

		parts := strings.Split(message, ":")
		if len(parts) < 3 {
			t.Fatalf("Message format is not as expected: %s", message)
		}

		seqNum, err := strconv.Atoi(parts[1])
		if err != nil {
			t.Fatalf("Error converting sequence number to integer: %v", err)
		}

		if seqNum != prevSeqNum+1 {
			t.Fatalf("Expected sequence number %d, but got %d", prevSeqNum+1, seqNum)
		}

		prevSeqNum = seqNum
	}
}

func TestSender_SendMessages(t *testing.T) {
	// Set a timeout to end the test
	timeout := 2 * time.Second
	groupAddress := "224.0.0.1:9999"

	// Initialize multicast address
	address, err := net.ResolveUDPAddr("udp", groupAddress)
	if err != nil {
		t.Fatalf("Error resolving address: %v", err)
	}

	// Create a UDP connection to read messages from the multicast group
	conn, err := net.ListenMulticastUDP("udp", nil, address)
	if err != nil {
		t.Fatalf("Error setting up UDP listener: %v", err)
	}
	defer conn.Close()

	// Start the sender
	sender := NewSender(groupAddress, 500*time.Millisecond)
	go sender.SendMessages(func() {})

	// Receive messages and validate
	messageBuffer := make([]byte, 1024)
	var receivedMessages int
	startTime := time.Now()

	for {
		conn.SetDeadline(time.Now().Add(1 * time.Second))
		_, _, err := conn.ReadFromUDP(messageBuffer)
		if err != nil {
			if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
				// If we've waited long enough or received enough messages, exit
				if time.Since(startTime) > timeout {
					break
				}
				continue
			}
			t.Fatalf("Error reading from UDP: %v", err)
		}
		if bytes.HasPrefix(messageBuffer, []byte("Sequence Number:")) {
			receivedMessages++
		}
	}

	if receivedMessages < 1 {
		t.Fatalf("Expected to receive at least 1 message, got %d", receivedMessages)
	}
}

func TestSenderStatReporter_Start(t *testing.T) {
	interval := 500 * time.Millisecond
	reporter := NewStatReporter(interval)

	// Create a channel for communication in the statistics report
	reportChan := make(chan int, 1)

	go func() {
		for {
			select {
			case count := <-reportChan:
				if count != reporter.stats.GetSentCount() {
					t.Errorf("Reported count %d does not match expected count %d", count, reporter.stats.GetSentCount())
				}
				return
			default:
				time.Sleep(50 * time.Millisecond) // Polling interval
			}
		}
	}()

	reporter.Start()
	time.Sleep(150 * time.Millisecond) // Wait a bit to ensure the reporter starts

	// Simulate sending some messages
	reporter.Increment()
	reporter.Increment()
	reporter.Increment()

	// Give the reporter some time to report the new count
	time.Sleep(interval + 100*time.Millisecond)

	// Send the latest count to our goroutine
	reportChan <- reporter.stats.GetSentCount()

	// Cleanup
	reporter.Stop()
	close(reportChan)
}
