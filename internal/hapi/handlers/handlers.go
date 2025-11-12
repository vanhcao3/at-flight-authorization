package handlers

import (
	"172.21.5.249/airtrans/at-flight-authorization/internal/hapi"
	"172.21.5.249/airtrans/at-flight-authorization/internal/hapi/handlers/common"
	"172.21.5.249/airtrans/at-flight-authorization/internal/hapi/handlers/flightauthorizationapproval"
	"172.21.5.249/airtrans/at-flight-authorization/internal/hapi/handlers/flightauthorizationproposal"
	"172.21.5.249/airtrans/at-flight-authorization/internal/hapi/handlers/flightnotification"

	"github.com/labstack/echo/v4"
)

func AttackAllRoutes(s *hapi.Server) {
	routes := []*echo.Route{
		common.GetVersionRoute(s),
		common.GetReadyRoute(s),
		common.GetHealthyRoute(s),
	}
	routes = append(routes, flightauthorizationproposal.AttachRoutes(s)...)
	routes = append(routes, flightauthorizationapproval.AttachRoutes(s)...)
	routes = append(routes, flightnotification.AttachRoutes(s)...)
	s.Router.Routes = routes
}
