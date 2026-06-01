package application

import (
	"context"
	"time"

	"github.com/eduaccess/eduaccess-api/internal/shared/apperror"
	"github.com/eduaccess/eduaccess-api/internal/student/domain"
	"github.com/google/uuid"
)

// UpdateStudentCommand holds mutable student profile fields.
type UpdateStudentCommand struct {
	RequesterSchoolID *uuid.UUID
	RequesterRole     string
	StudentID         uuid.UUID
	NIS               *string
	NISN              *string
	PhoneNumber       *string
	Address           *string
	Gender            *string
	Religion          *string
	BirthPlace        *string
	BirthDate         *time.Time
	TahunMasuk        *string
	JalurMasukSekolah *string
	EducationLevelID  *uuid.UUID
	ClassID           *uuid.UUID
	SubClassID        *uuid.UUID
}

// UpdateStudentHandler updates mutable student profile fields.
type UpdateStudentHandler struct {
	repo domain.StudentRepository
}

func NewUpdateStudentHandler(repo domain.StudentRepository) *UpdateStudentHandler {
	return &UpdateStudentHandler{repo: repo}
}

func (h *UpdateStudentHandler) Handle(ctx context.Context, cmd UpdateStudentCommand) (*domain.StudentProfile, error) {
	if cmd.RequesterRole != "superadmin" && cmd.RequesterRole != "admin_sekolah" {
		return nil, apperror.New(apperror.ErrForbidden, "only admin_sekolah or superadmin can update student profiles")
	}

	profile, err := h.repo.FindStudentByID(ctx, cmd.StudentID)
	if err != nil {
		return nil, err
	}

	if cmd.RequesterRole != "superadmin" {
		if cmd.RequesterSchoolID != nil && profile.SchoolID != *cmd.RequesterSchoolID {
			return nil, apperror.New(apperror.ErrForbidden, "access denied to this student")
		}
	}

	if cmd.NIS != nil {
		profile.NIS = *cmd.NIS
	}
	if cmd.NISN != nil {
		profile.NISN = *cmd.NISN
	}
	if cmd.PhoneNumber != nil {
		profile.PhoneNumber = *cmd.PhoneNumber
	}
	if cmd.Address != nil {
		profile.Address = *cmd.Address
	}
	if cmd.Gender != nil {
		profile.Gender = *cmd.Gender
	}
	if cmd.Religion != nil {
		profile.Religion = *cmd.Religion
	}
	if cmd.BirthPlace != nil {
		profile.BirthPlace = *cmd.BirthPlace
	}
	if cmd.BirthDate != nil {
		profile.BirthDate = cmd.BirthDate
	}
	if cmd.TahunMasuk != nil {
		profile.TahunMasuk = *cmd.TahunMasuk
	}
	if cmd.JalurMasukSekolah != nil {
		profile.JalurMasukSekolah = *cmd.JalurMasukSekolah
	}
	if cmd.EducationLevelID != nil {
		profile.EducationLevelID = cmd.EducationLevelID
	}
	if cmd.ClassID != nil {
		profile.ClassID = cmd.ClassID
	}
	if cmd.SubClassID != nil {
		profile.SubClassID = cmd.SubClassID
	}

	profile.UpdatedAt = time.Now()
	if err := h.repo.UpdateStudentProfile(ctx, profile); err != nil {
		return nil, err
	}
	return profile, nil
}
