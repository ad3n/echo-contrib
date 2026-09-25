package helpers

import (
	"errors"
	"net/http"

	"github.com/ad3n/echo/v5"
)

func DefaultStatusResolver(c *echo.Context, err error) int {
	status := 0
	if err != nil {
		if sc, ok := err.(echo.HTTPStatusCoder); ok {
			return sc.StatusCode()
		}

		var sc echo.HTTPStatusCoder
		if errors.As(err, &sc) {
			return sc.StatusCode()
		}
	}

	if eResp, uErr := echo.UnwrapResponse(c.Response()); uErr == nil {
		if eResp.Committed {
			status = eResp.Status
		}
	}

	if err != nil && status == 0 {
		status = http.StatusInternalServerError
	}

	return status
}
