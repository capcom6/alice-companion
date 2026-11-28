// Package auth provides API key authentication for the Alice webhook.
package auth

import (
	"crypto/subtle"

	"github.com/go-core-fx/fiberfx"
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/extractors"
	"github.com/gofiber/fiber/v3/middleware/keyauth"
)

// SecretQuery is the URL query parameter that carries the shared secret.
const SecretQuery = "key"

// New builds the API key middleware for the Alice webhook.
func New(secret string) fiber.Handler {
	return keyauth.New(keyauth.Config{
		Extractor: extractors.FromQuery(SecretQuery),
		Validator: func(_ fiber.Ctx, key string) (bool, error) {
			if secret == "" {
				return false, nil
			}

			return subtle.ConstantTimeCompare([]byte(key), []byte(secret)) == 1, nil
		},
		ErrorHandler: func(c fiber.Ctx, _ error) error {
			return c.Status(fiber.StatusUnauthorized).JSON(
				fiberfx.NewErrorResponse("unauthorized", fiber.StatusUnauthorized, nil),
			)
		},
	})
}
