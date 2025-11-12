package models

import "github.com/google/uuid"

type IntendedFlightAreaCoordinate struct {
	ID                   uuid.UUID `json:"id" gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	IntendedFlightAreaID uuid.UUID `json:"intended_flight_area_id" gorm:"type:uuid;index"`
	Latitude             float64   `json:"latitude" gorm:"not null"`
	Longitude            float64   `json:"longitude" gorm:"not null"`
}

// 5. Kich thuoc vung troi to chuc bay
type IntendedFlightArea struct {
	ID                   uuid.UUID                      `json:"id" gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	FlightNotificationID uuid.UUID                      `json:"flight_notification_id" gorm:"type:uuid;index"`
	Place                string                         `json:"place" gorm:"not null"`
	Commune              string                         `json:"commune" gorm:"not null"`
	Province             string                         `json:"province" gorm:"not null"`
	Polygon              []IntendedFlightAreaCoordinate `json:"polygon" gorm:"foreignKey:IntendedFlightAreaID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	Altitude             string                         `json:"altitude" gorm:"not null"`
}

type FlightNotification struct {
	ID                            uuid.UUID                   `json:"id" gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	Name                          string                      `json:"name" gorm:"not null"`
	FlightAuthorizationApprovalID uuid.UUID                   `json:"flight_authorization_approval_id" gorm:"type:uuid;index"`
	FlightAuthorizationApproval   FlightAuthorizationApproval `json:"flight_authorization_approval" gorm:"foreignKey:FlightAuthorizationApprovalID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	IntendedOperatingDuration     OperatingDuration           `json:"intended_operating_duration" gorm:"embedded"`
	IntendedFlightArea            []IntendedFlightArea        `json:"intended_flight_area" gorm:"foreignKey:FlightNotificationID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
}
