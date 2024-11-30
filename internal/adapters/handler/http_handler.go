package handler

import (
	"net/http"

	"github.com/a-berahman/shopping-cart/internal/core/ports"

	"github.com/labstack/echo/v4"
)

type Handler struct {
	service ports.CartService
}

type AddItemRequest struct {
	Name     string `json:"name" validate:"required,min=2,max=100"`
	Quantity int    `json:"quantity" validate:"required,min=1"`
}

func NewHandler(service ports.CartService) *Handler {
	return &Handler{
		service: service,
	}
}

// Register registers the routes for the handler
func (h *Handler) Register(e *echo.Echo) {
	e.POST("api/v1/items", h.AddItem)
	e.GET("api/v1/items", h.ListItems)
}

func (h *Handler) AddItem(c echo.Context) error {
	ctx := c.Request().Context()
	var req AddItemRequest

	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	if err := c.Validate(req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	item, err := h.service.AddItemToCart(ctx, req.Name, req.Quantity)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusCreated, item)
}

func (h *Handler) ListItems(c echo.Context) error {
	ctx := c.Request().Context()

	items, err := h.service.ListCartItems(ctx)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, items)
}
