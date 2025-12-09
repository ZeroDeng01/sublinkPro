package routers

import (
	"io"
	"sublink/middlewares"
	"sublink/services/sse"

	"github.com/gin-gonic/gin"
)

// SSE registers the Server-Sent Events route
func SSE(r *gin.Engine) {
	r.GET("/api/sse", middlewares.AuthToken, func(c *gin.Context) {
		broker := sse.GetSSEBroker()

		// Create a buffered channel for this client to handle rapid progress updates
		clientChan := make(chan []byte, 50)

		// Register this client
		broker.AddClient(clientChan)

		defer func() {
			// Unregister client when connection closes
			broker.RemoveClient(clientChan)
			// Safely close the channel - it may already be closed by the broker
			// if it was blocked during a broadcast
			func() {
				defer func() {
					if r := recover(); r != nil {
						// Channel was already closed by the broker, ignore
					}
				}()
				close(clientChan)
			}()
		}()

		// Set headers for SSE
		c.Writer.Header().Set("Content-Type", "text/event-stream")
		c.Writer.Header().Set("Cache-Control", "no-cache")
		c.Writer.Header().Set("Connection", "keep-alive")
		c.Writer.Header().Set("Transfer-Encoding", "chunked")

		// Flush the headers to ensure the client receives the response immediately
		c.Writer.Flush()

		// Listen for messages from the broker
		c.Stream(func(w io.Writer) bool {
			// Wait for a message from the broker
			if msg, ok := <-clientChan; ok {
				c.Writer.Write(msg)
				c.Writer.Flush()
				return true
			}
			return false
		})
	})
}
