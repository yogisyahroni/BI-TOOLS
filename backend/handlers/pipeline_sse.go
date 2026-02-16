package handlers

import (
	"bufio"
	"encoding/json"
	"fmt"
	"insight-engine-backend/services"
	"time"

	"github.com/gofiber/fiber/v2"
)

// StreamPipelineStatus provides real-time execution updates via SSE
func StreamPipelineStatus(c *fiber.Ctx) error {
	pipelineID := c.Params("id")
	executionID := c.Query("executionId")

	if pipelineID == "" {
		return c.Status(400).JSON(fiber.Map{"error": "pipeline ID required"})
	}

	c.Set("Content-Type", "text/event-stream")
	c.Set("Cache-Control", "no-cache")
	c.Set("Connection", "keep-alive")
	c.Set("Transfer-Encoding", "chunked")
	c.Set("Access-Control-Allow-Origin", "*")
	c.Set("X-Accel-Buffering", "no") // Disable nginx buffering

	c.Context().SetBodyStreamWriter(func(w *bufio.Writer) {
		ticker := time.NewTicker(500 * time.Millisecond)
		defer ticker.Stop()

		timeout := time.After(30 * time.Minute)
		var lastProgress int = -1

		for {
			select {
			case <-timeout:
				writeSSEEvent(w, "timeout", map[string]interface{}{
					"message": "Stream timed out",
				})
				w.Flush()
				return
			case <-ticker.C:
				if services.GlobalPipelineExecutor == nil {
					continue
				}

				// Check if there's an active execution for this pipeline
				var activeExec *services.ExecutionContext
				if executionID != "" {
					activeExec = services.GlobalPipelineExecutor.GetActiveRun(executionID)
				}

				if activeExec == nil {
					// No active execution — send idle status
					writeSSEEvent(w, "status", map[string]interface{}{
						"pipelineId": pipelineID,
						"status":     "IDLE",
						"progress":   0,
					})
					w.Flush()

					// If we previously had progress, the execution is done
					if lastProgress > 0 {
						writeSSEEvent(w, "complete", map[string]interface{}{
							"pipelineId": pipelineID,
							"status":     "COMPLETED",
							"progress":   100,
						})
						w.Flush()
						return
					}
					continue
				}

				// Active execution — send progress update
				if activeExec.Progress != lastProgress {
					lastProgress = activeExec.Progress
					writeSSEEvent(w, "progress", map[string]interface{}{
						"pipelineId":  pipelineID,
						"executionId": activeExec.ExecutionID,
						"status":      activeExec.Status,
						"progress":    activeExec.Progress,
						"elapsedMs":   int(time.Since(activeExec.StartedAt).Milliseconds()),
					})
					w.Flush()

					if activeExec.Progress >= 100 {
						writeSSEEvent(w, "complete", map[string]interface{}{
							"pipelineId":  pipelineID,
							"executionId": activeExec.ExecutionID,
							"status":      "COMPLETED",
							"progress":    100,
						})
						w.Flush()
						return
					}
				}
			}
		}
	})

	return nil
}

// writeSSEEvent writes a Server-Sent Event to the buffered writer
func writeSSEEvent(w *bufio.Writer, event string, data interface{}) {
	jsonBytes, err := json.Marshal(data)
	if err != nil {
		return
	}

	fmt.Fprintf(w, "event: %s\n", event)
	fmt.Fprintf(w, "data: %s\n\n", string(jsonBytes))
}
