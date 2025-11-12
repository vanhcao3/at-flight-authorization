package authorization

import (
	"fmt"
	"net/http"
	"time"

	"172.21.5.249/airtrans/at-flight-authorization/internal/hapi"
	"172.21.5.249/airtrans/at-flight-authorization/internal/service"
	"172.21.5.249/airtrans/at-flight-authorization/internal/types"
	"github.com/labstack/echo/v4"
)

func StreamAuthorizationLogsCountRoute(s *hapi.Server) *echo.Route {
	return s.Router.Root.GET("/authorization-logs/count/stream", streamAuthorizationLogsCountHandler(s))
}

func streamAuthorizationLogsCountHandler(s *hapi.Server) echo.HandlerFunc {
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

		rp := c.Response()
		rq := c.Request()
		rph := rp.Header()

		rph.Set(echo.HeaderContentType, "text/event-stream")
		rph.Set(echo.HeaderCacheControl, "no-cache")
		rph.Set(echo.HeaderConnection, "keep-alive")
		rp.WriteHeader(http.StatusOK)

		flusher, ok := rp.Writer.(http.Flusher)
		if !ok {
			return echo.NewHTTPError(http.StatusInternalServerError, "Streaming unsupported")
		}

		ticker := time.NewTicker(2 * time.Second)
		defer ticker.Stop()

		var lastCount int64 = -1

		for {
			select {
			case <-rq.Context().Done():
				return nil
			case <-ticker.C:
				count, err := s.Service.Count(c.Request().Context(), f)
				if err != nil {
					return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
				}

				if count != lastCount {
					fmt.Fprintf(rp, "data: %d\n\n", count)
					lastCount = count
					flusher.Flush()
				}
			}
		}
	}
}
