package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
)

type AuthorizationLog struct {
	ExampleId uuid.UUID `json:"id" gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	Result    string    `json:"result" gorm:"not null"`
	UserId    string    `json:"user_id" gorm:"not null"`
	Username  string    `json:"username" gorm:"not null"`
	Resource  string    `json:"resource" gorm:"not null"`
	Action    string    `json:"action" gorm:"not null"`
	Timestamp int64     `json:"timestamp" gorm:"not null"`
	Client    string    `json:"client" gorm:"not null"`
	URI       string    `json:"uri" gorm:"not null"`
}
type Example struct {
	Id               uuid.UUID        `json:"id" gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	AuthorizationLog AuthorizationLog `json:"something" gorm:"foreignKey:ExampleId;constraint:OnUpdate:CASCADE,OnDelete:SET NULL;"`
}

// 1. Thong tin to chuc de nghi cap phep bay
type Operator struct {
	ID                            uuid.UUID `json:"id" gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	FlightAuthorizationProposalID uuid.UUID `json:"flight_authorization_proposal_id" gorm:"type:uuid;index"`
	BusinessName                  string    `json:"business_name" gorm:"not null"`
	TaxIdentificationNumber       string    `json:"tax_identification_number" gorm:"not null"`
	Address                       string    `json:"address" gorm:"not null"`
	Nationality                   string    `json:"nationality" gorm:"not null"`
	PhoneNumber                   string    `json:"phone_number" gorm:"not null"`
	Fax                           string    `json:"fax"`
	Email                         string    `json:"email"`
}

type Dimension struct {
	Length int `json:"length" gorm:"not null"`
	Width  int `json:"width" gorm:"not null"`
	Height int `json:"height" gorm:"not null"`
}

// 2. Giay chung nhan dang ky
type DroneRegistration struct {
	DroneType          string    `json:"drone_type" gorm:"not null"`
	FactoryNumber      string    `json:"factory_number" gorm:"not null"`
	RegistrationNumber string    `json:"registration_number" gorm:"not null"`
	RegistrationDate   time.Time `json:"registration_date" gorm:"type:date;not null"`
}

// 3. Thong so ky thuat cua phuong tien bay
type DroneSpecification struct {
	MaximumTakeOffWeight int       `json:"maximum_take_off_weight" gorm:"not null"`
	Dimension            Dimension `json:"dimension" gorm:"embedded"`
	EngineType           string    `json:"engine_type"`
	OperatingFrequency   string    `json:"operating_frequency"`
	OperatingMethod      string    `json:"operating_method" gorm:"not null"`
}

type Drone struct {
	ID                            uuid.UUID          `json:"id" gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	FlightAuthorizationProposalID uuid.UUID          `json:"flight_authorization_proposal_id" gorm:"type:uuid;index"`
	DroneSpecification            DroneSpecification `json:"drone_specification" gorm:"embedded"`
	DroneRegistration             DroneRegistration  `json:"drone_registration" gorm:"embedded"`
}

type FlightAreaCoordinate struct {
	ID           uuid.UUID `json:"id" gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	FlightAreaID uuid.UUID `json:"flight_area_id" gorm:"type:uuid;index"`
	Latitude     float64   `json:"latitude" gorm:"not null"`
	Longitude    float64   `json:"longitude" gorm:"not null"`
}

// 5. Kich thuoc vung troi to chuc bay
type FlightArea struct {
	ID                            uuid.UUID              `json:"id" gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	FlightAuthorizationProposalID uuid.UUID              `json:"flight_authorization_proposal_id" gorm:"type:uuid;index"`
	Place                         string                 `json:"place" gorm:"not null"`
	Commune                       string                 `json:"commune" gorm:"not null"`
	Province                      string                 `json:"province" gorm:"not null"`
	Polygon                       []FlightAreaCoordinate `json:"polygon" gorm:"foreignKey:FlightAreaID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	Altitude                      string                 `json:"altitude" gorm:"not null"`
}

// 6. So ngay va thoi gian to chuc bay (gio, ngay, thang, nam)
type OperatingDuration struct {
	Duration int       `json:"duration" gorm:"not nul"`  //So ngay to chuc bay
	FromDay  time.Time `json:"from_day" gorm:"not null"` //Ngay bat dau (Ngay Thang Nam gio)
	ToDay    time.Time `json:"to_day" gorm:"not null"`   //Ngay ket thuc (Ngay Thang Nam gio)
}

// 8. Thong tin nguoi dieu khien phuong tien bay
type PilotLicense struct {
	LicenseNumber        string    `json:"license_number" gorm:"not null"`
	LicenseProvisionDate time.Time `json:"license_provision_date" gorm:"type:date;not null"`
}

type Pilot struct {
	ID                            uuid.UUID    `json:"id" gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	FlightAuthorizationProposalID uuid.UUID    `json:"flight_authorization_proposal_id" gorm:"type:uuid;index"`
	Name                          string       `json:"name" gorm:"not null"`
	Birthday                      time.Time    `json:"birthday" gorm:"type:date;not null"`
	IdentificationNumber          string       `json:"identification_number" gorm:"not null"`
	PhoneNumber                   string       `json:"phone_number" gorm:"not null"`
	PilotLicense                  PilotLicense `json:"pilot_license" gorm:"embedded"`
}

type FlightAuthorizationProposalStatus string

const (
	ProposalStatusPending   FlightAuthorizationProposalStatus = "PENDING"
	ProposalStatusApproved  FlightAuthorizationProposalStatus = "APPROVED"
	ProposalStatusNotified  FlightAuthorizationProposalStatus = "NOTIFIED"
	ProposalStatusActivated FlightAuthorizationProposalStatus = "ACTIVATED"
	ProposalStatusCompleted FlightAuthorizationProposalStatus = "COMPLETED"
)

type FlightAuthorizationProposal struct {
	ID                uuid.UUID                         `json:"id" gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	Name              string                            `json:"name" gorm:"not null"`
	Operator          Operator                          `json:"operator" gorm:"foreignKey:FlightAuthorizationProposalID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	Drones            []Drone                           `json:"drones" gorm:"foreignKey:FlightAuthorizationProposalID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	FlightPurpose     string                            `json:"flight_purpose" gorm:"not null"`
	FlightArea        []FlightArea                      `json:"flight_area" gorm:"foreignKey:FlightAuthorizationProposalID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	OperatingDuration OperatingDuration                 `json:"operating_duration" gorm:"embedded"`
	Airport           string                            `json:"airport" gorm:"not null"`
	Pilot             Pilot                             `json:"pilot" gorm:"foreignKey:FlightAuthorizationProposalID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	FilePaths         datatypes.JSONSlice[string]       `json:"file_paths" gorm:"type:json"`
	Status            FlightAuthorizationProposalStatus `json:"status" gorm:"type:varchar(32);default:PENDING"`
	CreatedAt         time.Time                         `json:"created_at"`
	UpdatedAt         time.Time                         `json:"updated_at"`
}
