package flightauthorizationproposal

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
		s.Router.Root.POST("/flight-authorization-proposals", createHandler(s)),
		s.Router.Root.GET("/flight-authorization-proposals/:id", getHandler(s)),
		s.Router.Root.GET("/flight-authorization-proposals", listHandler(s)),
		s.Router.Root.PUT("/flight-authorization-proposals/:id", updateHandler(s)),
		s.Router.Root.DELETE("/flight-authorization-proposals/:id", deleteHandler(s)),
	}
}

// CreateFlightAuthorizationProposal godoc
// @Summary Create flight authorization proposal
// @Description Create a new flight authorization proposal with nested resources
// @Tags flight-authorization-proposals
// @Accept json
// @Produce json
// @Param request body models.FlightAuthorizationProposal true "Flight authorization proposal payload"
// @Success 201 {object} models.FlightAuthorizationProposal
// @Failure 400 {object} types.ErrorResponse
// @Failure 500 {object} types.ErrorResponse
// @Router /flight-authorization-proposals [post]
func createHandler(s *hapi.Server) echo.HandlerFunc {
	return func(c echo.Context) error {
		payload := new(models.FlightAuthorizationProposal)
		if err := c.Bind(payload); err != nil {
			er := types.ErrorResponse{
				Message: "invalid request body",
			}
			return c.JSON(http.StatusBadRequest, er)
		}
		result, err := s.Service.CreateFlightAuthorizationProposal(c.Request().Context(), payload)
		if err != nil {
			log.Error().Err(err).Msg("CreateFlightAuthorizationProposal error")
			er := types.ErrorResponse{
				Message: s.I18n.Translate(err.Error(), language.Vietnamese),
			}
			return c.JSON(http.StatusInternalServerError, er)
		}
		return c.JSON(http.StatusCreated, result)
	}
}

// GetFlightAuthorizationProposal godoc
// @Summary Get flight authorization proposal
// @Description Retrieve a flight authorization proposal by ID
// @Tags flight-authorization-proposals
// @Produce json
// @Param id path string true "Flight authorization proposal ID"
// @Success 200 {object} models.FlightAuthorizationProposal
// @Failure 400 {object} types.ErrorResponse
// @Failure 404 {object} types.ErrorResponse
// @Failure 500 {object} types.ErrorResponse
// @Router /flight-authorization-proposals/{id} [get]
func getHandler(s *hapi.Server) echo.HandlerFunc {
	return func(c echo.Context) error {
		id, err := uuid.Parse(c.Param("id"))
		if err != nil {
			er := types.ErrorResponse{
				Message: "invalid id",
			}
			return c.JSON(http.StatusBadRequest, er)
		}
		result, err := s.Service.GetFlightAuthorizationProposal(c.Request().Context(), id)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				er := types.ErrorResponse{
					Message: "not found",
				}
				return c.JSON(http.StatusNotFound, er)
			}
			log.Error().Err(err).Msg("GetFlightAuthorizationProposal error")
			er := types.ErrorResponse{
				Message: s.I18n.Translate(err.Error(), language.Vietnamese),
			}
			return c.JSON(http.StatusInternalServerError, er)
		}
		return c.JSON(http.StatusOK, result)
	}
}

// ListFlightAuthorizationProposals godoc
// @Summary List flight authorization proposals
// @Description List flight authorization proposals filtered by query parameters
// @Tags flight-authorization-proposals
// @Produce json
// @Param page query int false "Page number (starting from 1)"
// @Param size query int false "Page size"
// @Param flight_purpose query string false "Filter by flight purpose"
// @Param airport query string false "Filter by airport"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} types.ErrorResponse
// @Failure 500 {object} types.ErrorResponse
// @Router /flight-authorization-proposals [get]
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
		items, total, err := s.Service.ListFlightAuthorizationProposals(c.Request().Context(), filters, page, size)
		if err != nil {
			log.Error().Err(err).Msg("ListFlightAuthorizationProposals error")
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

// UpdateFlightAuthorizationProposal godoc
// @Summary Update flight authorization proposal
// @Description Update a flight authorization proposal and its nested resources
// @Tags flight-authorization-proposals
// @Accept json
// @Produce json
// @Param id path string true "Flight authorization proposal ID"
// @Param request body models.FlightAuthorizationProposal true "Flight authorization proposal payload"
// @Success 200 {object} models.FlightAuthorizationProposal
// @Failure 400 {object} types.ErrorResponse
// @Failure 404 {object} types.ErrorResponse
// @Failure 500 {object} types.ErrorResponse
// @Router /flight-authorization-proposals/{id} [put]
func updateHandler(s *hapi.Server) echo.HandlerFunc {
	return func(c echo.Context) error {
		id, err := uuid.Parse(c.Param("id"))
		if err != nil {
			er := types.ErrorResponse{
				Message: "invalid id",
			}
			return c.JSON(http.StatusBadRequest, er)
		}
		payload := new(models.FlightAuthorizationProposal)
		if err := c.Bind(payload); err != nil {
			er := types.ErrorResponse{
				Message: "invalid request body",
			}
			return c.JSON(http.StatusBadRequest, er)
		}
		result, err := s.Service.UpdateFlightAuthorizationProposal(c.Request().Context(), id, payload)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				er := types.ErrorResponse{
					Message: "not found",
				}
				return c.JSON(http.StatusNotFound, er)
			}
			log.Error().Err(err).Msg("UpdateFlightAuthorizationProposal error")
			er := types.ErrorResponse{
				Message: s.I18n.Translate(err.Error(), language.Vietnamese),
			}
			return c.JSON(http.StatusInternalServerError, er)
		}
		return c.JSON(http.StatusOK, result)
	}
}

// DeleteFlightAuthorizationProposal godoc
// @Summary Delete flight authorization proposal
// @Description Delete a flight authorization proposal by ID
// @Tags flight-authorization-proposals
// @Produce json
// @Param id path string true "Flight authorization proposal ID"
// @Success 200 {object} types.SucceedResponse
// @Failure 400 {object} types.ErrorResponse
// @Failure 404 {object} types.ErrorResponse
// @Failure 500 {object} types.ErrorResponse
// @Router /flight-authorization-proposals/{id} [delete]
func deleteHandler(s *hapi.Server) echo.HandlerFunc {
	return func(c echo.Context) error {
		id, err := uuid.Parse(c.Param("id"))
		if err != nil {
			er := types.ErrorResponse{
				Message: "invalid id",
			}
			return c.JSON(http.StatusBadRequest, er)
		}
		err = s.Service.DeleteFlightAuthorizationProposal(c.Request().Context(), id)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				er := types.ErrorResponse{
					Message: "not found",
				}
				return c.JSON(http.StatusNotFound, er)
			}
			log.Error().Err(err).Msg("DeleteFlightAuthorizationProposal error")
			er := types.ErrorResponse{
				Message: s.I18n.Translate(err.Error(), language.Vietnamese),
			}
			return c.JSON(http.StatusInternalServerError, er)
		}
		return c.JSON(http.StatusOK, types.SucceedResponse{Success: true})
	}
}
