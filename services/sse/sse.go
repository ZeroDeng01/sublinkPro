package sse

import (
	"encoding/json"
	"fmt"
	"sublink/utils"
	"sync"
	"time"
)

// SSEBroker manages Server-Sent Events clients and broadcasting
type SSEBroker struct {
	// Events are pushed to this channel by the main events-gathering routine
	Notifier chan []byte

	// New client connections
	newClients chan chan []byte

	// Closed client connections
	closingClients chan chan []byte

	// Client connections registry
	clients map[chan []byte]bool

	// Mutex to protect the clients map
	mutex sync.Mutex
}

var (
	sseBroker *SSEBroker
	sseOnce   sync.Once
)

// GetSSEBroker returns the singleton instance of the SSEBroker
func GetSSEBroker() *SSEBroker {
	sseOnce.Do(func() {
		sseBroker = &SSEBroker{
			Notifier:       make(chan []byte, 100), // Buffer increased for rapid progress events
			newClients:     make(chan chan []byte),
			closingClients: make(chan chan []byte),
			clients:        make(map[chan []byte]bool),
		}
	})
	return sseBroker
}

// Listen starts the broker to listen for incoming and closing clients
func (broker *SSEBroker) Listen() {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case s := <-broker.newClients:
			// A new client has connected.
			// Register their message channel
			broker.mutex.Lock()
			broker.clients[s] = true
			broker.mutex.Unlock()
			utils.Info("Client added. %d registered clients", len(broker.clients))

		case s := <-broker.closingClients:
			// A client has detached and we want to stop sending them messages.
			broker.mutex.Lock()
			delete(broker.clients, s)
			broker.mutex.Unlock()
			utils.Info("Removed client. %d registered clients", len(broker.clients))

		case event := <-broker.Notifier:
			// We got a new event from the outside!
			// Send event to all connected clients
			broker.mutex.Lock()
			for clientMessageChan := range broker.clients {
				select {
				case clientMessageChan <- event:
				default:
					// If the client's channel buffer is full, just skip this message
					// The client will catch up with later messages
					// Don't disconnect - they might just be temporarily slow
				}
			}
			broker.mutex.Unlock()

		case <-ticker.C:
			// Send heartbeat to all clients
			broker.mutex.Lock()
			heartbeatMsg := []byte("event: heartbeat\ndata: ping\n\n")
			for clientMessageChan := range broker.clients {
				select {
				case clientMessageChan <- heartbeatMsg:
				default:
					// Client buffer full, skip heartbeat (they're probably processing messages)
				}
			}
			broker.mutex.Unlock()
		}
	}
}

// AddClient adds a client to the broker
func (broker *SSEBroker) AddClient(clientChan chan []byte) {
	broker.newClients <- clientChan
}

// RemoveClient removes a client from the broker
func (broker *SSEBroker) RemoveClient(clientChan chan []byte) {
	broker.closingClients <- clientChan
}

// Broadcast sends a message to all clients
func (broker *SSEBroker) Broadcast(message string) {
	broker.Notifier <- []byte(message)
}

// BroadcastJSONEvent sends structured JSON data to all SSE clients.
func (broker *SSEBroker) BroadcastJSONEvent(event string, payload interface{}) {
	jsonData, err := json.Marshal(payload)
	if err != nil {
		utils.Error("Error marshaling SSE payload: %v", err)
		return
	}
	msg := fmt.Sprintf("event: %s\ndata: %s\n\n", event, jsonData)
	broker.Notifier <- []byte(msg)
}

// ProgressPayload defines the payload for progress updates.
type ProgressPayload struct {
	TaskID      string      `json:"taskId"`              // 任务唯一标识
	TaskType    string      `json:"taskType"`            // 任务类型: speed_test, sub_update
	TaskName    string      `json:"taskName"`            // 任务名称（如订阅名称）
	Status      string      `json:"status"`              // started, progress, completed, error
	Current     int         `json:"current"`             // 当前进度
	Total       int         `json:"total"`               // 总数
	CurrentItem string      `json:"currentItem"`         // 当前处理的项目名称
	Result      interface{} `json:"result"`              // 当前项目的结果
	Message     string      `json:"message"`             // 可选的消息
	Time        string      `json:"time"`                // 时间戳
	StartTime   int64       `json:"startTime,omitempty"` // 任务开始时间戳(毫秒)
}

// BroadcastProgress sends a progress update to all clients.
func (broker *SSEBroker) BroadcastProgress(payload ProgressPayload) {
	// Ensure time is set
	if payload.Time == "" {
		payload.Time = time.Now().Format("2006-01-02 15:04:05")
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		utils.Error("Error marshaling SSE progress payload: %v", err)
		return
	}
	msg := fmt.Sprintf("event: task_progress\ndata: %s\n\n", jsonData)
	broker.Notifier <- []byte(msg)
}
