package apis

import (
	"net/http"
	"strings"

	"playwallet/internal/domain"
	"playwallet/pkg/errs"

	"github.com/labstack/echo/v4"
)

func (s *App) deposit(c echo.Context) error {
	var req domain.DepositReq
	if err := c.Bind(&req); err != nil {
		return errs.ErrInvalidParam
	}
	if req.Amt <= 0 {
		return errs.ValidationErrWithReason("deposit amt should be greater than 0")
	}
	if strings.TrimSpace(req.IdempotencyKey) == "" {
		return errs.ValidationErrWithReason("empty dempotency key")
	}
	if err := s.uc.Deposit(req); err != nil {
		return err
	}
	return c.JSON(http.StatusOK, map[string]any{
		"message": "succeed",
	})
}
