package main

import (
	"fmt"
	"net"
	"testing"
	"time"
)

func TestOutOfOrderDetector_CheckMessage(t *testing.T) {
	tests := []struct {
		messages         []string
		expectedMissed   int
		expectedReceived int
	}{
		{
			messages:         []string{"SequenceNumber:1", "SequenceNumber:2", "SequenceNumber:3"},
			expectedMissed:   0,
			expectedReceived: 3,
		},
		{
			messages:         []string{"SequenceNumber:1", "SequenceNumber:3"},
			expectedMissed:   1,
			expectedReceived: 2,
		},
		{
			messages:         []string{"SequenceNumber:1", "SequenceNumber:5"},
			expectedMissed:   3,
			expectedReceived: 2,
		},
	}

	for i, tt := range tests {
		t.Run(fmt.Sprintf("Case%d", i+1), func(t *testing.T) {
			detector := NewOutOfOrderDetector()
			missedCount := 0

			for _, msg := range tt.messages {
				detector.CheckMessage(msg)
			}

			// Calculate the number of missed messages
			missedCount = detector.expectedSeqNum - tt.expectedReceived - 1

			if missedCount != tt.expectedMissed {
				t.Errorf("Expected %d missed messages, but got %d", tt.expectedMissed, missedCount)
			}
		})
	}
}

func TestReceiver_Start(t *testing.T) {
	address := "224.0.0.1:9999"
	receiver, err := NewReceiver(address)
	if err != nil {
		t.Fatal("Error creating receiver:", err)
	}

	receivedCh := make(chan string, 10)

	// Run Receiver in a goroutine
	go func() {
		buf := make([]byte, 1024)
		n, _, err := receiver.conn.ReadFromUDP(buf)
		if err != nil {
			t.Error("Error reading:", err)
			return
		}
		receivedCh <- string(buf[:n])
	}()

	// Wait for Receiver to start receiving
	time.Sleep(100 * time.Millisecond)

	// Send data to the multicast address
	go func() {
		conn, err := net.Dial("udp", address)
		if err != nil {
			t.Error("Error reading:", err)
			return
		}
		defer conn.Close()

		_, err = conn.Write([]byte("Test Message"))
		if err != nil {
			t.Error("Error reading:", err)
			return
		}
	}()

	select {
	case received := <-receivedCh:
		if received != "Test Message" {
			t.Errorf("Expected 'Test Message', got '%s'", received)
		}
	case <-time.After(1 * time.Second):
		t.Error("Timeout waiting for message")
	}

	receiver.Stop()
}
