package flightnotification

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
		s.Router.Root.POST("/flight-notifications", createHandler(s)),
		s.Router.Root.GET("/flight-notifications/:id", getHandler(s)),
		s.Router.Root.GET("/flight-notifications", listHandler(s)),
		s.Router.Root.PUT("/flight-notifications/:id", updateHandler(s)),
		s.Router.Root.DELETE("/flight-notifications/:id", deleteHandler(s)),
	}
}

// CreateFlightNotification godoc
//	@Summary		Create flight notification
//	@Description	Create a new flight notification associated with an approval
//	@Tags			flight-notifications
//	@Accept			json
//	@Produce		json
//	@Param			request	body		models.FlightNotification	true	"Flight notification payload"
//	@Success		201		{object}	models.FlightNotification
//	@Failure		400		{object}	types.ErrorResponse
//	@Failure		500		{object}	types.ErrorResponse
//	@Router			/flight-notifications [post]
func createHandler(s *hapi.Server) echo.HandlerFunc {
	return func(c echo.Context) error {
		payload := new(models.FlightNotification)
		if err := c.Bind(payload); err != nil {
			er := types.ErrorResponse{
				Message: "invalid request body",
			}
			return c.JSON(http.StatusBadRequest, er)
		}
		result, err := s.Service.CreateFlightNotification(c.Request().Context(), payload)
		if err != nil {
			log.Error().Err(err).Msg("CreateFlightNotification error")
			er := types.ErrorResponse{
				Message: s.I18n.Translate(err.Error(), language.Vietnamese),
			}
			return c.JSON(http.StatusInternalServerError, er)
		}
		return c.JSON(http.StatusCreated, result)
	}
}

// GetFlightNotification godoc
//	@Summary		Get flight notification
//	@Description	Retrieve a flight notification by ID
//	@Tags			flight-notifications
//	@Produce		json
//	@Param			id	path		string	true	"Flight notification ID"
//	@Success		200	{object}	models.FlightNotification
//	@Failure		400	{object}	types.ErrorResponse
//	@Failure		404	{object}	types.ErrorResponse
//	@Failure		500	{object}	types.ErrorResponse
//	@Router			/flight-notifications/{id} [get]
func getHandler(s *hapi.Server) echo.HandlerFunc {
	return func(c echo.Context) error {
		id, err := uuid.Parse(c.Param("id"))
		if err != nil {
			er := types.ErrorResponse{
				Message: "invalid id",
			}
			return c.JSON(http.StatusBadRequest, er)
		}
		result, err := s.Service.GetFlightNotification(c.Request().Context(), id)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				er := types.ErrorResponse{
					Message: "not found",
				}
				return c.JSON(http.StatusNotFound, er)
			}
			log.Error().Err(err).Msg("GetFlightNotification error")
			er := types.ErrorResponse{
				Message: s.I18n.Translate(err.Error(), language.Vietnamese),
			}
			return c.JSON(http.StatusInternalServerError, er)
		}
		return c.JSON(http.StatusOK, result)
	}
}

// ListFlightNotifications godoc
//	@Summary		List flight notifications
//	@Description	List flight notifications filtered by query parameters
//	@Tags			flight-notifications
//	@Produce		json
//	@Param			page								query		int		false	"Page number (starting from 1)"
//	@Param			size								query		int		false	"Page size"
//	@Param			flight_authorization_approval_id	query		string	false	"Filter by approval ID"
//	@Success		200									{object}	map[string]interface{}
//	@Failure		400									{object}	types.ErrorResponse
//	@Failure		500									{object}	types.ErrorResponse
//	@Router			/flight-notifications [get]
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
		items, total, err := s.Service.ListFlightNotifications(c.Request().Context(), filters, page, size)
		if err != nil {
			log.Error().Err(err).Msg("ListFlightNotifications error")
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

// UpdateFlightNotification godoc
//	@Summary		Update flight notification
//	@Description	Update a flight notification and its intended flight areas
//	@Tags			flight-notifications
//	@Accept			json
//	@Produce		json
//	@Param			id		path		string						true	"Flight notification ID"
//	@Param			request	body		models.FlightNotification	true	"Flight notification payload"
//	@Success		200		{object}	models.FlightNotification
//	@Failure		400		{object}	types.ErrorResponse
//	@Failure		404		{object}	types.ErrorResponse
//	@Failure		500		{object}	types.ErrorResponse
//	@Router			/flight-notifications/{id} [put]
func updateHandler(s *hapi.Server) echo.HandlerFunc {
	return func(c echo.Context) error {
		id, err := uuid.Parse(c.Param("id"))
		if err != nil {
			er := types.ErrorResponse{
				Message: "invalid id",
			}
			return c.JSON(http.StatusBadRequest, er)
		}
		payload := new(models.FlightNotification)
		if err := c.Bind(payload); err != nil {
			er := types.ErrorResponse{
				Message: "invalid request body",
			}
			return c.JSON(http.StatusBadRequest, er)
		}
		result, err := s.Service.UpdateFlightNotification(c.Request().Context(), id, payload)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				er := types.ErrorResponse{
					Message: "not found",
				}
				return c.JSON(http.StatusNotFound, er)
			}
			log.Error().Err(err).Msg("UpdateFlightNotification error")
			er := types.ErrorResponse{
				Message: s.I18n.Translate(err.Error(), language.Vietnamese),
			}
			return c.JSON(http.StatusInternalServerError, er)
		}
		return c.JSON(http.StatusOK, result)
	}
}

// DeleteFlightNotification godoc
//	@Summary		Delete flight notification
//	@Description	Delete a flight notification by ID
//	@Tags			flight-notifications
//	@Produce		json
//	@Param			id	path		string	true	"Flight notification ID"
//	@Success		200	{object}	types.SucceedResponse
//	@Failure		400	{object}	types.ErrorResponse
//	@Failure		404	{object}	types.ErrorResponse
//	@Failure		500	{object}	types.ErrorResponse
//	@Router			/flight-notifications/{id} [delete]
func deleteHandler(s *hapi.Server) echo.HandlerFunc {
	return func(c echo.Context) error {
		id, err := uuid.Parse(c.Param("id"))
		if err != nil {
			er := types.ErrorResponse{
				Message: "invalid id",
			}
			return c.JSON(http.StatusBadRequest, er)
		}
		err = s.Service.DeleteFlightNotification(c.Request().Context(), id)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				er := types.ErrorResponse{
					Message: "not found",
				}
				return c.JSON(http.StatusNotFound, er)
			}
			log.Error().Err(err).Msg("DeleteFlightNotification error")
			er := types.ErrorResponse{
				Message: s.I18n.Translate(err.Error(), language.Vietnamese),
			}
			return c.JSON(http.StatusInternalServerError, er)
		}
		return c.JSON(http.StatusOK, types.SucceedResponse{Success: true})
	}
}
