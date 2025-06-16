package display

import (
	"context"
	"fmt"
	"sync"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

// MessageType represents different types of display messages
type MessageType int

const (
	MsgTypeStep MessageType = iota
	MsgTypeProcessComplete
	MsgTypeLiveUpdate
	MsgTypeStepMessage
	MsgTypeSystemStats
	MsgTypeCustom
)

// DisplayMessage represents a message to be processed by the display
type DisplayMessage struct {
	Type      MessageType            `json:"type"`
	StepNum   int                    `json:"step_num,omitempty"`
	Data      map[string]interface{} `json:"data"`
	Timestamp time.Time              `json:"timestamp"`
}

// AsyncManager handles asynchronous message processing for the display
type AsyncManager struct {
	config       *AsyncConfig
	messageQueue chan DisplayMessage
	program      *tea.Program
	ctx          context.Context
	cancel       context.CancelFunc
	wg           sync.WaitGroup
	mutex        sync.RWMutex
	running      bool
	
	// Message batching
	batchBuffer []DisplayMessage
	lastUpdate  time.Time
}

// NewAsyncManager creates a new async manager
func NewAsyncManager(config *AsyncConfig) *AsyncManager {
	ctx, cancel := context.WithCancel(context.Background())
	
	return &AsyncManager{
		config:       config,
		messageQueue: make(chan DisplayMessage, config.QueueSize),
		ctx:          ctx,
		cancel:       cancel,
		batchBuffer:  make([]DisplayMessage, 0, config.MaxBatchSize),
		lastUpdate:   time.Now(),
	}
}

// Start begins async message processing
func (am *AsyncManager) Start(program *tea.Program) error {
	am.mutex.Lock()
	defer am.mutex.Unlock()
	
	if am.running {
		return fmt.Errorf("async manager is already running")
	}
	
	am.program = program
	am.running = true
	
	// Start message processor
	am.wg.Add(1)
	go am.processMessages()
	
	// Start batch processor
	am.wg.Add(1)
	go am.processBatches()
	
	return nil
}

// Stop gracefully stops the async manager
func (am *AsyncManager) Stop() error {
	am.mutex.Lock()
	defer am.mutex.Unlock()
	
	if !am.running {
		return nil
	}
	
	am.running = false
	am.cancel()
	
	// Close the message queue
	close(am.messageQueue)
	
	// Wait for all goroutines to finish
	am.wg.Wait()
	
	return nil
}

// SendMessage sends a message to the display (non-blocking)
func (am *AsyncManager) SendMessage(msg DisplayMessage) error {
	am.mutex.RLock()
	defer am.mutex.RUnlock()
	
	if !am.running {
		return fmt.Errorf("async manager is not running")
	}
	
	msg.Timestamp = time.Now()
	
	select {
	case am.messageQueue <- msg:
		return nil
	default:
		// Queue is full, drop the message (prevent blocking)
		return fmt.Errorf("message queue is full, message dropped")
	}
}

// SendStepUpdate sends a step update message
func (am *AsyncManager) SendStepUpdate(stepNum int) error {
	return am.SendMessage(DisplayMessage{
		Type:    MsgTypeStep,
		StepNum: stepNum,
		Data:    make(map[string]interface{}),
	})
}

// SendLiveUpdate sends a live update message
func (am *AsyncManager) SendLiveUpdate(stepNum int, key string, value interface{}) error {
	return am.SendMessage(DisplayMessage{
		Type:    MsgTypeLiveUpdate,
		StepNum: stepNum,
		Data: map[string]interface{}{
			"key":   key,
			"value": value,
		},
	})
}

// SendStepMessage sends a custom step message
func (am *AsyncManager) SendStepMessage(stepNum int, template string, data map[string]interface{}) error {
	msgData := map[string]interface{}{
		"template": template,
		"data":     data,
	}
	
	return am.SendMessage(DisplayMessage{
		Type:    MsgTypeStepMessage,
		StepNum: stepNum,
		Data:    msgData,
	})
}

// SendProcessComplete sends a process completion message
func (am *AsyncManager) SendProcessComplete(complete bool) error {
	return am.SendMessage(DisplayMessage{
		Type: MsgTypeProcessComplete,
		Data: map[string]interface{}{
			"complete": complete,
		},
	})
}

// SendSystemStats sends system statistics
func (am *AsyncManager) SendSystemStats(cpu, ram float64) error {
	return am.SendMessage(DisplayMessage{
		Type: MsgTypeSystemStats,
		Data: map[string]interface{}{
			"cpu": cpu,
			"ram": ram,
		},
	})
}

// processMessages handles incoming messages from the queue
func (am *AsyncManager) processMessages() {
	defer am.wg.Done()
	
	for {
		select {
		case msg, ok := <-am.messageQueue:
			if !ok {
				// Queue closed, flush remaining batch and exit
				am.flushBatch()
				return
			}
			am.addToBatch(msg)
			
		case <-am.ctx.Done():
			// Context cancelled, flush and exit
			am.flushBatch()
			return
		}
	}
}

// processBatches handles periodic batch processing
func (am *AsyncManager) processBatches() {
	defer am.wg.Done()
	
	ticker := time.NewTicker(am.config.UpdateInterval)
	defer ticker.Stop()
	
	for {
		select {
		case <-ticker.C:
			am.flushBatch()
			
		case <-am.ctx.Done():
			return
		}
	}
}

// addToBatch adds a message to the current batch
func (am *AsyncManager) addToBatch(msg DisplayMessage) {
	am.batchBuffer = append(am.batchBuffer, msg)
	
	// Flush if batch is full
	if len(am.batchBuffer) >= am.config.MaxBatchSize {
		am.flushBatch()
	}
}

// flushBatch processes all messages in the current batch
func (am *AsyncManager) flushBatch() {
	if len(am.batchBuffer) == 0 {
		return
	}
	
	am.mutex.RLock()
	program := am.program
	am.mutex.RUnlock()
	
	if program == nil {
		am.batchBuffer = am.batchBuffer[:0]
		return
	}
	
	// Process each message in the batch
	for _, msg := range am.batchBuffer {
		teaMsg := am.convertToTeaMessage(msg)
		if teaMsg != nil {
			program.Send(teaMsg)
		}
	}
	
	// Clear the batch
	am.batchBuffer = am.batchBuffer[:0]
	am.lastUpdate = time.Now()
}

// convertToTeaMessage converts a DisplayMessage to a Bubble Tea message
func (am *AsyncManager) convertToTeaMessage(msg DisplayMessage) tea.Msg {
	switch msg.Type {
	case MsgTypeStep:
		return stepMsg(msg.StepNum)
		
	case MsgTypeProcessComplete:
		if complete, ok := msg.Data["complete"].(bool); ok {
			return processCompleteMsg(complete)
		}
		
	case MsgTypeLiveUpdate:
		return liveUpdateMsg{
			StepNum: msg.StepNum,
			Key:     msg.Data["key"].(string),
			Value:   msg.Data["value"],
		}
		
	case MsgTypeStepMessage:
		return stepMessageMsg{
			StepNum:  msg.StepNum,
			Template: msg.Data["template"].(string),
			Data:     msg.Data["data"].(map[string]interface{}),
		}
		
	case MsgTypeSystemStats:
		cpu, _ := msg.Data["cpu"].(float64)
		ram, _ := msg.Data["ram"].(float64)
		disk, _ := msg.Data["disk"].(float64)
		return systemStatsMsg{cpu: cpu, ram: ram, disk: disk}
		
	default:
		return customMsg{
			Type: msg.Type,
			Data: msg.Data,
		}
	}
	
	return nil
}

// GetQueueStats returns statistics about the message queue
func (am *AsyncManager) GetQueueStats() map[string]interface{} {
	am.mutex.RLock()
	defer am.mutex.RUnlock()
	
	return map[string]interface{}{
		"queue_length":    len(am.messageQueue),
		"queue_capacity":  cap(am.messageQueue),
		"batch_size":      len(am.batchBuffer),
		"running":         am.running,
		"last_update":     am.lastUpdate,
	}
}

// New message types for enhanced functionality
type liveUpdateMsg struct {
	StepNum int
	Key     string
	Value   interface{}
}

type stepMessageMsg struct {
	StepNum  int
	Template string
	Data     map[string]interface{}
}

type customMsg struct {
	Type MessageType
	Data map[string]interface{}
}