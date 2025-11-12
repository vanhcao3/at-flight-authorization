package flightauthorizationapproval

import (
	"errors"
	"net/http"
	"strconv"

	"172.21.5.249/airtrans/at-flight-authorization/internal/hapi"
	"172.21.5.249/airtrans/at-flight-authorization/internal/models"
	"172.21.5.249/airtrans/at-flight-authorization/internal/types"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog/log"
	"golang.org/x/text/language"
	"gorm.io/gorm"
)

func RegisterRoutes(s *hapi.Server) []*echo.Route {
	return []*echo.Route{
		s.Router.Root.POST("/flight-authorization-approvals", createHandler(s)),
		s.Router.Root.GET("/flight-authorization-approvals/:id", getHandler(s)),
		s.Router.Root.GET("/flight-authorization-approvals", listHandler(s)),
		s.Router.Root.PUT("/flight-authorization-approvals/:id", updateHandler(s)),
		s.Router.Root.DELETE("/flight-authorization-approvals/:id", deleteHandler(s)),
	}
}

// CreateFlightAuthorizationApproval godoc
// @Summary Create flight authorization approval
// @Description Create a new flight authorization approval with associated areas and parameters
// @Tags flight-authorization-approvals
// @Accept json
// @Produce json
// @Param request body models.FlightAuthorizationApproval true "Flight authorization approval payload"
// @Success 201 {object} models.FlightAuthorizationApproval
// @Failure 400 {object} types.ErrorResponse
// @Failure 500 {object} types.ErrorResponse
// @Router /flight-authorization-approvals [post]
func createHandler(s *hapi.Server) echo.HandlerFunc {
	return func(c echo.Context) error {
		payload := new(models.FlightAuthorizationApproval)
		if err := c.Bind(payload); err != nil {
			er := types.ErrorResponse{
				Message: "invalid request body",
			}
			return c.JSON(http.StatusBadRequest, er)
		}
		result, err := s.Service.CreateFlightAuthorizationApproval(c.Request().Context(), payload)
		if err != nil {
			log.Error().Err(err).Msg("CreateFlightAuthorizationApproval error")
			er := types.ErrorResponse{
				Message: s.I18n.Translate(err.Error(), language.Vietnamese),
			}
			return c.JSON(http.StatusInternalServerError, er)
		}
		return c.JSON(http.StatusCreated, result)
	}
}

// GetFlightAuthorizationApproval godoc
// @Summary Get flight authorization approval
// @Description Retrieve a flight authorization approval by ID
// @Tags flight-authorization-approvals
// @Produce json
// @Param id path string true "Flight authorization approval ID"
// @Success 200 {object} models.FlightAuthorizationApproval
// @Failure 400 {object} types.ErrorResponse
// @Failure 404 {object} types.ErrorResponse
// @Failure 500 {object} types.ErrorResponse
// @Router /flight-authorization-approvals/{id} [get]
func getHandler(s *hapi.Server) echo.HandlerFunc {
	return func(c echo.Context) error {
		id, err := uuid.Parse(c.Param("id"))
		if err != nil {
			er := types.ErrorResponse{
				Message: "invalid id",
			}
			return c.JSON(http.StatusBadRequest, er)
		}
		result, err := s.Service.GetFlightAuthorizationApproval(c.Request().Context(), id)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				er := types.ErrorResponse{
					Message: "not found",
				}
				return c.JSON(http.StatusNotFound, er)
			}
			log.Error().Err(err).Msg("GetFlightAuthorizationApproval error")
			er := types.ErrorResponse{
				Message: s.I18n.Translate(err.Error(), language.Vietnamese),
			}
			return c.JSON(http.StatusInternalServerError, er)
		}
		return c.JSON(http.StatusOK, result)
	}
}

