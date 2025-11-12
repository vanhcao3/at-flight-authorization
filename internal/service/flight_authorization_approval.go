package service

import (
	"context"
	"errors"

	"172.21.5.249/airtrans/at-flight-authorization/internal/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type FlightAuthorizationApprovalListOptions struct {
	FlightAuthorizationProposalID uuid.UUID
	Page                          int
	Size                          int
}

func (s *Service) CreateFlightAuthorizationApproval(ctx context.Context, approval *models.FlightAuthorizationApproval) (*models.FlightAuthorizationApproval, error) {
	if approval == nil {
		return nil, errors.New("approval payload is nil")
	}
	if approval.FlightAuthorizationProposalID == uuid.Nil {
		return nil, errors.New("flight_authorization_proposal_id is required")
	}
	prepareFlightAuthorizationApproval(approval, uuid.New())
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return tx.Omit("FlightAuthorizationProposal").
			Session(&gorm.Session{FullSaveAssociations: true}).
			Create(approval).Error
	})
	if err != nil {
		return nil, err
	}
	return s.GetFlightAuthorizationApprovalByID(ctx, approval.ID)
}

func (s *Service) GetFlightAuthorizationApprovalByID(ctx context.Context, id uuid.UUID) (*models.FlightAuthorizationApproval, error) {
	var approval models.FlightAuthorizationApproval
	tx := applyFlightAuthorizationApprovalPreloads(s.db.WithContext(ctx))
	if err := tx.First(&approval, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &approval, nil
}

func (s *Service) ListFlightAuthorizationApprovals(ctx context.Context, opts FlightAuthorizationApprovalListOptions) ([]models.FlightAuthorizationApproval, error) {
	var approvals []models.FlightAuthorizationApproval
	tx := applyFlightAuthorizationApprovalPreloads(s.db.WithContext(ctx).Model(&models.FlightAuthorizationApproval{}))
	if opts.FlightAuthorizationProposalID != uuid.Nil {
		tx = tx.Where("flight_authorization_proposal_id = ?", opts.FlightAuthorizationProposalID)
	}
	if opts.Page > 0 && opts.Size > 0 {
		tx = tx.Offset((opts.Page - 1) * opts.Size).Limit(opts.Size)
	}
	if err := tx.Order("created_at DESC").Find(&approvals).Error; err != nil {
		return nil, err
	}
	return approvals, nil
}

func (s *Service) UpdateFlightAuthorizationApproval(ctx context.Context, id uuid.UUID, approval *models.FlightAuthorizationApproval) (*models.FlightAuthorizationApproval, error) {
	if approval == nil {
		return nil, errors.New("approval payload is nil")
	}
	existing, err := s.GetFlightAuthorizationApprovalByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if approval.FlightAuthorizationProposalID == uuid.Nil {
		approval.FlightAuthorizationProposalID = existing.FlightAuthorizationProposalID
	}
	prepareFlightAuthorizationApproval(approval, id)
	err = s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&models.FlightAuthorizationApproval{}).
			Where("id = ?", id).
			Omit("FlightAuthorizationProposal", "AuthorizedFlightArea", "FlightParameter").
			Updates(approval).Error; err != nil {
			return err
		}
		if err := tx.Where("flight_authorization_approval_id = ?", id).Delete(&models.AuthorizedFlightArea{}).Error; err != nil {
			return err
		}
		if err := tx.Where("flight_authorization_approval_id = ?", id).Delete(&models.FlightParameter{}).Error; err != nil {
			return err
		}
		if approval.FlightParameter.ID != uuid.Nil {
			if err := tx.Create(&approval.FlightParameter).Error; err != nil {
				return err
			}
		}
		for i := range approval.AuthorizedFlightArea {
			area := approval.AuthorizedFlightArea[i]
			if err := tx.Session(&gorm.Session{FullSaveAssociations: true}).Create(&area).Error; err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return s.GetFlightAuthorizationApprovalByID(ctx, id)
}

func (s *Service) DeleteFlightAuthorizationApproval(ctx context.Context, id uuid.UUID) error {
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		result := tx.Delete(&models.FlightAuthorizationApproval{}, "id = ?", id)
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return gorm.ErrRecordNotFound
		}
		return nil
	})
}

func applyFlightAuthorizationApprovalPreloads(tx *gorm.DB) *gorm.DB {
	preloads := []string{
		"FlightAuthorizationProposal",
		"FlightAuthorizationProposal.Operator",
		"FlightAuthorizationProposal.Drones",
		"FlightAuthorizationProposal.FlightArea",
		"FlightAuthorizationProposal.FlightArea.Polygon",
		"FlightAuthorizationProposal.Pilot",
		"AuthorizedFlightArea",
		"AuthorizedFlightArea.Polygon",
		"FlightParameter",
	}
	for _, preload := range preloads {
		tx = tx.Preload(preload)
	}
	return tx
}

func prepareFlightAuthorizationApproval(approval *models.FlightAuthorizationApproval, id uuid.UUID) {
	approval.ID = id
	hasFlightParameter := approval.FlightParameter.Altitude != "" ||
		approval.FlightParameter.Radius != "" ||
		approval.FlightParameter.TakeOffAndLandingArea.Place != "" ||
		approval.FlightParameter.TakeOffAndLandingArea.Commune != "" ||
		approval.FlightParameter.TakeOffAndLandingArea.Province != ""
	if hasFlightParameter {
		if approval.FlightParameter.ID == uuid.Nil {
			approval.FlightParameter.ID = uuid.New()
		}
		approval.FlightParameter.FlightAuthorizationApprovalID = approval.ID
	} else {
		approval.FlightParameter = models.FlightParameter{}
	}
	for i := range approval.AuthorizedFlightArea {
		if approval.AuthorizedFlightArea[i].ID == uuid.Nil {
			approval.AuthorizedFlightArea[i].ID = uuid.New()
		}
		approval.AuthorizedFlightArea[i].FlightAuthorizationApprovalID = approval.ID
		for j := range approval.AuthorizedFlightArea[i].Polygon {
			if approval.AuthorizedFlightArea[i].Polygon[j].ID == uuid.Nil {
				approval.AuthorizedFlightArea[i].Polygon[j].ID = uuid.New()
			}
			approval.AuthorizedFlightArea[i].Polygon[j].AuthorizedFlightAreaID = approval.AuthorizedFlightArea[i].ID
		}
	}
}
