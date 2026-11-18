package alice

import (
	"crypto/subtle"

	"github.com/capcom6/alice-companion/internal/alice"
	"github.com/go-core-fx/fiberfx/handler"
	"github.com/gofiber/fiber/v3"
	"go.uber.org/zap"
)

// SecretHeader is the request header that carries the shared secret.
const SecretHeader = "X-Alice-Secret"

// Handler serves the Alice Dialogs webhook endpoint. The route is authenticated
// by owner decision: an intentionally unauthenticated mode is NOT implemented,
// and any future change to that decision belongs in this comment.
type Handler struct {
	secret string

	logger *zap.Logger
}

// New creates a new Alice webhook HTTP handler.
func New(cfg alice.Config, logger *zap.Logger) handler.Handler {
	return &Handler{
		secret: cfg.SharedSecret,
		logger: logger,
	}
}

// Register registers the Alice webhook route on the version 1 router.
func (h *Handler) Register(r fiber.Router) {
	r.Post("/alice/webhook", h.handle)
}

//	@Summary		Handle Alice Dialogs request
//	@Description	Accepts an Alice Dialogs request and answers a greeting on a new session or an echo on a repeat utterance. Every request must carry the shared secret. The route performs no memory or IoT access.
//	@Tags			alice
//	@Accept			json
//	@Produce		json
//	@Param			X-Alice-Secret	header		string					true	"Shared secret from ALICE__SHARED_SECRET"
//	@Param			request			body		alice.WebhookRequest	true	"Alice Dialogs request"
//	@Success		200				{object}	alice.Response
//	@Failure		400				{object}	ErrorResponse
//	@Failure		401				{object}	ErrorResponse
//	@Router			/alice/webhook [post]
//
// handle authorizes the caller, decodes the Dialogs request, and answers it.
func (h *Handler) handle(c fiber.Ctx) error {
	if !h.authorized(c) {
		return c.Status(fiber.StatusUnauthorized).JSON(ErrorResponse{Error: "unauthorized"})
	}

	var req alice.WebhookRequest
	if err := c.Bind().Body(&req); err != nil {
		h.logger.Warn("alice webhook decode failed", zap.Error(err))
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{Error: "invalid request"})
	}

	return c.JSON(alice.Reply(&req))
}

func (h *Handler) authorized(c fiber.Ctx) bool {
	if h.secret == "" {
		return false
	}

	provided := c.Get(SecretHeader)
	if provided == "" {
		return false
	}

	return subtle.ConstantTimeCompare([]byte(provided), []byte(h.secret)) == 1
}
