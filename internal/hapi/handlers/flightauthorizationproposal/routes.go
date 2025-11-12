package flightauthorizationproposal

import (
	"errors"
	"net/http"

	"172.21.5.249/airtrans/at-flight-authorization/internal/hapi"
	"172.21.5.249/airtrans/at-flight-authorization/internal/hapi/handlers/common"
	"172.21.5.249/airtrans/at-flight-authorization/internal/models"
	"172.21.5.249/airtrans/at-flight-authorization/internal/service"
	"172.21.5.249/airtrans/at-flight-authorization/internal/types"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
)

func AttachRoutes(s *hapi.Server) []*echo.Route {
	return []*echo.Route{
		registerCreate(s),
		registerGetByID(s),
		registerList(s),
		registerUpdate(s),
		registerDelete(s),
	}
}

func registerCreate(s *hapi.Server) *echo.Route {
	return s.Router.Root.POST("/flight-authorization-proposals", createHandler(s))
}

func registerGetByID(s *hapi.Server) *echo.Route {
	return s.Router.Root.GET("/flight-authorization-proposals/:id", getByIDHandler(s))
}

func registerList(s *hapi.Server) *echo.Route {
	return s.Router.Root.GET("/flight-authorization-proposals", listHandler(s))
}

func registerUpdate(s *hapi.Server) *echo.Route {
	return s.Router.Root.PUT("/flight-authorization-proposals/:id", updateHandler(s))
}

func registerDelete(s *hapi.Server) *echo.Route {
	return s.Router.Root.DELETE("/flight-authorization-proposals/:id", deleteHandler(s))
}

func createHandler(s *hapi.Server) echo.HandlerFunc {
	return func(c echo.Context) error {
		var payload models.FlightAuthorizationProposal
		if err := c.Bind(&payload); err != nil {
			return c.JSON(http.StatusBadRequest, types.ErrorResponse{Code: http.StatusBadRequest, Message: err.Error()})
		}
		created, err := s.Service.CreateFlightAuthorizationProposal(c.Request().Context(), &payload)
		if err != nil {
			return handleServiceError(c, err)
		}
		return c.JSON(http.StatusCreated, created)
	}
}

func getByIDHandler(s *hapi.Server) echo.HandlerFunc {
	return func(c echo.Context) error {
		id, err := uuid.Parse(c.Param("id"))
		if err != nil {
			return c.JSON(http.StatusBadRequest, types.ErrorResponse{Code: http.StatusBadRequest, Message: "invalid id"})
		}
		proposal, err := s.Service.GetFlightAuthorizationProposalByID(c.Request().Context(), id)
		if err != nil {
			return handleServiceError(c, err)
		}
		return c.JSON(http.StatusOK, proposal)
	}
}

func listHandler(s *hapi.Server) echo.HandlerFunc {
	return func(c echo.Context) error {
		page, size, err := common.ParsePageSize(c)
		if err != nil {
			return c.JSON(http.StatusBadRequest, types.ErrorResponse{Code: http.StatusBadRequest, Message: err.Error()})
		}
		opts := service.FlightAuthorizationProposalListOptions{
			FlightPurpose: c.QueryParam("flight_purpose"),
			Airport:       c.QueryParam("airport"),
			Page:          page,
			Size:          size,
		}
		proposals, err := s.Service.ListFlightAuthorizationProposals(c.Request().Context(), opts)
		if err != nil {
			return handleServiceError(c, err)
		}
		return c.JSON(http.StatusOK, proposals)
	}
}

func updateHandler(s *hapi.Server) echo.HandlerFunc {
	return func(c echo.Context) error {
		id, err := uuid.Parse(c.Param("id"))
		if err != nil {
			return c.JSON(http.StatusBadRequest, types.ErrorResponse{Code: http.StatusBadRequest, Message: "invalid id"})
		}
		var payload models.FlightAuthorizationProposal
		if err := c.Bind(&payload); err != nil {
			return c.JSON(http.StatusBadRequest, types.ErrorResponse{Code: http.StatusBadRequest, Message: err.Error()})
		}
		updated, err := s.Service.UpdateFlightAuthorizationProposal(c.Request().Context(), id, &payload)
		if err != nil {
			return handleServiceError(c, err)
		}
		return c.JSON(http.StatusOK, updated)
	}
}

func deleteHandler(s *hapi.Server) echo.HandlerFunc {
	return func(c echo.Context) error {
		id, err := uuid.Parse(c.Param("id"))
		if err != nil {
			return c.JSON(http.StatusBadRequest, types.ErrorResponse{Code: http.StatusBadRequest, Message: "invalid id"})
		}
		if err := s.Service.DeleteFlightAuthorizationProposal(c.Request().Context(), id); err != nil {
			return handleServiceError(c, err)
		}
		return c.NoContent(http.StatusNoContent)
	}
}

func handleServiceError(c echo.Context, err error) error {
	code := http.StatusInternalServerError
	if errors.Is(err, gorm.ErrRecordNotFound) {
		code = http.StatusNotFound
	}
	return c.JSON(code, types.ErrorResponse{Code: code, Message: err.Error()})
}
