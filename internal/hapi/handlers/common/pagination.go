package common

import (
	"fmt"
	"strconv"

	"github.com/labstack/echo/v4"
)

func ParsePageSize(c echo.Context) (int, int, error) {
	page := 0
	size := 0
	if v := c.QueryParam("page"); v != "" {
		p, err := strconv.Atoi(v)
		if err != nil || p < 1 {
			return 0, 0, fmt.Errorf("invalid page")
		}
		page = p
	}
	if v := c.QueryParam("size"); v != "" {
		s, err := strconv.Atoi(v)
		if err != nil || s < 1 {
			return 0, 0, fmt.Errorf("invalid size")
		}
		size = s
	}
	return page, size, nil
}
