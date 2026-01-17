package admin

import (
	"context"
	"fmt"
	"time"

	"github.com/anthropics/agentsmesh/backend/internal/domain/admin"
	"github.com/anthropics/agentsmesh/backend/internal/domain/organization"
	"github.com/anthropics/agentsmesh/backend/internal/domain/promocode"
	"github.com/anthropics/agentsmesh/backend/internal/domain/user"
)

// PromoCodeListFilter represents filter options for listing promo codes
type PromoCodeListFilter struct {
	Type     *promocode.PromoCodeType
	PlanName *string
	IsActive *bool
	Search   *string
	Page     int
	PageSize int
}

// PromoCodeListResult represents the result of listing promo codes
type PromoCodeListResult struct {
	Data       []*promocode.PromoCode `json:"data"`
	Total      int64                  `json:"total"`
	Page       int                    `json:"page"`
	PageSize   int                    `json:"page_size"`
	TotalPages int                    `json:"total_pages"`
}

// PromoCodeUpdateInput represents the input for updating a promo code
type PromoCodeUpdateInput struct {
	Name           *string
	Description    *string
	MaxUses        *int
	MaxUsesPerOrg  *int
	ExpiresAt      *time.Time
	ClearExpiresAt bool
}

// RedemptionWithDetails represents a redemption with user and organization details
type RedemptionWithDetails struct {
	ID             int64      `json:"id"`
	PromoCodeID    int64      `json:"promo_code_id"`
	OrganizationID int64      `json:"organization_id"`
	UserID         int64      `json:"user_id"`
	PlanName       string     `json:"plan_name"`
	DurationMonths int        `json:"duration_months"`
	NewPeriodEnd   time.Time  `json:"new_period_end"`
	IPAddress      *string    `json:"ip_address,omitempty"`
	CreatedAt      time.Time  `json:"created_at"`
	User           *user.User `json:"user,omitempty"`
	Organization   *organization.Organization `json:"organization,omitempty"`
}

// RedemptionListResult represents the result of listing redemptions
type RedemptionListResult struct {
	Data       []*RedemptionWithDetails `json:"data"`
	Total      int64                    `json:"total"`
	Page       int                      `json:"page"`
	PageSize   int                      `json:"page_size"`
	TotalPages int                      `json:"total_pages"`
}

// createAuditLog is a helper to create audit log entries
func (s *Service) createAuditLog(ctx context.Context, adminUserID int64, action admin.AuditAction, targetType admin.TargetType, targetID int64, oldData, newData interface{}) {
	// Best effort - don't fail the main operation if audit logging fails
	_ = s.LogActionFromContext(ctx, adminUserID, action, targetType, targetID, oldData, newData, "", "")
}

// ListPromoCodes lists promo codes with filtering and pagination
func (s *Service) ListPromoCodes(ctx context.Context, filter *PromoCodeListFilter) (*PromoCodeListResult, error) {
	query := s.db.Model(&promocode.PromoCode{})

	if filter.Type != nil {
		query = query.Where("type = ?", *filter.Type)
	}
	if filter.PlanName != nil {
		query = query.Where("plan_name = ?", *filter.PlanName)
	}
	if filter.IsActive != nil {
		query = query.Where("is_active = ?", *filter.IsActive)
	}
	if filter.Search != nil && *filter.Search != "" {
		search := "%" + *filter.Search + "%"
		query = query.Where("code ILIKE ? OR name ILIKE ?", search, search)
	}

	var total int64
	if err := query.Count(&total); err != nil {
		return nil, fmt.Errorf("failed to count promo codes: %w", err)
	}

	pagination := normalizePagination(filter.Page, filter.PageSize, total)

	var codes []*promocode.PromoCode
	if err := query.Order("created_at DESC").
		Offset(pagination.Offset).
		Limit(pagination.PageSize).
		Find(&codes); err != nil {
		return nil, fmt.Errorf("failed to list promo codes: %w", err)
	}

	return &PromoCodeListResult{
		Data:       codes,
		Total:      total,
		Page:       pagination.Page,
		PageSize:   pagination.PageSize,
		TotalPages: pagination.TotalPages,
	}, nil
}

// GetPromoCode gets a promo code by ID
func (s *Service) GetPromoCode(ctx context.Context, id int64) (*promocode.PromoCode, error) {
	var code promocode.PromoCode
	if err := s.db.Model(&promocode.PromoCode{}).Where("id = ?", id).First(&code); err != nil {
		return nil, ErrPromoCodeNotFound
	}
	return &code, nil
}

// CreatePromoCode creates a new promo code
func (s *Service) CreatePromoCode(ctx context.Context, code *promocode.PromoCode, adminUserID int64) error {
	// Check if code already exists
	var existing promocode.PromoCode
	if err := s.db.Model(&promocode.PromoCode{}).Where("code = ?", code.Code).First(&existing); err == nil {
		return ErrPromoCodeAlreadyExists
	}

	if err := s.db.Create(code); err != nil {
		return fmt.Errorf("failed to create promo code: %w", err)
	}

	// Create audit log
	s.createAuditLog(ctx, adminUserID, admin.AuditActionCreate, admin.AuditTargetPromoCode, code.ID, nil, code)

	return nil
}

