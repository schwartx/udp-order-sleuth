package main

import (
	"fmt"
	"net"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"sync"
	"syscall"
)

type OutOfOrderDetector struct {
	expectedSeqNum int
	mux            sync.Mutex
}

func NewOutOfOrderDetector() *OutOfOrderDetector {
	return &OutOfOrderDetector{
		expectedSeqNum: 1,
	}
}

// CheckMessage checks if the given message is out of order and updates the expected sequence number.
func (o *OutOfOrderDetector) CheckMessage(msg string) bool {
	o.mux.Lock()
	defer o.mux.Unlock()

	// We want to split the string to get the sequence number, e.g., get 20 from "SequenceNumber:20:MessageContent"
	parts := strings.Split(msg, ":")
	if len(parts) < 3 {
		fmt.Println("Invalid message format:", msg)
		return false
	}

	seqNum, err := strconv.Atoi(parts[1]) // Note that we now use parts[1] to get the sequence number
	if err != nil {
		fmt.Println("Failed to parse sequence number:", parts[1])
		return false
	}

	if seqNum == o.expectedSeqNum {
		o.expectedSeqNum++
		return false
	} else if seqNum > o.expectedSeqNum {
		missedMessages := seqNum - o.expectedSeqNum
		fmt.Printf("OutOfOrder: Missed %d messages. Expected %d but received %d.\n", missedMessages, o.expectedSeqNum, seqNum)
		o.expectedSeqNum = seqNum + 1
		return true
	}

	// If seqNum < o.expectedSeqNum, it means we received a duplicate message, which can be handled as needed
	fmt.Printf("Duplicate: Received duplicate message with sequence number %d.\n", seqNum)
	return false
}

type Receiver struct {
	address string
	conn    *net.UDPConn
	stopCh  chan bool
	doneCh  chan struct{}
}

func NewReceiver(address string) (*Receiver, error) {
	addr, err := net.ResolveUDPAddr("udp", address)
	if err != nil {
		return nil, err
	}

	conn, err := net.ListenMulticastUDP("udp", nil, addr)
	if err != nil {
		return nil, err
	}

	return &Receiver{
		address: address,
		conn:    conn,
		stopCh:  make(chan bool),
		doneCh:  make(chan struct{}),
	}, nil
}

func (r *Receiver) Start(handler func(string)) {
	defer close(r.doneCh)
	buf := make([]byte, 1024)
	// joinMulticastGroup(r.conn, *address)

	for {
		select {
		case <-r.stopCh:
			return
		default:
			n, _, err := r.conn.ReadFromUDP(buf)
			if err != nil {
				fmt.Println("Error reading:", err)
				continue
			}

			msg := string(buf[:n])
			fmt.Println("Received message:", msg)

			// Pass the message to the handler function, which will handle out-of-order detection and statistics, etc.
			handler(msg)
		}
	}
}

func (r *Receiver) Stop() {
	fmt.Println("Stopping receiver...")
	// r.stopCh <- true
	// <-r.doneCh // Wait for the Start function to end
	// r.conn.Close()
	fmt.Println("Receiver stopped.")
}

type ReceiverStatistics struct {
	mux             sync.Mutex
	totalReceived   int
	outOfOrderCount int
}

func NewReceiverStatistics() *ReceiverStatistics {
	return &ReceiverStatistics{}
}

// IncrementReceived increases the count of received messages
func (s *ReceiverStatistics) IncrementReceived() {
	s.mux.Lock()
	defer s.mux.Unlock()
	s.totalReceived++
}

// IncrementOutOfOrder increases the count of out-of-order messages
func (s *ReceiverStatistics) IncrementOutOfOrder() {
	s.mux.Lock()
	defer s.mux.Unlock()
	s.outOfOrderCount++
}

// Report prints the statistics
func (s *ReceiverStatistics) Report() {
	s.mux.Lock()
	defer s.mux.Unlock()
	fmt.Printf("Total Received Messages: %d\n", s.totalReceived)
	fmt.Printf("Out of Order Messages: %d\n", s.outOfOrderCount)
}

type ReceiverApp struct {
	detector *OutOfOrderDetector
	receiver *Receiver
	stats    *ReceiverStatistics
}

func NewReceiverApp(address string) (*ReceiverApp, error) {
	// Initialize the out-of-order detection module
	detector := NewOutOfOrderDetector()

	// Initialize the statistics module
	stats := NewReceiverStatistics()

	// Initialize the receiver module
	receiver, err := NewReceiver(address)
	if err != nil {
		return nil, err
	}

	app := &ReceiverApp{
		detector: detector,
		receiver: receiver,
		stats:    stats,
	}

	return app, nil
}

// Start the receiver

func (app *ReceiverApp) Start() {
	go app.receiver.Start(func(message string) {
		// Process each received message
		isOutOfOrder := app.detector.CheckMessage(message)
		app.stats.IncrementReceived()
		if isOutOfOrder {
			app.stats.IncrementOutOfOrder()
		}
	})
}

func (app *ReceiverApp) Stop() {
	app.receiver.Stop()
}

func receiver() {
	app, err := NewReceiverApp(*address)
	if err != nil {
		fmt.Println("Error initializing receiver app:", err)
		return
	}

	app.Start()

	// Create a channel to listen for system interrupt signals
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	// Block until an interrupt signal is received
	<-c

	fmt.Println("\nReceived interrupt signal. Exiting...")

	// Stop the receiver program
	app.Stop()

	// Print statistics
	app.stats.Report()
}
