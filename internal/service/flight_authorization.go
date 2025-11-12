package service

import (
	"context"
	"errors"
	"time"

	"172.21.5.249/airtrans/at-flight-authorization/internal/models"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
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

func prepareProposalForCreate(payload *models.FlightAuthorizationProposal) {
	if payload == nil {
		return
	}
	if payload.ID == uuid.Nil {
		payload.ID = uuid.New()
	}
	if payload.Operator.ID == uuid.Nil {
		payload.Operator.ID = uuid.New()
	}
	payload.Operator.FlightAuthorizationProposalID = payload.ID
	for i := range payload.Drones {
		if payload.Drones[i].ID == uuid.Nil {
			payload.Drones[i].ID = uuid.New()
		}
		payload.Drones[i].FlightAuthorizationProposalID = payload.ID
	}
	for i := range payload.FlightArea {
		if payload.FlightArea[i].ID == uuid.Nil {
			payload.FlightArea[i].ID = uuid.New()
		}
		payload.FlightArea[i].FlightAuthorizationProposalID = payload.ID
		for j := range payload.FlightArea[i].Polygon {
			if payload.FlightArea[i].Polygon[j].ID == uuid.Nil {
				payload.FlightArea[i].Polygon[j].ID = uuid.New()
			}
			payload.FlightArea[i].Polygon[j].FlightAreaID = payload.FlightArea[i].ID
		}
	}
	if payload.Pilot.ID == uuid.Nil {
		payload.Pilot.ID = uuid.New()
	}
	payload.Pilot.FlightAuthorizationProposalID = payload.ID
}

func prepareApprovalForCreate(payload *models.FlightAuthorizationApproval) {
	if payload == nil {
		return
	}
	if payload.ID == uuid.Nil {
		payload.ID = uuid.New()
	}
	for i := range payload.AuthorizedFlightArea {
		if payload.AuthorizedFlightArea[i].ID == uuid.Nil {
			payload.AuthorizedFlightArea[i].ID = uuid.New()
		}
		payload.AuthorizedFlightArea[i].FlightAuthorizationApprovalID = payload.ID
		for j := range payload.AuthorizedFlightArea[i].Polygon {
			if payload.AuthorizedFlightArea[i].Polygon[j].ID == uuid.Nil {
				payload.AuthorizedFlightArea[i].Polygon[j].ID = uuid.New()
			}
			payload.AuthorizedFlightArea[i].Polygon[j].AuthorizedFlightAreaID = payload.AuthorizedFlightArea[i].ID
		}
	}
	if payload.FlightParameter.ID == uuid.Nil {
		payload.FlightParameter.ID = uuid.New()
	}
	payload.FlightParameter.FlightAuthorizationApprovalID = payload.ID
}

func prepareNotificationForCreate(payload *models.FlightNotification) {
	if payload == nil {
		return
	}
	if payload.ID == uuid.Nil {
		payload.ID = uuid.New()
	}
	for i := range payload.IntendedFlightArea {
		if payload.IntendedFlightArea[i].ID == uuid.Nil {
			payload.IntendedFlightArea[i].ID = uuid.New()
		}
		payload.IntendedFlightArea[i].FlightNotificationID = payload.ID
		for j := range payload.IntendedFlightArea[i].Polygon {
			if payload.IntendedFlightArea[i].Polygon[j].ID == uuid.Nil {
				payload.IntendedFlightArea[i].Polygon[j].ID = uuid.New()
			}
			payload.IntendedFlightArea[i].Polygon[j].IntendedFlightAreaID = payload.IntendedFlightArea[i].ID
		}
	}
}

func (s *Service) CreateFlightAuthorizationProposal(ctx context.Context, payload *models.FlightAuthorizationProposal) (*models.FlightAuthorizationProposal, error) {
	if payload == nil {
		return nil, errors.New("payload is nil")
	}
	prepareProposalForCreate(payload)
	payload.Status = models.ProposalStatusPending
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
		case "status":
			tx = tx.Where("status = ?", value)
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
			"name":                      payload.Name,
			"flight_purpose":            payload.FlightPurpose,
			"take_off_and_landing_area": payload.TakeOffAndLandingArea,
			"duration":                  payload.OperatingDuration.Duration,
			"from_day":                  payload.OperatingDuration.FromDay,
			"to_day":                    payload.OperatingDuration.ToDay,
		}
		if err := tx.Model(&models.FlightAuthorizationProposal{}).Where("id = ?", id).Updates(updates).Error; err != nil {
			return err
		}
		if err := tx.Where("flight_authorization_proposal_id = ?", id).Delete(&models.Operator{}).Error; err != nil {
			return err
		}
		payload.Operator.ID = uuid.New()
		payload.Operator.FlightAuthorizationProposalID = id
		if err := tx.Create(&payload.Operator).Error; err != nil {
			return err
		}
		if err := tx.Where("flight_authorization_proposal_id = ?", id).Delete(&models.Drone{}).Error; err != nil {
			return err
		}
		for i := range payload.Drones {
			if payload.Drones[i].ID == uuid.Nil {
				payload.Drones[i].ID = uuid.New()
			}
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
			if payload.FlightArea[i].ID == uuid.Nil {
				payload.FlightArea[i].ID = uuid.New()
			}
			payload.FlightArea[i].FlightAuthorizationProposalID = id
			for j := range payload.FlightArea[i].Polygon {
				if payload.FlightArea[i].Polygon[j].ID == uuid.Nil {
					payload.FlightArea[i].Polygon[j].ID = uuid.New()
				}
				payload.FlightArea[i].Polygon[j].FlightAreaID = payload.FlightArea[i].ID
			}
		}
		if len(payload.FlightArea) > 0 {
			if err := tx.Session(&gorm.Session{FullSaveAssociations: true}).Create(&payload.FlightArea).Error; err != nil {
				return err
			}
		}
		if err := tx.Where("flight_authorization_proposal_id = ?", id).Delete(&models.Pilot{}).Error; err != nil {
			return err
		}
		payload.Pilot.ID = uuid.New()
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
	prepareApprovalForCreate(payload)
	var out models.FlightAuthorizationApproval
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Session(&gorm.Session{FullSaveAssociations: true}).Create(payload).Error; err != nil {
			return err
		}
		if err := s.recalcProposalStatusWithTx(ctx, tx, payload.FlightAuthorizationProposalID); err != nil {
			return err
		}
		return approvalPreload(tx).First(&out, "id = ?", payload.ID).Error
	})
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
		oldProposalID := existing.FlightAuthorizationProposalID
		updates := map[string]interface{}{
			"name":                             payload.Name,
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
			if payload.AuthorizedFlightArea[i].ID == uuid.Nil {
				payload.AuthorizedFlightArea[i].ID = uuid.New()
			}
			payload.AuthorizedFlightArea[i].FlightAuthorizationApprovalID = id
			for j := range payload.AuthorizedFlightArea[i].Polygon {
				if payload.AuthorizedFlightArea[i].Polygon[j].ID == uuid.Nil {
					payload.AuthorizedFlightArea[i].Polygon[j].ID = uuid.New()
				}
				payload.AuthorizedFlightArea[i].Polygon[j].AuthorizedFlightAreaID = payload.AuthorizedFlightArea[i].ID
			}
		}
		if len(payload.AuthorizedFlightArea) > 0 {
			if err := tx.Session(&gorm.Session{FullSaveAssociations: true}).Create(&payload.AuthorizedFlightArea).Error; err != nil {
				return err
			}
		}
		if err := tx.Where("flight_authorization_approval_id = ?", id).Delete(&models.FlightParameter{}).Error; err != nil {
			return err
		}
		if payload.FlightParameter.ID == uuid.Nil {
			payload.FlightParameter.ID = uuid.New()
		}
		payload.FlightParameter.FlightAuthorizationApprovalID = id
		if err := tx.Create(&payload.FlightParameter).Error; err != nil {
			return err
		}
		if err := approvalPreload(tx).First(&result, "id = ?", id).Error; err != nil {
			return err
		}
		if err := s.recalcProposalStatusWithTx(ctx, tx, result.FlightAuthorizationProposalID); err != nil {
			return err
		}
		if oldProposalID != result.FlightAuthorizationProposalID {
			if err := s.recalcProposalStatusWithTx(ctx, tx, oldProposalID); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (s *Service) DeleteFlightAuthorizationApproval(ctx context.Context, id uuid.UUID) error {
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var approval models.FlightAuthorizationApproval
		if err := tx.First(&approval, "id = ?", id).Error; err != nil {
			return err
		}
		if err := tx.Delete(&models.FlightAuthorizationApproval{}, "id = ?", id).Error; err != nil {
			return err
		}
		return s.recalcProposalStatusWithTx(ctx, tx, approval.FlightAuthorizationProposalID)
	})
}

