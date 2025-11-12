package authorization

import (
	"net/http"
	"time"

	"172.21.5.249/airtrans/at-flight-authorization/internal/hapi"
	"172.21.5.249/airtrans/at-flight-authorization/internal/service"
	"172.21.5.249/airtrans/at-flight-authorization/internal/types"
	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog/log"
	"golang.org/x/text/language"
)

func CountAuthorizationLogsRoute(s *hapi.Server) *echo.Route {
	return s.Router.Root.GET("/authorization-logs/count", countAuthorizationLogsHandler(s))
}

func countAuthorizationLogsHandler(s *hapi.Server) echo.HandlerFunc {
	return func(c echo.Context) error {
		f := service.AuthorizationLogFilter{
			Username: c.QueryParam("username"),
			Resource: c.QueryParam("resource"),
			Action:   c.QueryParam("action"),
			Client:   c.QueryParam("client"),
			Result:   c.QueryParam("result"),
			URI:      c.QueryParam("uri"),
		}

		if sp := c.QueryParam("start"); sp != "" {
			t, err := time.Parse(time.RFC3339, sp)
			if err != nil {
				er := types.ErrorResponse{
					Message: "Invalid start time, time must be in RFC3339 format",
				}
				return c.JSON(http.StatusBadRequest, er)
			}
			f.StartTime = &t
		}

		if sp := c.QueryParam("end"); sp != "" {
			t, err := time.Parse(time.RFC3339, sp)
			if err != nil {
				er := types.ErrorResponse{
					Message: "Invalid end time, time must be in RFC3339 format",
				}
				return c.JSON(http.StatusBadRequest, er)
			}
			f.EndTime = &t
		}

		role, err := s.Service.Count(c.Request().Context(), f)
		if err != nil {
			log.Error().Msgf("CountAuthorizationLogs error")
			er := types.ErrorResponse{
				Message: s.I18n.Translate(err.Error(), language.Vietnamese),
			}
			return c.JSON(http.StatusInternalServerError, er)
		}
		return c.JSON(http.StatusOK, role)
	}
}
