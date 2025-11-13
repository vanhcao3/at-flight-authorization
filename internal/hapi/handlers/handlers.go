package handlers

import (
	"172.21.5.249/airtrans/at-flight-authorization/internal/hapi"
	"172.21.5.249/airtrans/at-flight-authorization/internal/hapi/handlers/common"
	"172.21.5.249/airtrans/at-flight-authorization/internal/hapi/handlers/flightauthorizationapproval"
	"172.21.5.249/airtrans/at-flight-authorization/internal/hapi/handlers/flightauthorizationproposal"
	"172.21.5.249/airtrans/at-flight-authorization/internal/hapi/handlers/flightnotification"
	ws "172.21.5.249/airtrans/at-flight-authorization/internal/hapi/handlers/websocket"

	"github.com/labstack/echo/v4"
)

func AttackAllRoutes(s *hapi.Server) {
	routes := []*echo.Route{
		common.GetVersionRoute(s),
		common.GetReadyRoute(s),
		common.GetHealthyRoute(s),
	}
	routes = append(routes, flightauthorizationproposal.RegisterRoutes(s)...)
	routes = append(routes, flightauthorizationapproval.RegisterRoutes(s)...)
	routes = append(routes, flightnotification.RegisterRoutes(s)...)
	routes = append(routes, ws.RegisterRoutes(s)...)
	s.Router.Routes = routes
}
