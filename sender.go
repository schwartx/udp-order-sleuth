package main

import (
	"fmt"
	"net"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

// MessageGenerator is used to generate messages with incrementing sequence numbers.
type MessageGenerator struct {
	seqNum    int
	seqNumMux sync.Mutex
}

// NewMessageGenerator initializes a new message generator.
func NewMessageGenerator() *MessageGenerator {
	return &MessageGenerator{}
}

// Generate creates a new message with an incrementing sequence number.
func (mg *MessageGenerator) Generate() string {
	mg.seqNumMux.Lock()
	defer mg.seqNumMux.Unlock()

	mg.seqNum++
	message := fmt.Sprintf("SequenceNumber:%d:MessageContent", mg.seqNum)
	return message
}

// Sender is used to send multicast messages.
type Sender struct {
	address      string
	messageGen   *MessageGenerator
	sendInterval time.Duration
}

// NewSender initializes a new Sender.
func NewSender(address string, interval time.Duration) *Sender {
	return &Sender{
		address:      address,
		messageGen:   NewMessageGenerator(),
		sendInterval: interval,
	}
}

// SendMessages sends multicast messages.
func (s *Sender) SendMessages(onSent func()) error {
	addr, err := net.ResolveUDPAddr("udp", s.address)
	if err != nil {
		return err
	}

	conn, err := net.DialUDP("udp", nil, addr)
	if err != nil {
		return err
	}
	defer conn.Close()

	for {
		message := s.messageGen.Generate()
		_, err := conn.Write([]byte(message))
		if err != nil {
			return err
		}

		if onSent != nil { // <- Call the callback here
			onSent()
		}

		fmt.Println("Sent:", message)
		time.Sleep(s.sendInterval)
	}
}

// SendStatistics tracks statistics for sent messages.
type SendStatistics struct {
	sentCount int
	mux       sync.Mutex
}

// Increment updates the count of sent messages.
func (s *SendStatistics) Increment() {
	s.mux.Lock()
	defer s.mux.Unlock()
	s.sentCount++
}

// GetSentCount retrieves the count of sent messages.
func (s *SendStatistics) GetSentCount() int {
	s.mux.Lock()
	defer s.mux.Unlock()
	return s.sentCount
}

// SenderStatReporter periodically reports statistics.
type SenderStatReporter struct {
	stats  *SendStatistics
	ticker *time.Ticker
	quit   chan struct{}
}

// NewStatReporter initializes a new statistics reporter.
func NewStatReporter(interval time.Duration) *SenderStatReporter {
	return &SenderStatReporter{
		stats:  &SendStatistics{},
		ticker: time.NewTicker(interval),
		quit:   make(chan struct{}),
	}
}

// Start begins periodic reporting of statistics.
func (sr *SenderStatReporter) Start() {
	go func() {
		for {
			select {
			case <-sr.ticker.C:
				sentCount := sr.stats.GetSentCount()
				fmt.Printf("Sent messages: %d\n", sentCount)
			case <-sr.quit:
				return
			}
		}
	}()
}

// Stop halts reporting and outputs final statistics.
func (sr *SenderStatReporter) Stop() {
	sr.ticker.Stop()
	close(sr.quit)

	// Output final statistics.
	sentCount := sr.stats.GetSentCount()
	fmt.Printf("Final count of sent messages: %d\n", sentCount)
}

// Increment updates statistics and notifies the reporter.
func (sr *SenderStatReporter) Increment() {
	sr.stats.Increment()
}

type SenderApp struct {
	sender     *Sender
	statReport *SenderStatReporter
}

func NewSenderApp(address string, sendInterval, reportInterval time.Duration) *SenderApp {
	return &SenderApp{
		sender:     NewSender(address, sendInterval),
		statReport: NewStatReporter(reportInterval),
	}
}

func (app *SenderApp) Start() {
	go app.statReport.Start()
	go app.sender.SendMessages(app.statReport.Increment)
}

func (app *SenderApp) Stop() {
	app.statReport.Stop()
	// For brevity, the sender is not closed here, but in a real application, a more graceful shutdown logic might be needed
}

func sender(sendInterval time.Duration) {
	reportInterval := 5 * time.Second

	app := NewSenderApp(*address, sendInterval, reportInterval)
	app.Start()

	// Create a channel to listen for system interrupt signals
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	// Block until an interrupt signal is received
	<-c

	fmt.Println("\nReceived interrupt signal. Exiting...")

	// Here you can add any cleanup or program stopping logic
	// For example: app.Stop()
	app.Stop()
}
