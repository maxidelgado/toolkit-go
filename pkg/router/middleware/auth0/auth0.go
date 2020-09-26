package auth0

import (
	"encoding/json"
	"errors"
	"github.com/dgrijalva/jwt-go"
	"github.com/gofiber/fiber"
	"github/maxidelgado/toolkit-go/pkg/ctxhelper"
	"github/maxidelgado/toolkit-go/pkg/logger"
	"go.uber.org/zap"
	"net/http"
	"strings"
)

const (
	UserHeader = "X-User-Id"
)

var (
	config Config
)

type Config struct {
	Audience  string // http://localhost:3000
	Authority string // https://dev-xxxxxx.eu.auth0.com/
	Jwks      Jwks
}

type Response struct {
	Message string `json:"message"`
}

type Jwks struct {
	Keys []JSONWebKeys `json:"keys"`
}

type JSONWebKeys struct {
	Kty string   `json:"kty"`
	Kid string   `json:"kid"`
	Use string   `json:"use"`
	N   string   `json:"n"`
	E   string   `json:"e"`
	X5c []string `json:"x5c"`
}

func validationKeyGetter(token *jwt.Token) (interface{}, error) {
	// Verify 'aud' claim
	checkAud := token.Claims.(jwt.MapClaims).VerifyAudience(config.Audience, false)
	if !checkAud {
		return token, errors.New("invalid audience")
	}
	// Verify 'iss' claim
	checkIss := token.Claims.(jwt.MapClaims).VerifyIssuer(config.Authority, false)
	if !checkIss {
		return token, errors.New("invalid issuer")
	}

	cert, err := getPemCert(token)
	if err != nil {
		return nil, err
	}

	result, _ := jwt.ParseRSAPublicKeyFromPEM([]byte(cert))
	return result, nil
}

func getPemCert(token *jwt.Token) (string, error) {
	cert := ""
	var jwks = config.Jwks
	if len(jwks.Keys) == 0 {
		resp, err := http.Get(config.Authority + ".well-known/jwks.json")

		if err != nil {
			return cert, err
		}
		defer resp.Body.Close()

		err = json.NewDecoder(resp.Body).Decode(&jwks)
		if err != nil {
			return cert, err
		}
	}

	for k := range jwks.Keys {
		if token.Header["kid"] == jwks.Keys[k].Kid {
			cert = "-----BEGIN CERTIFICATE-----\n" + jwks.Keys[k].X5c[0] + "\n-----END CERTIFICATE-----"
		}
	}

	if cert == "" {
		err := errors.New("unable to find appropriate key")
		return cert, err
	}

	return cert, nil
}

func extractor(c *fiber.Ctx) (string, error) {
	authHeader := c.Get("Authorization")
	if authHeader == "" {
		return "", nil
	}
	authHeaderParts := strings.Split(authHeader, " ")
	if len(authHeaderParts) != 2 || strings.ToLower(authHeaderParts[0]) != "bearer" {
		return "", errors.New("authorization header format must be Bearer {token}")
	}
	return authHeaderParts[1], nil
}

func checkJWT(c *fiber.Ctx) bool {
	token, err := extractor(c)
	log := logger.WithContext(c.Context())
	if err != nil {
		log.Error("error extracting jwt", zap.Error(err))
		return false
	}

	parsedToken, err := jwt.Parse(token, validationKeyGetter)
	if err != nil {
		log.Error("error parsing token", zap.Error(err))
		return false
	}

	claims := parsedToken.Claims.(jwt.MapClaims)
	ctx := ctxhelper.WithContext(c.Context())
	ctx.SetUser(ctxhelper.User{
		Id: claims["sub"].(string),
	})

	return parsedToken.Valid
}

// Protected does check your JWT token and validates it
func Protected(cfg Config) func(*fiber.Ctx) {
	config = cfg
	return func(c *fiber.Ctx) {
		if config.Audience == "" || config.Authority == "" {
			err := fiber.NewError(http.StatusInternalServerError, "missing jwt config")
			c.Next(err)
			return
		}
		if checkJWT(c) {
			c.Next()
		} else {
			c.Next(fiber.NewError(http.StatusUnauthorized, "this route is protected"))
		}
	}
}
