package flightauthorizationapproval

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
	return s.Router.Root.POST("/flight-authorization-approvals", createHandler(s))
}

func registerGetByID(s *hapi.Server) *echo.Route {
	return s.Router.Root.GET("/flight-authorization-approvals/:id", getByIDHandler(s))
}

func registerList(s *hapi.Server) *echo.Route {
	return s.Router.Root.GET("/flight-authorization-approvals", listHandler(s))
}

func registerUpdate(s *hapi.Server) *echo.Route {
	return s.Router.Root.PUT("/flight-authorization-approvals/:id", updateHandler(s))
}

func registerDelete(s *hapi.Server) *echo.Route {
	return s.Router.Root.DELETE("/flight-authorization-approvals/:id", deleteHandler(s))
}

func createHandler(s *hapi.Server) echo.HandlerFunc {
	return func(c echo.Context) error {
		var payload models.FlightAuthorizationApproval
		if err := c.Bind(&payload); err != nil {
			return c.JSON(http.StatusBadRequest, types.ErrorResponse{Code: http.StatusBadRequest, Message: err.Error()})
		}
		created, err := s.Service.CreateFlightAuthorizationApproval(c.Request().Context(), &payload)
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
		approval, err := s.Service.GetFlightAuthorizationApprovalByID(c.Request().Context(), id)
		if err != nil {
			return handleServiceError(c, err)
		}
		return c.JSON(http.StatusOK, approval)
	}
}

func listHandler(s *hapi.Server) echo.HandlerFunc {
	return func(c echo.Context) error {
		page, size, err := common.ParsePageSize(c)
		if err != nil {
			return c.JSON(http.StatusBadRequest, types.ErrorResponse{Code: http.StatusBadRequest, Message: err.Error()})
		}
		var proposalID uuid.UUID
		if v := c.QueryParam("flight_authorization_proposal_id"); v != "" {
			proposalID, err = uuid.Parse(v)
			if err != nil {
				return c.JSON(http.StatusBadRequest, types.ErrorResponse{Code: http.StatusBadRequest, Message: "invalid flight_authorization_proposal_id"})
			}
		}
		opts := service.FlightAuthorizationApprovalListOptions{
			FlightAuthorizationProposalID: proposalID,
			Page:                          page,
			Size:                          size,
		}
		approvals, err := s.Service.ListFlightAuthorizationApprovals(c.Request().Context(), opts)
		if err != nil {
			return handleServiceError(c, err)
		}
		return c.JSON(http.StatusOK, approvals)
	}
}

func updateHandler(s *hapi.Server) echo.HandlerFunc {
	return func(c echo.Context) error {
		id, err := uuid.Parse(c.Param("id"))
		if err != nil {
			return c.JSON(http.StatusBadRequest, types.ErrorResponse{Code: http.StatusBadRequest, Message: "invalid id"})
		}
		var payload models.FlightAuthorizationApproval
		if err := c.Bind(&payload); err != nil {
			return c.JSON(http.StatusBadRequest, types.ErrorResponse{Code: http.StatusBadRequest, Message: err.Error()})
		}
		updated, err := s.Service.UpdateFlightAuthorizationApproval(c.Request().Context(), id, &payload)
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
		if err := s.Service.DeleteFlightAuthorizationApproval(c.Request().Context(), id); err != nil {
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
