package alice

import (
	"fmt"

	"github.com/capcom6/alice-companion/internal/alice"
	"github.com/capcom6/alice-companion/internal/server/auth"
	"github.com/go-core-fx/fiberfx/handler"
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v3"
	"go.uber.org/zap"
)

// Handler serves the Alice Dialogs webhook endpoint. The route is authenticated
// by owner decision: an intentionally unauthenticated mode is NOT implemented,
// and any future change to that decision belongs in this comment.
type Handler struct {
	handler.Base

	secret string

	logger *zap.Logger
}

// New creates a new Alice webhook HTTP handler.
func New(cfg alice.Config, validator *validator.Validate, logger *zap.Logger) handler.Handler {
	return &Handler{
		Base: handler.Base{
			Validator: validator,
		},

		secret: cfg.SharedSecret,

		logger: logger,
	}
}

// Register registers the Alice webhook route on the version 1 router.
func (h *Handler) Register(r fiber.Router) {
	r.Post("/alice/webhook", auth.New(h.secret), h.handle)
}

//	@Summary		Handle Alice Dialogs request
//	@Description	Accepts an Alice Dialogs request and answers a greeting on a new session or an echo on a repeat utterance. Every request must carry the shared secret. The route performs no memory or IoT access.
//	@Tags			alice
//	@Accept			json
//	@Produce		json
//	@Param			key		query		string					true	"Shared secret from ALICE__SHARED_SECRET"
//	@Param			request	body		alice.WebhookRequest	true	"Alice Dialogs request"
//	@Success		200		{object}	alice.Response
//	@Failure		400		{object}	fiberfx.ErrorResponse
//	@Failure		401		{object}	fiberfx.ErrorResponse
//	@Router			/alice/webhook [post]
//
// handle decodes the Dialogs request and answers it. Authentication runs in
// the keyauth middleware registered on this route.
func (h *Handler) handle(c fiber.Ctx) error {
	var req alice.WebhookRequest
	if err := h.BodyParserValidator(c, &req); err != nil {
		return fmt.Errorf("invalid request: %w", err)
	}

	return c.JSON(alice.Reply(&req))
}
