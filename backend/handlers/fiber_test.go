package handlers

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
)

type TestPayload struct {
	Name string `json:"name"`
}

func TestFiberBodyParser(t *testing.T) {
	app := fiber.New()

	app.Post("/test", func(c *fiber.Ctx) error {
		var p TestPayload
		if err := c.BodyParser(&p); err != nil {
			return err
		}
		return c.JSON(p)
	})

	payload := TestPayload{Name: "Fiber"}
	body, _ := json.Marshal(payload)

	req := httptest.NewRequest(http.MethodPost, "/test", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.ContentLength = int64(len(body))

	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	respBody, _ := io.ReadAll(resp.Body)
	assert.Contains(t, string(respBody), "Fiber")
}
