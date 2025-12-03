package sse

import (
	"encoding/json"
	"fmt"
	"log"
	"sync"
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
			Notifier:       make(chan []byte, 1),
			newClients:     make(chan chan []byte),
			closingClients: make(chan chan []byte),
			clients:        make(map[chan []byte]bool),
		}
	})
	return sseBroker
}

// Listen starts the broker to listen for incoming and closing clients
func (broker *SSEBroker) Listen() {
	for {
		select {
		case s := <-broker.newClients:
			// A new client has connected.
			// Register their message channel
			broker.mutex.Lock()
			broker.clients[s] = true
			broker.mutex.Unlock()
			log.Printf("Client added. %d registered clients", len(broker.clients))

		case s := <-broker.closingClients:
			// A client has detached and we want to stop sending them messages.
			broker.mutex.Lock()
			delete(broker.clients, s)
			broker.mutex.Unlock()
			log.Printf("Removed client. %d registered clients", len(broker.clients))

		case event := <-broker.Notifier:
			// We got a new event from the outside!
			// Send event to all connected clients
			broker.mutex.Lock()
			for clientMessageChan := range broker.clients {
				select {
				case clientMessageChan <- event:
				default:
					// If the client's channel is blocked, remove the client
					// This prevents one slow client from blocking the entire broadcast
					log.Println("Client channel blocked, removing client")
					delete(broker.clients, clientMessageChan)
					close(clientMessageChan)
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

// BroadcastEvent sends a JSON message to all clients
// You can use this to send structured data
func (broker *SSEBroker) BroadcastEvent(event string, payload interface{}) {
	jsonData, err := json.Marshal(payload)
	if err != nil {
		log.Printf("Error marshaling SSE payload: %v", err)
		return
	}
	msg := fmt.Sprintf("event: %s\ndata: %s\n\n", event, jsonData)
	broker.Notifier <- []byte(msg)
}
