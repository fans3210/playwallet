package middlewares

import (
	"errors"
	"log/slog"
	"net/http"

	"playwallet/pkg/errs"

	"github.com/labstack/echo/v4"
)

func ErrorConvMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		err := next(c)
		if err == nil {
			return nil
		}
		slog.Error("[InternalErrConv] Log Err", "error", err)
		errMsg := err.Error()
		switch {
		case errors.Is(err, errs.ErrInvalidParam):
			return echo.NewHTTPError(http.StatusBadRequest, errMsg)
		case errors.Is(err, errs.ErrInvalidPlayer):
			return echo.NewHTTPError(http.StatusBadRequest, errMsg)
		default:
			// geenric error, prefer not to expose err msg
			return echo.NewHTTPError(http.StatusInternalServerError, "internal server error1")
		}
	}
}
