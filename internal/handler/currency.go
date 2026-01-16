package handler

import (
	"currency-converter-v2/internal/model"
	"currency-converter-v2/internal/service"
	"net/http"

	"github.com/gin-gonic/gin"
)

type CurrencyHandler struct {
	currencyService service.CurrencyServiceInterface
}

func NewCurrencyHandler(currencyService service.CurrencyServiceInterface) *CurrencyHandler {
	return &CurrencyHandler{
		currencyService: currencyService,
	}
}

func (h *CurrencyHandler) Convert(c *gin.Context) {
	var req model.ConvertRequest
	err := c.ShouldBindQuery(&req)
	if err != nil {
		c.JSON(http.StatusBadRequest, model.ErrorResponse{
			Error:   "Invalid request",
			Details: err.Error(),
		})
		return
	}
	result, rate, err := h.currencyService.Convert(c.Request.Context(), req.From, req.To, req.Amount)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.ErrorResponse{
			Error:   "Conversion failed",
			Details: err.Error(),
		})
		return
	}
	c.JSON(200, model.ConvertResponse{
		From:   req.From,
		To:     req.To,
		Amount: req.Amount,
		Rate:   rate,
		Result: result,
	})
}
