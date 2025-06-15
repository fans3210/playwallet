package apis

import (
	"fmt"
	"net/http"
	"strconv"

	"playwallet/internal/domain"
	"playwallet/pkg/consts"
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

	resMsg := consts.DepositSuccessMsg
	if req.Type == domain.Transfer {
		resMsg = consts.TransferReqSent
	}
	return c.JSON(http.StatusOK, map[string]any{
		"message": resMsg,
	})
}

func (s *App) transactions(c echo.Context) error {
	pageOpt := domain.PageOpt{
		Page:    1,
		PerPage: 10,
	}
	if err := c.Bind(&pageOpt); err != nil {
		return errs.ErrInvalidParam
	}
	if !pageOpt.IsValid() {
		return errs.ErrInvalidParam
	}
	uidstr := c.Param("userid")
	userID, err := strconv.ParseInt(uidstr, 10, 64)
	if err != nil {
		return errs.ErrInvalidPlayer
	}
	total, trans, err := s.uc.Transactions(userID, pageOpt)
	if err != nil {
		return fmt.Errorf("failed to get transactions for user: %d, %w", userID, err)
	}
	res := domain.GetTransactionsRes{
		PageOpt: pageOpt,
		Total:   int(total),
		Records: trans,
	}
	return c.JSON(http.StatusOK, res)
}
