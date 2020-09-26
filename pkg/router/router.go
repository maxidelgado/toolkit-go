package router

import (
	"net/http"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/compress"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/fiber/v2/middleware/requestid"
	"github.com/gofiber/helmet/v2"
	"github.com/maxidelgado/toolkit-go/pkg/ctxhelper"
	"github.com/maxidelgado/toolkit-go/pkg/logger"
	"go.uber.org/zap"
)

const (
	ApiKeyHeader = "X-Api-Key"
	UserIdHeader = "X-User-Id"

	defaultReadTimeout  = 10 * time.Second
	defaultWriteTimeout = 10 * time.Second
	defaultIdleTimeout  = 10 * time.Second
)

type Handler interface {
	RegisterRoutes(app *fiber.App)
}

type router struct {
	Config
	app    *fiber.App
	logger *zap.Logger
}

func New(config Config) *router {
	app := fiber.New(getConfig(config))

	r := &router{
		Config: config,
		app:    app,
		logger: logger.Logger(&config.Logging),
	}

	app.Use(
		recover.New(),
		requestid.New(),
		setupContext,
		logRequest,
		logResponse,
		compress.New(),
		helmet.New(),
		cors.New(),
	)

	return r
}

func (r *router) Engine() *fiber.App {
	return r.app
}

func getConfig(config Config) fiber.Config {
	readTimeout := defaultReadTimeout
	writeTimeout := defaultWriteTimeout
	idleTimeout := defaultIdleTimeout
	disableStartupMsg := false

	switch {
	case config.Timeout.Read != 0:
		readTimeout = time.Duration(config.Timeout.Read) * time.Second
	case config.Timeout.Write != 0:
		writeTimeout = time.Duration(config.Timeout.Write) * time.Second
	case config.Timeout.Idle != 0:
		idleTimeout = time.Duration(config.Timeout.Idle) * time.Second
	case config.Logging.Level != "debug":
		disableStartupMsg = true
	}

	return fiber.Config{
		ErrorHandler:          defaultErrorHandler,
		ReadTimeout:           readTimeout,
		WriteTimeout:          writeTimeout,
		IdleTimeout:           idleTimeout,
		DisableStartupMessage: disableStartupMsg,
	}
}

func setupContext(c *fiber.Ctx) error {
	rid := c.Response().Header.Peek(fiber.HeaderXRequestID)
	uid := c.Get(UserIdHeader)
	apiKey := c.Get(ApiKeyHeader)

	ch := ctxhelper.WithContext(c.Context())
	ch.SetApiKey(ctxhelper.ApiKey{Id: apiKey})
	ch.SetUser(ctxhelper.User{Id: uid})
	ch.SetRequestId(string(rid))

	c.Locals(ctxhelper.Key, ch)
	return c.Next()
}

func logRequest(c *fiber.Ctx) error {
	log := logger.WithContext(c.Context())
	log.Info(
		"request",
		zap.String("method", c.Method()),
		zap.String("path", c.Path()),
	)
	return c.Next()
}

func logResponse(c *fiber.Ctx) error {
	start := time.Now()
	err := c.Next()
	duration := time.Since(start)
	log := logger.WithContext(c.Context())
	log.Info("response",
		zap.Int64("rt", duration.Milliseconds()),
		zap.Int("status", c.Response().StatusCode()),
		zap.String("body", string(c.Response().Body())),
	)
	return err
}

func defaultErrorHandler(ctx *fiber.Ctx, err error) error {
	code := http.StatusInternalServerError
	if e, ok := err.(*fiber.Error); ok {
		code = e.Code
	}
	ctx.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)
	return ctx.Status(code).JSON(err.Error())
}

type Config struct {
	Timeout Timeout       `yaml:"timeout"`
	Logging logger.Config `yaml:"logging"`
}

type Timeout struct {
	Write int `yaml:"write"`
	Read  int `yaml:"read"`
	Idle  int `yaml:"idle"`
}