func (s *Service) CreateFlightNotification(ctx context.Context, payload *models.FlightNotification) (*models.FlightNotification, error) {
	if payload == nil {
		return nil, errors.New("payload is nil")
	}
	prepareNotificationForCreate(payload)
	var out models.FlightNotification
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Session(&gorm.Session{FullSaveAssociations: true}).Create(payload).Error; err != nil {
			return err
		}
		if err := notificationPreload(tx).First(&out, "id = ?", payload.ID).Error; err != nil {
			return err
		}
		proposalID := out.FlightAuthorizationApproval.FlightAuthorizationProposalID
		return s.recalcProposalStatusWithTx(ctx, tx, proposalID)
	})
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
			"name":                             payload.Name,
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
			if payload.IntendedFlightArea[i].ID == uuid.Nil {
				payload.IntendedFlightArea[i].ID = uuid.New()
			}
			payload.IntendedFlightArea[i].FlightNotificationID = id
			for j := range payload.IntendedFlightArea[i].Polygon {
				if payload.IntendedFlightArea[i].Polygon[j].ID == uuid.Nil {
					payload.IntendedFlightArea[i].Polygon[j].ID = uuid.New()
				}
				payload.IntendedFlightArea[i].Polygon[j].IntendedFlightAreaID = payload.IntendedFlightArea[i].ID
			}
		}
		if len(payload.IntendedFlightArea) > 0 {
			if err := tx.Session(&gorm.Session{FullSaveAssociations: true}).Create(&payload.IntendedFlightArea).Error; err != nil {
				return err
			}
		}
		if err := notificationPreload(tx).First(&result, "id = ?", id).Error; err != nil {
			return err
		}
		proposalID := result.FlightAuthorizationApproval.FlightAuthorizationProposalID
		return s.recalcProposalStatusWithTx(ctx, tx, proposalID)
	})
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (s *Service) DeleteFlightNotification(ctx context.Context, id uuid.UUID) error {
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var notification models.FlightNotification
		if err := notificationPreload(tx).First(&notification, "id = ?", id).Error; err != nil {
			return err
		}
		if err := tx.Delete(&models.FlightNotification{}, "id = ?", id).Error; err != nil {
			return err
		}
		proposalID := notification.FlightAuthorizationApproval.FlightAuthorizationProposalID
		return s.recalcProposalStatusWithTx(ctx, tx, proposalID)
	})
}

