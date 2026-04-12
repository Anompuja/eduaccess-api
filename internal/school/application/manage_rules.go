package application

import (
	"context"
	"time"

	"github.com/eduaccess/eduaccess-api/internal/school/domain"
	"github.com/eduaccess/eduaccess-api/internal/shared/apperror"
	"github.com/google/uuid"
)

// ── List Rules ────────────────────────────────────────────────────────────────

// ListRulesQuery lists school rules for a school.
type ListRulesQuery struct {
	RequesterSchoolID *uuid.UUID
	RequesterRole     string
	SchoolID          uuid.UUID
}

// ListRulesHandler returns all rules for a school.
type ListRulesHandler struct {
	repo domain.SchoolRepository
}

func NewListRulesHandler(repo domain.SchoolRepository) *ListRulesHandler {
	return &ListRulesHandler{repo: repo}
}

func (h *ListRulesHandler) Handle(ctx context.Context, q ListRulesQuery) ([]*domain.SchoolRule, error) {
	if err := guardSchoolAccess(q.RequesterRole, q.RequesterSchoolID, q.SchoolID); err != nil {
		return nil, err
	}
	return h.repo.ListRules(ctx, q.SchoolID)
}

// ── Upsert Rules ──────────────────────────────────────────────────────────────

// RuleInput is a single key-value pair to upsert.
type RuleInput struct {
	Key   string
	Value string
	Note  string
}

// UpsertRulesCommand upserts multiple rules for a school.
type UpsertRulesCommand struct {
	RequesterSchoolID *uuid.UUID
	RequesterRole     string
	SchoolID          uuid.UUID
	Rules             []RuleInput
}

// UpsertRulesHandler creates or updates school rules.
type UpsertRulesHandler struct {
	repo domain.SchoolRepository
}

func NewUpsertRulesHandler(repo domain.SchoolRepository) *UpsertRulesHandler {
	return &UpsertRulesHandler{repo: repo}
}

func (h *UpsertRulesHandler) Handle(ctx context.Context, cmd UpsertRulesCommand) ([]*domain.SchoolRule, error) {
	if cmd.RequesterRole != "superadmin" && cmd.RequesterRole != "admin_sekolah" {
		return nil, apperror.New(apperror.ErrForbidden, "only admin_sekolah or superadmin can manage school rules")
	}
	if err := guardSchoolAccess(cmd.RequesterRole, cmd.RequesterSchoolID, cmd.SchoolID); err != nil {
		return nil, err
	}

	now := time.Now()
	for _, ri := range cmd.Rules {
		rule := &domain.SchoolRule{
			ID:        uuid.New(),
			SchoolID:  cmd.SchoolID,
			Key:       ri.Key,
			Value:     ri.Value,
			Note:      ri.Note,
			CreatedAt: now,
			UpdatedAt: now,
		}
		if err := h.repo.UpsertRule(ctx, rule); err != nil {
			return nil, err
		}
	}

	return h.repo.ListRules(ctx, cmd.SchoolID)
}

// ── helper ────────────────────────────────────────────────────────────────────

func guardSchoolAccess(role string, requesterSchoolID *uuid.UUID, targetSchoolID uuid.UUID) error {
	if role == "superadmin" {
		return nil
	}
	if requesterSchoolID == nil || *requesterSchoolID != targetSchoolID {
		return apperror.New(apperror.ErrForbidden, "access denied to this school")
	}
	return nil
}
