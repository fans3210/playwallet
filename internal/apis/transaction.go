package apis

import (
	"net/http"

	"playwallet/internal/domain"
	"playwallet/pkg/errs"

	"github.com/labstack/echo/v4"
)

func (s *App) makeTransaction(c echo.Context) error {
	var req domain.TransactionReq
	if err := c.Bind(&req); err != nil {
		return errs.ErrInvalidParam
	}
	if err := req.Validate(); err != nil {
		return err
	}
	if err := s.uc.MakeTransaction(req); err != nil {
		return err
	}

	resMsg := "succeed"
	if req.Type == domain.Transfer {
		resMsg = "request sent"
	}
	return c.JSON(http.StatusOK, map[string]any{
		"message": resMsg,
	})
}
