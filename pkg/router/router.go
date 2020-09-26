package router

import (
	"net/http"
	"time"

	"github.com/gofiber/cors"
	"github.com/gofiber/fiber"
	"github.com/gofiber/fiber/middleware"
	"github.com/gofiber/helmet"
	"github.com/maxidelgado/toolkit-go/pkg/ctxhelper"
	"github.com/maxidelgado/toolkit-go/pkg/logger"
	"github.com/maxidelgado/toolkit-go/pkg/router/middleware/auth0"
	"go.uber.org/zap"
)

const (
	ApiKeyHeader = "X-Api-Key"
	UserIdHeader = "X-User-Id"

	defaultReadTimeout  = 10 * time.Second
	defaultWriteTimeout = 10 * time.Second
	defaultIdleTimeout  = 10 * time.Second
)

type router struct {
	Config
	app    *fiber.App
	logger *zap.Logger
}

func New(config Config) *router {
	app := fiber.New(&fiber.Settings{
		ErrorHandler: defaultErrorHandler,
	})

	r := &router{
		Config: config,
		app:    app,
		logger: logger.Logger(&config.Logging),
	}

	// App config
	app.Settings.ReadTimeout = defaultReadTimeout
	app.Settings.WriteTimeout = defaultWriteTimeout
	app.Settings.IdleTimeout = defaultIdleTimeout

	if config.Timeout.Read != 0 {
		app.Settings.ReadTimeout = time.Duration(config.Timeout.Read) * time.Second
	}

	if config.Timeout.Write != 0 {
		app.Settings.WriteTimeout = time.Duration(config.Timeout.Write) * time.Second
	}

	if config.Timeout.Idle != 0 {
		app.Settings.IdleTimeout = time.Duration(config.Timeout.Idle) * time.Second
	}

	// Show Fiber logo on console for debug mode
	if config.Logging.Level != "debug" {
		app.Settings.DisableStartupMessage = true
	}

	app.Use(
		middleware.RequestID(),
		setupContext,
		logRequest,
		logResponse,
		middleware.Compress(),
		helmet.New(),
		cors.New(),
	)
	// Add default middleware

	// Add opt middleware
	if config.Protected {
		app.Use(auth0.Protected(config.Authorization))
	}

	return r
}

func (r *router) Engine() *fiber.App {
	return r.app
}

func (r *router) ValidateJWT(config auth0.Config) *router {
	r.app.Use(auth0.Protected(config))
	return r
}

func setupContext(c *fiber.Ctx) {
	rid := c.Fasthttp.Response.Header.Peek(fiber.HeaderXRequestID)
	uid := c.Get(UserIdHeader)
	apiKey := c.Get(ApiKeyHeader)

	ch := ctxhelper.WithContext(c.Context())
	ch.SetApiKey(ctxhelper.ApiKey{Id: apiKey})
	ch.SetUser(ctxhelper.User{Id: uid})
	ch.SetRequestId(string(rid))

	c.Locals(ctxhelper.Key, ch)
	c.Next()
}

func logRequest(c *fiber.Ctx) {
	log := logger.WithContext(c.Context())
	log.Info(
		"request",
		zap.String("method", c.Method()),
		zap.String("path", c.Path()),
	)
	c.Next()
}

func logResponse(c *fiber.Ctx) {
	start := time.Now()
	c.Next()
	duration := time.Since(start)
	log := logger.WithContext(c.Context())
	log.Info("response",
		zap.Int64("rt", duration.Milliseconds()),
		zap.Int("status", c.Fasthttp.Response.StatusCode()),
		zap.String("body", string(c.Fasthttp.Response.Body())),
	)
}

func defaultErrorHandler(ctx *fiber.Ctx, err error) {
	code := http.StatusInternalServerError
	if e, ok := err.(*fiber.Error); ok {
		code = e.Code
	}
	ctx.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)
	ctx.Status(code).SendString(err.Error())
}

type Config struct {
	Host          string        `yaml:"host"`
	Port          string        `yaml:"port"`
	UseAdapter    bool          `yaml:"useAdapter"`
	Protected     bool          `yaml:"protected"`
	Timeout       Timeout       `yaml:"timeout"`
	Logging       logger.Config `yaml:"logging"`
	Authorization auth0.Config  `yaml:"authorization"`
}

type Timeout struct {
	Write int `yaml:"write"`
	Read  int `yaml:"read"`
	Idle  int `yaml:"idle"`
}
