package service

import (
	"context"
	"errors"

	"172.21.5.249/airtrans/at-flight-authorization/internal/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

func proposalPreload(tx *gorm.DB) *gorm.DB {
	return tx.
		Preload("Operator").
		Preload("Drones").
		Preload("FlightArea").
		Preload("FlightArea.Polygon").
		Preload("Pilot")
}

func approvalPreload(tx *gorm.DB) *gorm.DB {
	return tx.
		Preload("AuthorizedFlightArea").
		Preload("AuthorizedFlightArea.Polygon").
		Preload("FlightParameter").
		Preload("FlightAuthorizationProposal", func(tx *gorm.DB) *gorm.DB {
			return proposalPreload(tx)
		})
}

func notificationPreload(tx *gorm.DB) *gorm.DB {
	return tx.
		Preload("IntendedFlightArea").
		Preload("IntendedFlightArea.Polygon").
		Preload("FlightAuthorizationApproval", func(tx *gorm.DB) *gorm.DB {
			return approvalPreload(tx)
		})
}

func (s *Service) CreateFlightAuthorizationProposal(ctx context.Context, payload *models.FlightAuthorizationProposal) (*models.FlightAuthorizationProposal, error) {
	if payload == nil {
		return nil, errors.New("payload is nil")
	}
	err := s.db.WithContext(ctx).Session(&gorm.Session{FullSaveAssociations: true}).Create(payload).Error
	if err != nil {
		return nil, err
	}
	var out models.FlightAuthorizationProposal
	err = proposalPreload(s.db.WithContext(ctx)).First(&out, "id = ?", payload.ID).Error
	if err != nil {
		return nil, err
	}
	return &out, nil
}

func (s *Service) GetFlightAuthorizationProposal(ctx context.Context, id uuid.UUID) (*models.FlightAuthorizationProposal, error) {
	var out models.FlightAuthorizationProposal
	err := proposalPreload(s.db.WithContext(ctx)).First(&out, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &out, nil
}

func applyProposalFilters(tx *gorm.DB, filters map[string][]string) (*gorm.DB, error) {
	for key, values := range filters {
		if len(values) == 0 {
			continue
		}
		value := values[0]
		if value == "" {
			continue
		}
		switch key {
		case "id":
			if _, err := uuid.Parse(value); err != nil {
				return tx, err
			}
			tx = tx.Where("id = ?", value)
		case "flight_purpose":
			tx = tx.Where("flight_purpose = ?", value)
		case "airport":
			tx = tx.Where("airport = ?", value)
		default:
		}
	}
	return tx, nil
}

func (s *Service) ListFlightAuthorizationProposals(ctx context.Context, filters map[string][]string, page, size int) ([]models.FlightAuthorizationProposal, int64, error) {
	base := s.db.WithContext(ctx).Model(&models.FlightAuthorizationProposal{})
	base, err := applyProposalFilters(base, filters)
	if err != nil {
		return nil, 0, err
	}
	var total int64
	err = base.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}
	query := proposalPreload(s.db.WithContext(ctx))
	query, err = applyProposalFilters(query, filters)
	if err != nil {
		return nil, 0, err
	}
	if size > 0 {
		if page <= 0 {
			page = 1
		}
		offset := (page - 1) * size
		query = query.Offset(offset).Limit(size)
	}
	var items []models.FlightAuthorizationProposal
	err = query.Order("created_at DESC").Find(&items).Error
	if err != nil {
		return nil, 0, err
	}
	return items, total, nil
}

func (s *Service) UpdateFlightAuthorizationProposal(ctx context.Context, id uuid.UUID, payload *models.FlightAuthorizationProposal) (*models.FlightAuthorizationProposal, error) {
	if payload == nil {
		return nil, errors.New("payload is nil")
	}
	var result models.FlightAuthorizationProposal
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var existing models.FlightAuthorizationProposal
		if err := tx.First(&existing, "id = ?", id).Error; err != nil {
			return err
		}
		updates := map[string]interface{}{
			"flight_purpose": payload.FlightPurpose,
			"airport":        payload.Airport,
			"duration":       payload.OperatingDuration.Duration,
			"from_day":       payload.OperatingDuration.FromDay,
			"to_day":         payload.OperatingDuration.ToDay,
		}
		if err := tx.Model(&models.FlightAuthorizationProposal{}).Where("id = ?", id).Updates(updates).Error; err != nil {
			return err
		}
		if err := tx.Where("flight_authorization_proposal_id = ?", id).Delete(&models.Operator{}).Error; err != nil {
			return err
		}
		payload.Operator.FlightAuthorizationProposalID = id
		if err := tx.Create(&payload.Operator).Error; err != nil {
			return err
		}
		if err := tx.Where("flight_authorization_proposal_id = ?", id).Delete(&models.Drone{}).Error; err != nil {
			return err
		}
		for i := range payload.Drones {
			payload.Drones[i].FlightAuthorizationProposalID = id
		}
		if len(payload.Drones) > 0 {
			if err := tx.Session(&gorm.Session{FullSaveAssociations: true}).Create(&payload.Drones).Error; err != nil {
				return err
			}
		}
		if err := tx.Where("flight_authorization_proposal_id = ?", id).Delete(&models.FlightArea{}).Error; err != nil {
			return err
		}
		for i := range payload.FlightArea {
			payload.FlightArea[i].FlightAuthorizationProposalID = id
		}
		if len(payload.FlightArea) > 0 {
			if err := tx.Session(&gorm.Session{FullSaveAssociations: true}).Create(&payload.FlightArea).Error; err != nil {
				return err
			}
		}
		if err := tx.Where("flight_authorization_proposal_id = ?", id).Delete(&models.Pilot{}).Error; err != nil {
			return err
		}
		payload.Pilot.FlightAuthorizationProposalID = id
		if err := tx.Create(&payload.Pilot).Error; err != nil {
			return err
		}
		return proposalPreload(tx).First(&result, "id = ?", id).Error
	})
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (s *Service) DeleteFlightAuthorizationProposal(ctx context.Context, id uuid.UUID) error {
	res := s.db.WithContext(ctx).Delete(&models.FlightAuthorizationProposal{}, "id = ?", id)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func (s *Service) CreateFlightAuthorizationApproval(ctx context.Context, payload *models.FlightAuthorizationApproval) (*models.FlightAuthorizationApproval, error) {
	if payload == nil {
		return nil, errors.New("payload is nil")
	}
	err := s.db.WithContext(ctx).Session(&gorm.Session{FullSaveAssociations: true}).Create(payload).Error
	if err != nil {
		return nil, err
	}
	var out models.FlightAuthorizationApproval
	err = approvalPreload(s.db.WithContext(ctx)).First(&out, "id = ?", payload.ID).Error
	if err != nil {
		return nil, err
	}
	return &out, nil
}

func (s *Service) GetFlightAuthorizationApproval(ctx context.Context, id uuid.UUID) (*models.FlightAuthorizationApproval, error) {
	var out models.FlightAuthorizationApproval
	err := approvalPreload(s.db.WithContext(ctx)).First(&out, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &out, nil
}

func applyApprovalFilters(tx *gorm.DB, filters map[string][]string) (*gorm.DB, error) {
	for key, values := range filters {
		if len(values) == 0 {
			continue
		}
		value := values[0]
		if value == "" {
			continue
		}
		switch key {
		case "id":
			if _, err := uuid.Parse(value); err != nil {
				return tx, err
			}
			tx = tx.Where("id = ?", value)
		case "flight_authorization_proposal_id":
			if _, err := uuid.Parse(value); err != nil {
				return tx, err
			}
			tx = tx.Where("flight_authorization_proposal_id = ?", value)
		default:
		}
	}
	return tx, nil
}

func (s *Service) ListFlightAuthorizationApprovals(ctx context.Context, filters map[string][]string, page, size int) ([]models.FlightAuthorizationApproval, int64, error) {
	base := s.db.WithContext(ctx).Model(&models.FlightAuthorizationApproval{})
	base, err := applyApprovalFilters(base, filters)
	if err != nil {
		return nil, 0, err
	}
	var total int64
	err = base.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}
	query := approvalPreload(s.db.WithContext(ctx))
	query, err = applyApprovalFilters(query, filters)
	if err != nil {
		return nil, 0, err
	}
	if size > 0 {
		if page <= 0 {
			page = 1
		}
		offset := (page - 1) * size
		query = query.Offset(offset).Limit(size)
	}
	var items []models.FlightAuthorizationApproval
	err = query.Order("created_at DESC").Find(&items).Error
	if err != nil {
		return nil, 0, err
	}
	return items, total, nil
}

func (s *Service) UpdateFlightAuthorizationApproval(ctx context.Context, id uuid.UUID, payload *models.FlightAuthorizationApproval) (*models.FlightAuthorizationApproval, error) {
	if payload == nil {
		return nil, errors.New("payload is nil")
	}
	var result models.FlightAuthorizationApproval
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var existing models.FlightAuthorizationApproval
		if err := tx.First(&existing, "id = ?", id).Error; err != nil {
			return err
		}
		updates := map[string]interface{}{
			"flight_authorization_proposal_id": payload.FlightAuthorizationProposalID,
			"flight_negotiation_authorities":   payload.FlightNegotiationAuthorities,
			"duration":                         payload.AuthorizedOperatingDuration.Duration,
			"from_day":                         payload.AuthorizedOperatingDuration.FromDay,
			"to_day":                           payload.AuthorizedOperatingDuration.ToDay,
		}
		if err := tx.Model(&models.FlightAuthorizationApproval{}).Where("id = ?", id).Updates(updates).Error; err != nil {
			return err
		}
		if err := tx.Where("flight_authorization_approval_id = ?", id).Delete(&models.AuthorizedFlightArea{}).Error; err != nil {
			return err
		}
		for i := range payload.AuthorizedFlightArea {
			payload.AuthorizedFlightArea[i].FlightAuthorizationApprovalID = id
		}
		if len(payload.AuthorizedFlightArea) > 0 {
			if err := tx.Session(&gorm.Session{FullSaveAssociations: true}).Create(&payload.AuthorizedFlightArea).Error; err != nil {
				return err
			}
		}
		if err := tx.Where("flight_authorization_approval_id = ?", id).Delete(&models.FlightParameter{}).Error; err != nil {
			return err
		}
		payload.FlightParameter.FlightAuthorizationApprovalID = id
		if err := tx.Create(&payload.FlightParameter).Error; err != nil {
			return err
		}
		return approvalPreload(tx).First(&result, "id = ?", id).Error
	})
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (s *Service) DeleteFlightAuthorizationApproval(ctx context.Context, id uuid.UUID) error {
	res := s.db.WithContext(ctx).Delete(&models.FlightAuthorizationApproval{}, "id = ?", id)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func (s *Service) CreateFlightNotification(ctx context.Context, payload *models.FlightNotification) (*models.FlightNotification, error) {
	if payload == nil {
		return nil, errors.New("payload is nil")
	}
	err := s.db.WithContext(ctx).Session(&gorm.Session{FullSaveAssociations: true}).Create(payload).Error
	if err != nil {
		return nil, err
	}
	var out models.FlightNotification
	err = notificationPreload(s.db.WithContext(ctx)).First(&out, "id = ?", payload.ID).Error
	if err != nil {
		return nil, err
	}
	return &out, nil
}

func (s *Service) GetFlightNotification(ctx context.Context, id uuid.UUID) (*models.FlightNotification, error) {
	var out models.FlightNotification
	err := notificationPreload(s.db.WithContext(ctx)).First(&out, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &out, nil
}

func applyNotificationFilters(tx *gorm.DB, filters map[string][]string) (*gorm.DB, error) {
	for key, values := range filters {
		if len(values) == 0 {
			continue
		}
		value := values[0]
		if value == "" {
			continue
		}
		switch key {
		case "id":
			if _, err := uuid.Parse(value); err != nil {
				return tx, err
			}
			tx = tx.Where("id = ?", value)
		case "flight_authorization_approval_id":
			if _, err := uuid.Parse(value); err != nil {
				return tx, err
			}
			tx = tx.Where("flight_authorization_approval_id = ?", value)
		default:
		}
	}
	return tx, nil
}

func (s *Service) ListFlightNotifications(ctx context.Context, filters map[string][]string, page, size int) ([]models.FlightNotification, int64, error) {
	base := s.db.WithContext(ctx).Model(&models.FlightNotification{})
	base, err := applyNotificationFilters(base, filters)
	if err != nil {
		return nil, 0, err
	}
	var total int64
	err = base.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}
	query := notificationPreload(s.db.WithContext(ctx))
	query, err = applyNotificationFilters(query, filters)
	if err != nil {
		return nil, 0, err
	}
	if size > 0 {
		if page <= 0 {
			page = 1
		}
		offset := (page - 1) * size
		query = query.Offset(offset).Limit(size)
	}
	var items []models.FlightNotification
	err = query.Order("id DESC").Find(&items).Error
	if err != nil {
		return nil, 0, err
	}
	return items, total, nil
}

func (s *Service) UpdateFlightNotification(ctx context.Context, id uuid.UUID, payload *models.FlightNotification) (*models.FlightNotification, error) {
	if payload == nil {
		return nil, errors.New("payload is nil")
	}
	var result models.FlightNotification
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var existing models.FlightNotification
		if err := tx.First(&existing, "id = ?", id).Error; err != nil {
			return err
		}
		updates := map[string]interface{}{
			"flight_authorization_approval_id": payload.FlightAuthorizationApprovalID,
			"duration":                         payload.IntendedOperatingDuration.Duration,
			"from_day":                         payload.IntendedOperatingDuration.FromDay,
			"to_day":                           payload.IntendedOperatingDuration.ToDay,
		}
		if err := tx.Model(&models.FlightNotification{}).Where("id = ?", id).Updates(updates).Error; err != nil {
			return err
		}
		if err := tx.Where("flight_notification_id = ?", id).Delete(&models.IntendedFlightArea{}).Error; err != nil {
			return err
		}
		for i := range payload.IntendedFlightArea {
			payload.IntendedFlightArea[i].FlightNotificationID = id
		}
		if len(payload.IntendedFlightArea) > 0 {
			if err := tx.Session(&gorm.Session{FullSaveAssociations: true}).Create(&payload.IntendedFlightArea).Error; err != nil {
				return err
			}
		}
		return notificationPreload(tx).First(&result, "id = ?", id).Error
	})
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (s *Service) DeleteFlightNotification(ctx context.Context, id uuid.UUID) error {
	res := s.db.WithContext(ctx).Delete(&models.FlightNotification{}, "id = ?", id)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}
