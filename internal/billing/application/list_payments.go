package application

import (
	"context"

	authdomain "github.com/eduaccess/eduaccess-api/internal/auth/domain"
	billingdomain "github.com/eduaccess/eduaccess-api/internal/billing/domain"
	"github.com/eduaccess/eduaccess-api/internal/shared/apperror"
	"github.com/google/uuid"
)

type ListPaymentsQuery struct {
	RequesterRole     string
	RequesterSchoolID *uuid.UUID
	SchoolID          *uuid.UUID
	Status            string
	Search            string
	Page              int
	PerPage           int
}

type ListPaymentsResult struct {
	Payments []*billingdomain.PaymentTransaction
	Page     int
	PerPage  int
	Total    int64
}

type ListPaymentsHandler struct {
	payments billingdomain.PaymentRepository
}

func NewListPaymentsHandler(payments billingdomain.PaymentRepository) *ListPaymentsHandler {
	return &ListPaymentsHandler{payments: payments}
}

func (h *ListPaymentsHandler) Handle(ctx context.Context, q ListPaymentsQuery) (*ListPaymentsResult, error) {
	var schoolID *uuid.UUID

	switch q.RequesterRole {
	case authdomain.RoleSuperadmin:
		schoolID = q.SchoolID
	case authdomain.RoleAdminSekolah:
		if q.RequesterSchoolID == nil {
			return nil, apperror.New(apperror.ErrForbidden, "school context required")
		}
		schoolID = q.RequesterSchoolID
	default:
		return nil, apperror.New(apperror.ErrForbidden, "only admin_sekolah or superadmin can view subscription payments")
	}

	page := q.Page
	if page < 1 {
		page = 1
	}
	perPage := q.PerPage
	if perPage < 1 || perPage > 100 {
		perPage = 20
	}

	payments, total, err := h.payments.List(ctx, billingdomain.PaymentFilter{
		SchoolID: schoolID,
		Status:   q.Status,
		Search:   q.Search,
		Offset:   (page - 1) * perPage,
		Limit:    perPage,
	})
	if err != nil {
		return nil, err
	}

	return &ListPaymentsResult{
		Payments: payments,
		Page:     page,
		PerPage:  perPage,
		Total:    total,
	}, nil
}