func (s *Service) recalcProposalStatus(ctx context.Context, proposalID uuid.UUID) error {
	return s.recalcProposalStatusWithTx(ctx, s.db.WithContext(ctx), proposalID)
}

func (s *Service) recalcProposalStatusWithTx(ctx context.Context, tx *gorm.DB, proposalID uuid.UUID) error {
	if proposalID == uuid.Nil {
		return nil
	}
	var notifications []models.FlightNotification
	err := tx.
		Joins("JOIN flight_authorization_approvals ON flight_authorization_approvals.id = flight_notifications.flight_authorization_approval_id").
		Where("flight_authorization_approvals.flight_authorization_proposal_id = ?", proposalID).
		Find(&notifications).Error
	if err != nil {
		return err
	}
	now := time.Now()
	if len(notifications) == 0 {
		var approvalsCount int64
		if err := tx.Model(&models.FlightAuthorizationApproval{}).
			Where("flight_authorization_proposal_id = ?", proposalID).
			Count(&approvalsCount).Error; err != nil {
			return err
		}
		status := models.ProposalStatusPending
		if approvalsCount > 0 {
			status = models.ProposalStatusApproved
		}
		return tx.Model(&models.FlightAuthorizationProposal{}).
			Where("id = ?", proposalID).
			Update("status", status).Error
	}
	idx := 0
	for i := 1; i < len(notifications); i++ {
		if isLaterDuration(notifications[i].IntendedOperatingDuration, notifications[idx].IntendedOperatingDuration) {
			idx = i
		}
	}
	status := proposalStatusFromNotification(now, notifications[idx].IntendedOperatingDuration)
	return tx.Model(&models.FlightAuthorizationProposal{}).
		Where("id = ?", proposalID).
		Update("status", status).Error
}

func isLaterDuration(a, b models.OperatingDuration) bool {
	if a.ToDay.IsZero() {
		if b.ToDay.IsZero() {
			if a.FromDay.IsZero() {
				return false
			}
			if b.FromDay.IsZero() {
				return false
			}
			return a.FromDay.After(b.FromDay)
		}
		return false
	}
	if b.ToDay.IsZero() {
		return true
	}
	if a.ToDay.Equal(b.ToDay) {
		if a.FromDay.IsZero() {
			return false
		}
		if b.FromDay.IsZero() {
			return true
		}
		return a.FromDay.After(b.FromDay)
	}
	return a.ToDay.After(b.ToDay)
}

func proposalStatusFromNotification(now time.Time, duration models.OperatingDuration) models.FlightAuthorizationProposalStatus {
	if duration.FromDay.IsZero() || duration.ToDay.IsZero() {
		return models.ProposalStatusNotified
	}
	if now.Before(duration.FromDay) {
		return models.ProposalStatusNotified
	}
	if now.After(duration.ToDay) {
		return models.ProposalStatusCompleted
	}
	return models.ProposalStatusActivated
}

func (s *Service) refreshProposalStatuses(ctx context.Context) error {
	var proposalIDs []uuid.UUID
	err := s.db.WithContext(ctx).
		Model(&models.FlightAuthorizationProposal{}).
		Where("status IN ?", []models.FlightAuthorizationProposalStatus{
			models.ProposalStatusNotified,
			models.ProposalStatusActivated,
		}).
		Pluck("id", &proposalIDs).Error
	if err != nil {
		return err
	}
	for _, id := range proposalIDs {
		if err := s.recalcProposalStatus(ctx, id); err != nil {
			return err
		}
	}
	return nil
}

func (s *Service) startProposalStatusWatcher() {
	interval := s.statusInterval
	if interval <= 0 {
		interval = time.Minute
	}
	go func() {
		if err := s.refreshProposalStatuses(context.Background()); err != nil {
			log.Error().Err(err).Msg("refresh proposal statuses")
		}
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for range ticker.C {
			ctx, cancel := context.WithTimeout(context.Background(), interval)
			if err := s.refreshProposalStatuses(ctx); err != nil {
				log.Error().Err(err).Msg("refresh proposal statuses")
			}
			cancel()
		}
	}()
}