// UpdatePromoCode updates a promo code
func (s *Service) UpdatePromoCode(ctx context.Context, id int64, input *PromoCodeUpdateInput, adminUserID int64) (*promocode.PromoCode, error) {
	var code promocode.PromoCode
	if err := s.db.Model(&promocode.PromoCode{}).Where("id = ?", id).First(&code); err != nil {
		return nil, ErrPromoCodeNotFound
	}

	oldData := code

	if input.Name != nil {
		code.Name = *input.Name
	}
	if input.Description != nil {
		code.Description = *input.Description
	}
	if input.MaxUses != nil {
		code.MaxUses = input.MaxUses
	}
	if input.MaxUsesPerOrg != nil {
		code.MaxUsesPerOrg = *input.MaxUsesPerOrg
	}
	if input.ClearExpiresAt {
		code.ExpiresAt = nil
	} else if input.ExpiresAt != nil {
		code.ExpiresAt = input.ExpiresAt
	}

	code.UpdatedAt = time.Now()

	if err := s.db.Save(&code); err != nil {
		return nil, fmt.Errorf("failed to update promo code: %w", err)
	}

	// Create audit log
	s.createAuditLog(ctx, adminUserID, admin.AuditActionUpdate, admin.AuditTargetPromoCode, code.ID, &oldData, &code)

	return &code, nil
}

// ActivatePromoCode activates a promo code
func (s *Service) ActivatePromoCode(ctx context.Context, id int64, adminUserID int64) error {
	var code promocode.PromoCode
	if err := s.db.Model(&promocode.PromoCode{}).Where("id = ?", id).First(&code); err != nil {
		return ErrPromoCodeNotFound
	}

	oldData := code
	code.IsActive = true
	code.UpdatedAt = time.Now()

	if err := s.db.Save(&code); err != nil {
		return fmt.Errorf("failed to activate promo code: %w", err)
	}

	// Create audit log
	s.createAuditLog(ctx, adminUserID, admin.AuditActionActivate, admin.AuditTargetPromoCode, code.ID, &oldData, &code)

	return nil
}

// DeactivatePromoCode deactivates a promo code
func (s *Service) DeactivatePromoCode(ctx context.Context, id int64, adminUserID int64) error {
	var code promocode.PromoCode
	if err := s.db.Model(&promocode.PromoCode{}).Where("id = ?", id).First(&code); err != nil {
		return ErrPromoCodeNotFound
	}

	oldData := code
	code.IsActive = false
	code.UpdatedAt = time.Now()

	if err := s.db.Save(&code); err != nil {
		return fmt.Errorf("failed to deactivate promo code: %w", err)
	}

	// Create audit log
	s.createAuditLog(ctx, adminUserID, admin.AuditActionDeactivate, admin.AuditTargetPromoCode, code.ID, &oldData, &code)

	return nil
}

// DeletePromoCode deletes a promo code
func (s *Service) DeletePromoCode(ctx context.Context, id int64, adminUserID int64) error {
	var code promocode.PromoCode
	if err := s.db.Model(&promocode.PromoCode{}).Where("id = ?", id).First(&code); err != nil {
		return ErrPromoCodeNotFound
	}

	// Check if there are any redemptions
	var redemptionCount int64
	if err := s.db.Table("promo_code_redemptions").Where("promo_code_id = ?", id).Count(&redemptionCount); err != nil {
		return fmt.Errorf("failed to count redemptions: %w", err)
	}
	if redemptionCount > 0 {
		return ErrPromoCodeHasRedemptions
	}

	// Delete the promo code
	if err := s.db.Delete(&promocode.PromoCode{}, id); err != nil {
		return fmt.Errorf("failed to delete promo code: %w", err)
	}

	// Create audit log
	s.createAuditLog(ctx, adminUserID, admin.AuditActionDelete, admin.AuditTargetPromoCode, id, &code, nil)

	return nil
}

// ListPromoCodeRedemptions lists redemptions for a promo code
func (s *Service) ListPromoCodeRedemptions(ctx context.Context, promoCodeID int64, page, pageSize int) (*RedemptionListResult, error) {
	// Check if promo code exists
	var code promocode.PromoCode
	if err := s.db.Model(&promocode.PromoCode{}).Where("id = ?", promoCodeID).First(&code); err != nil {
		return nil, ErrPromoCodeNotFound
	}

	query := s.db.Table("promo_code_redemptions").Where("promo_code_id = ?", promoCodeID)

	var total int64
	if err := query.Count(&total); err != nil {
		return nil, fmt.Errorf("failed to count redemptions: %w", err)
	}

	pagination := normalizePagination(page, pageSize, total)

	var redemptions []*promocode.Redemption
	if err := s.db.Model(&promocode.Redemption{}).
		Where("promo_code_id = ?", promoCodeID).
		Order("created_at DESC").
		Offset(pagination.Offset).
		Limit(pagination.PageSize).
		Find(&redemptions); err != nil {
		return nil, fmt.Errorf("failed to list redemptions: %w", err)
	}

	// Fetch user and organization details
	result := make([]*RedemptionWithDetails, len(redemptions))
	for i, r := range redemptions {
		detail := &RedemptionWithDetails{
			ID:             r.ID,
			PromoCodeID:    r.PromoCodeID,
			OrganizationID: r.OrganizationID,
			UserID:         r.UserID,
			PlanName:       r.PlanName,
			DurationMonths: r.DurationMonths,
			NewPeriodEnd:   r.NewPeriodEnd,
			IPAddress:      r.IPAddress,
			CreatedAt:      r.CreatedAt,
		}

		// Fetch user
		var u user.User
		if err := s.db.Model(&user.User{}).Where("id = ?", r.UserID).First(&u); err == nil {
			detail.User = &u
		}

		// Fetch organization
		var org organization.Organization
		if err := s.db.Model(&organization.Organization{}).Where("id = ?", r.OrganizationID).First(&org); err == nil {
			detail.Organization = &org
		}

		result[i] = detail
	}

	return &RedemptionListResult{
		Data:       result,
		Total:      total,
		Page:       pagination.Page,
		PageSize:   pagination.PageSize,
		TotalPages: pagination.TotalPages,
	}, nil
}
