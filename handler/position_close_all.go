package handler

import (
	"net/http"

	"github.com/gofiber/fiber/v2"
)

func (h *Handler) CloseAllPositionHandler(c *fiber.Ctx) error {

	if err := h.mmakerClient.CancelAllOpenOrdersService(c.UserContext()); err != nil {
		return c.Status(500).JSON(map[string]string{"error": err.Error()})
	}
	if err := h.mmakerClient.CloseAllPosition(c.UserContext()); err != nil {
		return c.Status(500).JSON(map[string]string{"error": err.Error()})
	}

	return c.Status(http.StatusOK).JSON(map[string]string{"message": "success"})

}
