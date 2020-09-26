package adapter

import (
	"context"
	"log"
	"net/http"
	"testing"

	"github.com/aws/aws-lambda-go/events"
	"github.com/gofiber/fiber"
	"github.com/stretchr/testify/assert"
)

func TestFiberLambda(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		log.Println("Starting test")
		app := fiber.New()
		app.Get("/ping", func(c *fiber.Ctx) {
			log.Println("Handler!!")
			c.SendString("pong")
		})

		adapter := New(app)

		req := events.APIGatewayProxyRequest{
			Path:       "/ping",
			HTTPMethod: "GET",
		}

		t.Run("Proxies with context the event correctly", func(t *testing.T) {
			resp, err := adapter.ProxyWithContext(context.Background(), req)

			assert.NoError(t, err)
			assert.Equal(t, http.StatusOK, resp.StatusCode)
		})

		t.Run("Proxies the event correctly", func(t *testing.T) {
			resp, err := adapter.Proxy(req)

			assert.NoError(t, err)
			assert.Equal(t, http.StatusOK, resp.StatusCode)
		})
	})
}
