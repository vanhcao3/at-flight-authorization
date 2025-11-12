package models

import (
	"database/sql/driver"
	"fmt"
	"strconv"
	"strings"
	"time"
)

var dateLayouts = []string{
	time.RFC3339,
	"2006-01-02",
}

type Date struct {
	time.Time
}

func (d *Date) UnmarshalJSON(data []byte) error {
	str := strings.TrimSpace(string(data))
	if str == "null" || str == "" {
		d.Time = time.Time{}
		return nil
	}
	s, err := strconv.Unquote(str)
	if err != nil {
		return err
	}
	if s == "" {
		d.Time = time.Time{}
		return nil
	}
	t, err := parseDate(s)
	if err != nil {
		return err
	}
	d.Time = t
	return nil
}

func (d Date) MarshalJSON() ([]byte, error) {
	if d.Time.IsZero() {
		return []byte("null"), nil
	}
	return []byte(`"` + d.Time.Format("2006-01-02") + `"`), nil
}

func (d Date) Value() (driver.Value, error) {
	if d.Time.IsZero() {
		return nil, nil
	}
	return d.Time.Format("2006-01-02"), nil
}

func (d *Date) Scan(value interface{}) error {
	if value == nil {
		d.Time = time.Time{}
		return nil
	}
	switch v := value.(type) {
	case time.Time:
		d.Time = v
		return nil
	case string:
		return d.assignFromString(v)
	case []byte:
		return d.assignFromString(string(v))
	default:
		return fmt.Errorf("cannot scan %T into Date", value)
	}
}

func (d *Date) assignFromString(value string) error {
	value = strings.TrimSpace(value)
	if value == "" {
		d.Time = time.Time{}
		return nil
	}
	t, err := parseDate(value)
	if err != nil {
		return err
	}
	d.Time = t
	return nil
}

func parseDate(value string) (time.Time, error) {
	for _, layout := range dateLayouts {
		if t, err := time.Parse(layout, value); err == nil {
			return t, nil
		}
	}
	return time.Time{}, fmt.Errorf("invalid date %q", value)
}
