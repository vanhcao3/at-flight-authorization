package service

import (
	"context"
	"errors"

	"172.21.5.249/airtrans/at-flight-authorization/internal/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type FlightNotificationListOptions struct {
	FlightAuthorizationApprovalID uuid.UUID
	Page                          int
	Size                          int
}

func (s *Service) CreateFlightNotification(ctx context.Context, notification *models.FlightNotification) (*models.FlightNotification, error) {
	if notification == nil {
		return nil, errors.New("notification payload is nil")
	}
	if notification.FlightAuthorizationApprovalID == uuid.Nil {
		return nil, errors.New("flight_authorization_approval_id is required")
	}
	prepareFlightNotification(notification, uuid.New())
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return tx.Omit("FlightAuthorizationApproval").
			Session(&gorm.Session{FullSaveAssociations: true}).
			Create(notification).Error
	})
	if err != nil {
		return nil, err
	}
	return s.GetFlightNotificationByID(ctx, notification.ID)
}

func (s *Service) GetFlightNotificationByID(ctx context.Context, id uuid.UUID) (*models.FlightNotification, error) {
	var notification models.FlightNotification
	tx := applyFlightNotificationPreloads(s.db.WithContext(ctx))
	if err := tx.First(&notification, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &notification, nil
}

func (s *Service) ListFlightNotifications(ctx context.Context, opts FlightNotificationListOptions) ([]models.FlightNotification, error) {
	var notifications []models.FlightNotification
	tx := applyFlightNotificationPreloads(s.db.WithContext(ctx).Model(&models.FlightNotification{}))
	if opts.FlightAuthorizationApprovalID != uuid.Nil {
		tx = tx.Where("flight_authorization_approval_id = ?", opts.FlightAuthorizationApprovalID)
	}
	if opts.Page > 0 && opts.Size > 0 {
		tx = tx.Offset((opts.Page - 1) * opts.Size).Limit(opts.Size)
	}
	if err := tx.Order("created_at DESC").Find(&notifications).Error; err != nil {
		return nil, err
	}
	return notifications, nil
}

func (s *Service) UpdateFlightNotification(ctx context.Context, id uuid.UUID, notification *models.FlightNotification) (*models.FlightNotification, error) {
	if notification == nil {
		return nil, errors.New("notification payload is nil")
	}
	existing, err := s.GetFlightNotificationByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if notification.FlightAuthorizationApprovalID == uuid.Nil {
		notification.FlightAuthorizationApprovalID = existing.FlightAuthorizationApprovalID
	}
	prepareFlightNotification(notification, id)
	err = s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&models.FlightNotification{}).
			Where("id = ?", id).
			Omit("FlightAuthorizationApproval", "IntendedFlightArea").
			Updates(notification).Error; err != nil {
			return err
		}
		if err := tx.Where("flight_notification_id = ?", id).Delete(&models.IntendedFlightArea{}).Error; err != nil {
			return err
		}
		for i := range notification.IntendedFlightArea {
			area := notification.IntendedFlightArea[i]
			if err := tx.Session(&gorm.Session{FullSaveAssociations: true}).Create(&area).Error; err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return s.GetFlightNotificationByID(ctx, id)
}

func (s *Service) DeleteFlightNotification(ctx context.Context, id uuid.UUID) error {
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		result := tx.Delete(&models.FlightNotification{}, "id = ?", id)
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return gorm.ErrRecordNotFound
		}
		return nil
	})
}

func applyFlightNotificationPreloads(tx *gorm.DB) *gorm.DB {
	preloads := []string{
		"FlightAuthorizationApproval",
		"FlightAuthorizationApproval.FlightAuthorizationProposal",
		"FlightAuthorizationApproval.AuthorizedFlightArea",
		"FlightAuthorizationApproval.AuthorizedFlightArea.Polygon",
		"FlightAuthorizationApproval.FlightParameter",
		"IntendedFlightArea",
		"IntendedFlightArea.Polygon",
	}
	for _, preload := range preloads {
		tx = tx.Preload(preload)
	}
	return tx
}

func prepareFlightNotification(notification *models.FlightNotification, id uuid.UUID) {
	notification.ID = id
	for i := range notification.IntendedFlightArea {
		if notification.IntendedFlightArea[i].ID == uuid.Nil {
			notification.IntendedFlightArea[i].ID = uuid.New()
		}
		notification.IntendedFlightArea[i].FlightNotificationID = notification.ID
		for j := range notification.IntendedFlightArea[i].Polygon {
			if notification.IntendedFlightArea[i].Polygon[j].ID == uuid.Nil {
				notification.IntendedFlightArea[i].Polygon[j].ID = uuid.New()
			}
			notification.IntendedFlightArea[i].Polygon[j].IntendedFlightAreaID = notification.IntendedFlightArea[i].ID
		}
	}
}
