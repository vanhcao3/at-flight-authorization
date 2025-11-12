package handlers

import (
	"172.21.5.249/airtrans/at-flight-authorization/internal/hapi"
	"172.21.5.249/airtrans/at-flight-authorization/internal/hapi/handlers/common"

	"github.com/labstack/echo/v4"
)

func AttackAllRoutes(s *hapi.Server) {
	s.Router.Routes = []*echo.Route{
		// GET /-/version
		common.GetVersionRoute(s),
		// GET /-/ready
		common.GetReadyRoute(s),
		// GET /-/healthy
		common.GetHealthyRoute(s),
	}
}