// ListFlightAuthorizationApprovals godoc
// @Summary List flight authorization approvals
// @Description List flight authorization approvals filtered by query parameters
// @Tags flight-authorization-approvals
// @Produce json
// @Param page query int false "Page number (starting from 1)"
// @Param size query int false "Page size"
// @Param flight_authorization_proposal_id query string false "Filter by proposal ID"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} types.ErrorResponse
// @Failure 500 {object} types.ErrorResponse
// @Router /flight-authorization-approvals [get]
func listHandler(s *hapi.Server) echo.HandlerFunc {
	return func(c echo.Context) error {
		page := 1
		if v := c.QueryParam("page"); v != "" {
			n, err := strconv.Atoi(v)
			if err != nil || n <= 0 {
				er := types.ErrorResponse{
					Message: "invalid page",
				}
				return c.JSON(http.StatusBadRequest, er)
			}
			page = n
		}
		size := 0
		if v := c.QueryParam("size"); v != "" {
			n, err := strconv.Atoi(v)
			if err != nil || n < 0 {
				er := types.ErrorResponse{
					Message: "invalid size",
				}
				return c.JSON(http.StatusBadRequest, er)
			}
			size = n
		}
		filters := map[string][]string{}
		for key, values := range c.QueryParams() {
			if key == "page" || key == "size" {
				continue
			}
			if len(values) == 0 {
				continue
			}
			filters[key] = values
		}
		items, total, err := s.Service.ListFlightAuthorizationApprovals(c.Request().Context(), filters, page, size)
		if err != nil {
			log.Error().Err(err).Msg("ListFlightAuthorizationApprovals error")
			er := types.ErrorResponse{
				Message: s.I18n.Translate(err.Error(), language.Vietnamese),
			}
			return c.JSON(http.StatusInternalServerError, er)
		}
		c.Response().Header().Set("X-Total-Count", strconv.FormatInt(total, 10))
		resp := map[string]interface{}{
			"data":  items,
			"page":  page,
			"size":  size,
			"total": total,
		}
		return c.JSON(http.StatusOK, resp)
	}
}

// UpdateFlightAuthorizationApproval godoc
// @Summary Update flight authorization approval
// @Description Update a flight authorization approval and its related resources
// @Tags flight-authorization-approvals
// @Accept json
// @Produce json
// @Param id path string true "Flight authorization approval ID"
// @Param request body models.FlightAuthorizationApproval true "Flight authorization approval payload"
// @Success 200 {object} models.FlightAuthorizationApproval
// @Failure 400 {object} types.ErrorResponse
// @Failure 404 {object} types.ErrorResponse
// @Failure 500 {object} types.ErrorResponse
// @Router /flight-authorization-approvals/{id} [put]
func updateHandler(s *hapi.Server) echo.HandlerFunc {
	return func(c echo.Context) error {
		id, err := uuid.Parse(c.Param("id"))
		if err != nil {
			er := types.ErrorResponse{
				Message: "invalid id",
			}
			return c.JSON(http.StatusBadRequest, er)
		}
		payload := new(models.FlightAuthorizationApproval)
		if err := c.Bind(payload); err != nil {
			er := types.ErrorResponse{
				Message: "invalid request body",
			}
			return c.JSON(http.StatusBadRequest, er)
		}
		result, err := s.Service.UpdateFlightAuthorizationApproval(c.Request().Context(), id, payload)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				er := types.ErrorResponse{
					Message: "not found",
				}
				return c.JSON(http.StatusNotFound, er)
			}
			log.Error().Err(err).Msg("UpdateFlightAuthorizationApproval error")
			er := types.ErrorResponse{
				Message: s.I18n.Translate(err.Error(), language.Vietnamese),
			}
			return c.JSON(http.StatusInternalServerError, er)
		}
		return c.JSON(http.StatusOK, result)
	}
}

// DeleteFlightAuthorizationApproval godoc
// @Summary Delete flight authorization approval
// @Description Delete a flight authorization approval by ID
// @Tags flight-authorization-approvals
// @Produce json
// @Param id path string true "Flight authorization approval ID"
// @Success 200 {object} types.SucceedResponse
// @Failure 400 {object} types.ErrorResponse
// @Failure 404 {object} types.ErrorResponse
// @Failure 500 {object} types.ErrorResponse
// @Router /flight-authorization-approvals/{id} [delete]
func deleteHandler(s *hapi.Server) echo.HandlerFunc {
	return func(c echo.Context) error {
		id, err := uuid.Parse(c.Param("id"))
		if err != nil {
			er := types.ErrorResponse{
				Message: "invalid id",
			}
			return c.JSON(http.StatusBadRequest, er)
		}
		err = s.Service.DeleteFlightAuthorizationApproval(c.Request().Context(), id)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				er := types.ErrorResponse{
					Message: "not found",
				}
				return c.JSON(http.StatusNotFound, er)
			}
			log.Error().Err(err).Msg("DeleteFlightAuthorizationApproval error")
			er := types.ErrorResponse{
				Message: s.I18n.Translate(err.Error(), language.Vietnamese),
			}
			return c.JSON(http.StatusInternalServerError, er)
		}
		return c.JSON(http.StatusOK, types.SucceedResponse{Success: true})
	}
}
