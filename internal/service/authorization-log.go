package service

import (
	"context"
	"time"

	"172.21.5.249/airtrans/at-flight-authorization/internal/models"
	"gorm.io/gorm"
)

type AuthorizationLogFilter struct {
	Username  string
	Resource  string
	Action    string
	Client    string
	StartTime *time.Time
	EndTime   *time.Time
	Result    string
	Since     int64
	URI       string
}

func (s *Service) List(ctx context.Context, f AuthorizationLogFilter) ([]models.AuthorizationLog, error) {
	var logs []models.AuthorizationLog
	tx := s.buildQuery(ctx, f)

	if err := tx.Order("timestamp DESC").Find(&logs).Error; err != nil {
		return nil, err
	}

	return logs, nil
}

func (s *Service) Count(ctx context.Context, f AuthorizationLogFilter) (int64, error) {
	var count int64
	tx := s.buildQuery(ctx, f)

	if err := tx.Count(&count).Error; err != nil {
		return 0, err
	}

	return count, nil
}

func (s *Service) buildQuery(ctx context.Context, f AuthorizationLogFilter) *gorm.DB {
	tx := s.db.WithContext(ctx).Model(&models.AuthorizationLog{})

	if f.Username != "" {
		tx = tx.Where("username = ?", f.Username)
	}
	if f.Resource != "" {
		tx = tx.Where("resource = ?", f.Resource)
	}
	if f.Action != "" {
		tx = tx.Where("action = ?", f.Action)
	}
	if f.Client != "" {
		tx = tx.Where("client = ?", f.Client)
	}
	if f.Since > 0 {
		tx = tx.Where("timestamp > ?", f.Since)
	} else if f.StartTime != nil {
		tx = tx.Where("timestamp >= ?", f.StartTime.Unix())
	}
	if f.EndTime != nil {
		tx = tx.Where("timestamp <= ?", f.EndTime.Unix())
	}
	if f.Result != "" {
		tx = tx.Where("result = ?", f.Result)
	}
	if f.URI != "" {
		tx = tx.Where("uri = ?", f.URI)
	}

	return tx
}
