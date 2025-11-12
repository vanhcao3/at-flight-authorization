package flightnotification

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
	return s.Router.Root.POST("/flight-notifications", createHandler(s))
}

func registerGetByID(s *hapi.Server) *echo.Route {
	return s.Router.Root.GET("/flight-notifications/:id", getByIDHandler(s))
}

func registerList(s *hapi.Server) *echo.Route {
	return s.Router.Root.GET("/flight-notifications", listHandler(s))
}

func registerUpdate(s *hapi.Server) *echo.Route {
	return s.Router.Root.PUT("/flight-notifications/:id", updateHandler(s))
}

func registerDelete(s *hapi.Server) *echo.Route {
	return s.Router.Root.DELETE("/flight-notifications/:id", deleteHandler(s))
}

func createHandler(s *hapi.Server) echo.HandlerFunc {
	return func(c echo.Context) error {
		var payload models.FlightNotification
		if err := c.Bind(&payload); err != nil {
			return c.JSON(http.StatusBadRequest, types.ErrorResponse{Code: http.StatusBadRequest, Message: err.Error()})
		}
		created, err := s.Service.CreateFlightNotification(c.Request().Context(), &payload)
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
		notification, err := s.Service.GetFlightNotificationByID(c.Request().Context(), id)
		if err != nil {
			return handleServiceError(c, err)
		}
		return c.JSON(http.StatusOK, notification)
	}
}

func listHandler(s *hapi.Server) echo.HandlerFunc {
	return func(c echo.Context) error {
		page, size, err := common.ParsePageSize(c)
		if err != nil {
			return c.JSON(http.StatusBadRequest, types.ErrorResponse{Code: http.StatusBadRequest, Message: err.Error()})
		}
		var approvalID uuid.UUID
		if v := c.QueryParam("flight_authorization_approval_id"); v != "" {
			approvalID, err = uuid.Parse(v)
			if err != nil {
				return c.JSON(http.StatusBadRequest, types.ErrorResponse{Code: http.StatusBadRequest, Message: "invalid flight_authorization_approval_id"})
			}
		}
		opts := service.FlightNotificationListOptions{
			FlightAuthorizationApprovalID: approvalID,
			Page:                          page,
			Size:                          size,
		}
		notifications, err := s.Service.ListFlightNotifications(c.Request().Context(), opts)
		if err != nil {
			return handleServiceError(c, err)
		}
		return c.JSON(http.StatusOK, notifications)
	}
}

func updateHandler(s *hapi.Server) echo.HandlerFunc {
	return func(c echo.Context) error {
		id, err := uuid.Parse(c.Param("id"))
		if err != nil {
			return c.JSON(http.StatusBadRequest, types.ErrorResponse{Code: http.StatusBadRequest, Message: "invalid id"})
		}
		var payload models.FlightNotification
		if err := c.Bind(&payload); err != nil {
			return c.JSON(http.StatusBadRequest, types.ErrorResponse{Code: http.StatusBadRequest, Message: err.Error()})
		}
		updated, err := s.Service.UpdateFlightNotification(c.Request().Context(), id, &payload)
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
		if err := s.Service.DeleteFlightNotification(c.Request().Context(), id); err != nil {
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
