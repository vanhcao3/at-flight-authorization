package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
)

type TakeOffAndLandingArea struct {
	Place     string  `json:"place" gorm:"not null"`
	Commune   string  `json:"commune" gorm:"not null"`
	Province  string  `json:"province" gorm:"not null"`
	Latitude  float64 `json:"latitude" gorm:"not null"`
	Longitude float64 `json:"longitude" gorm:"not null"`
}

type FlightParameter struct {
	ID                            uuid.UUID             `json:"id" gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	FlightAuthorizationApprovalID uuid.UUID             `json:"flight_authorization_approval_id" gorm:"type:uuid;index"`
	Altitude                      string                `json:"altitude" gorm:"not null"`
	Radius                        string                `json:"radius" gorm:"not null"`
	TakeOffAndLandingArea         TakeOffAndLandingArea `json:"take_off_and_landing_area" gorm:"embedded"`
}

type AuthorizedFlightAreaCoordinate struct {
	ID                     uuid.UUID `json:"id" gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	AuthorizedFlightAreaID uuid.UUID `json:"authorized_flight_area_id" gorm:"type:uuid;index"`
	Latitude               float64   `json:"latitude" gorm:"not null"`
	Longitude              float64   `json:"longitude" gorm:"not null"`
}

// 5. Kich thuoc vung troi to chuc bay
type AuthorizedFlightArea struct {
	ID                            uuid.UUID                        `json:"id" gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	FlightAuthorizationApprovalID uuid.UUID                        `json:"flight_authorization_approval_id" gorm:"type:uuid;index"`
	Place                         string                           `json:"place" gorm:"not null"`
	Commune                       string                           `json:"commune" gorm:"not null"`
	Province                      string                           `json:"province" gorm:"not null"`
	Polygon                       []AuthorizedFlightAreaCoordinate `json:"polygon" gorm:"foreignKey:AuthorizedFlightAreaID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	Altitude                      string                           `json:"altitude" gorm:"not null"`
}

type FlightAuthorizationApproval struct {
	ID                            uuid.UUID                   `json:"id" gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	FlightAuthorizationProposalID uuid.UUID                   `json:"flight_authorization_proposal_id" gorm:"type:uuid;index"`
	FlightAuthorizationProposal   FlightAuthorizationProposal `json:"flight_authorization_proposal" gorm:"foreignKey:FlightAuthorizationProposalID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	AuthorizedFlightArea          []AuthorizedFlightArea      `json:"authorized_flight_area" gorm:"foreignKey:FlightAuthorizationApprovalID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	FlightParameter               FlightParameter             `json:"flight_parameter" gorm:"foreignKey:FlightAuthorizationApprovalID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	AuthorizedOperatingDuration   OperatingDuration           `json:"authorized_operating_duration" gorm:"embedded"`
	FlightNegotiationAuthorities  datatypes.JSONSlice[string] `json:"flight_negotiation_authorities" gorm:"type:json"`
	CreatedAt                     time.Time                   `json:"created_at"`
	UpdatedAt                     time.Time                   `json:"updated_at"`
}
