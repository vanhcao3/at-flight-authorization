package service

import (
	"context"
	"errors"

	"172.21.5.249/airtrans/at-flight-authorization/internal/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type FlightAuthorizationProposalListOptions struct {
	FlightPurpose string
	Airport       string
	Page          int
	Size          int
}

func (s *Service) CreateFlightAuthorizationProposal(ctx context.Context, proposal *models.FlightAuthorizationProposal) (*models.FlightAuthorizationProposal, error) {
	if proposal == nil {
		return nil, errors.New("proposal payload is nil")
	}
	prepareFlightAuthorizationProposal(proposal, uuid.New())
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return tx.Session(&gorm.Session{FullSaveAssociations: true}).Create(proposal).Error
	})
	if err != nil {
		return nil, err
	}
	return s.GetFlightAuthorizationProposalByID(ctx, proposal.ID)
}

func (s *Service) GetFlightAuthorizationProposalByID(ctx context.Context, id uuid.UUID) (*models.FlightAuthorizationProposal, error) {
	var proposal models.FlightAuthorizationProposal
	tx := applyFlightAuthorizationProposalPreloads(s.db.WithContext(ctx))
	if err := tx.First(&proposal, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &proposal, nil
}

func (s *Service) ListFlightAuthorizationProposals(ctx context.Context, opts FlightAuthorizationProposalListOptions) ([]models.FlightAuthorizationProposal, error) {
	var proposals []models.FlightAuthorizationProposal
	tx := applyFlightAuthorizationProposalPreloads(s.db.WithContext(ctx).Model(&models.FlightAuthorizationProposal{}))
	if opts.FlightPurpose != "" {
		tx = tx.Where("flight_purpose ILIKE ?", "%"+opts.FlightPurpose+"%")
	}
	if opts.Airport != "" {
		tx = tx.Where("airport ILIKE ?", "%"+opts.Airport+"%")
	}
	if opts.Page > 0 && opts.Size > 0 {
		tx = tx.Offset((opts.Page - 1) * opts.Size).Limit(opts.Size)
	}
	if err := tx.Order("created_at DESC").Find(&proposals).Error; err != nil {
		return nil, err
	}
	return proposals, nil
}

func (s *Service) UpdateFlightAuthorizationProposal(ctx context.Context, id uuid.UUID, proposal *models.FlightAuthorizationProposal) (*models.FlightAuthorizationProposal, error) {
	if proposal == nil {
		return nil, errors.New("proposal payload is nil")
	}
	_, err := s.GetFlightAuthorizationProposalByID(ctx, id)
	if err != nil {
		return nil, err
	}
	prepareFlightAuthorizationProposal(proposal, id)
	err = s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&models.FlightAuthorizationProposal{}).
			Where("id = ?", id).
			Omit("Operator", "Drones", "FlightArea", "Pilot").
			Updates(proposal).Error; err != nil {
			return err
		}
		if err := tx.Where("flight_authorization_proposal_id = ?", id).Delete(&models.Operator{}).Error; err != nil {
			return err
		}
		if err := tx.Where("flight_authorization_proposal_id = ?", id).Delete(&models.Drone{}).Error; err != nil {
			return err
		}
		if err := tx.Where("flight_authorization_proposal_id = ?", id).Delete(&models.FlightArea{}).Error; err != nil {
			return err
		}
		if err := tx.Where("flight_authorization_proposal_id = ?", id).Delete(&models.Pilot{}).Error; err != nil {
			return err
		}
		if err := tx.Session(&gorm.Session{FullSaveAssociations: true}).Create(&proposal.Operator).Error; err != nil {
			return err
		}
		for i := range proposal.Drones {
			item := proposal.Drones[i]
			if err := tx.Create(&item).Error; err != nil {
				return err
			}
		}
		for i := range proposal.FlightArea {
			area := proposal.FlightArea[i]
			if err := tx.Session(&gorm.Session{FullSaveAssociations: true}).Create(&area).Error; err != nil {
				return err
			}
		}
		if err := tx.Session(&gorm.Session{FullSaveAssociations: true}).Create(&proposal.Pilot).Error; err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return s.GetFlightAuthorizationProposalByID(ctx, id)
}

func (s *Service) DeleteFlightAuthorizationProposal(ctx context.Context, id uuid.UUID) error {
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		result := tx.Delete(&models.FlightAuthorizationProposal{}, "id = ?", id)
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return gorm.ErrRecordNotFound
		}
		return nil
	})
}

func applyFlightAuthorizationProposalPreloads(tx *gorm.DB) *gorm.DB {
	preloads := []string{
		"Operator",
		"Drones",
		"FlightArea",
		"FlightArea.Polygon",
		"Pilot",
	}
	for _, preload := range preloads {
		tx = tx.Preload(preload)
	}
	return tx
}

func prepareFlightAuthorizationProposal(proposal *models.FlightAuthorizationProposal, id uuid.UUID) {
	proposal.ID = id
	if proposal.Operator.ID == uuid.Nil {
		proposal.Operator.ID = uuid.New()
	}
	proposal.Operator.FlightAuthorizationProposalID = proposal.ID
	for i := range proposal.Drones {
		if proposal.Drones[i].ID == uuid.Nil {
			proposal.Drones[i].ID = uuid.New()
		}
		proposal.Drones[i].FlightAuthorizationProposalID = proposal.ID
	}
	for i := range proposal.FlightArea {
		if proposal.FlightArea[i].ID == uuid.Nil {
			proposal.FlightArea[i].ID = uuid.New()
		}
		proposal.FlightArea[i].FlightAuthorizationProposalID = proposal.ID
		for j := range proposal.FlightArea[i].Polygon {
			if proposal.FlightArea[i].Polygon[j].ID == uuid.Nil {
				proposal.FlightArea[i].Polygon[j].ID = uuid.New()
			}
			proposal.FlightArea[i].Polygon[j].FlightAreaID = proposal.FlightArea[i].ID
		}
	}
	if proposal.Pilot.ID == uuid.Nil {
		proposal.Pilot.ID = uuid.New()
	}
	proposal.Pilot.FlightAuthorizationProposalID = proposal.ID
}
