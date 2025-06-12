package apis

import (
	"fmt"
	"net/http"
	"strconv"

	"playwallet/pkg/errs"

	"github.com/labstack/echo/v4"
)

func (s *App) getBalacne(c echo.Context) error {
	uidstr := c.Param("userid")
	userID, err := strconv.ParseInt(uidstr, 10, 64)
	if err != nil {
		return errs.ErrInvalidPlayer
	}
	balanceInfo, err := s.uc.CheckBalance(userID)
	if err != nil {
		return fmt.Errorf("failed to check balance for user: %d, %w", userID, err)
	}
	return c.JSON(http.StatusOK, balanceInfo)
}
