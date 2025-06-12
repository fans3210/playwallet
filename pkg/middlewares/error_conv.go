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
		var validationErr errs.ValidationError
		switch {
		case errors.Is(err, errs.ErrInvalidParam),
			errors.Is(err, errs.ErrInvalidPlayer),
			errors.As(err, &validationErr):
			return echo.NewHTTPError(http.StatusBadRequest, errMsg)
		case errors.Is(err, errs.ErrNotFound):
			return echo.NewHTTPError(http.StatusNotFound, errMsg)
		case errors.Is(err, errs.ErrNotAllowed):
			return echo.NewHTTPError(http.StatusForbidden, errMsg)
		case errors.Is(err, errs.ErrInsufficientBalance):
			return echo.NewHTTPError(http.StatusForbidden, errMsg)
		case errors.Is(err, errs.ErrDuplicate):
			return echo.NewHTTPError(http.StatusConflict, errMsg)
		default:
			// geenric error, prefer not to expose err msg
			return echo.NewHTTPError(http.StatusInternalServerError, "internal server error1")
		}
	}
}
