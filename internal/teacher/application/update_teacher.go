package application

import (
	"context"
	"time"

	"github.com/eduaccess/eduaccess-api/internal/shared/apperror"
	"github.com/eduaccess/eduaccess-api/internal/teacher/domain"
	authdomain "github.com/eduaccess/eduaccess-api/internal/auth/domain"
	"github.com/google/uuid"
)

// UpdateTeacherCommand represents the command to update a teacher profile.
type UpdateTeacherCommand struct {
	RequesterSchoolID   *uuid.UUID
	RequesterRole       string
	TeacherID           uuid.UUID
	Name                *string
	Email               *string
	Username            *string
	PhoneNumber         *string
	Address             *string
	Gender              *string
	Religion            *string
	BirthPlace          *string
	BirthDate           *string
	NIK                 *string
	KTPImagePath        *string
	NIP                 *string
	NUPTK               *string
	Kewarganegaraan     *string
	GolonganDarah       *string
	BeratBadan          *string
	TinggiBadan         *string
	PenyakitYangSeringKambuh *string
	KelainanJasmani     *string
	PenyakitKronisYangPernahDiderita *string
	RTRW                *string
	KodePos             *string
	PendidikanTerakhir  *string
	Jurusan             *string
	TahunLulus          *string
	TahunMasuk          *string
}

// UpdateTeacherHandler handles updating a teacher profile.
type UpdateTeacherHandler struct {
	teacherRepo domain.TeacherRepository
	userUpdater UserUpdater
}

// NewUpdateTeacherHandler creates a new UpdateTeacherHandler.
func NewUpdateTeacherHandler(
	teacherRepo domain.TeacherRepository,
	userUpdater UserUpdater,
) *UpdateTeacherHandler {
	return &UpdateTeacherHandler{
		teacherRepo: teacherRepo,
		userUpdater: userUpdater,
	}
}

// Handle updates a teacher profile with authorization checks.
func (h *UpdateTeacherHandler) Handle(ctx context.Context, cmd UpdateTeacherCommand) (*domain.TeacherProfile, error) {
	if cmd.RequesterRole != authdomain.RoleSuperadmin && cmd.RequesterRole != authdomain.RoleAdminSekolah {
		return nil, apperror.New(apperror.ErrForbidden, "only admin_sekolah or superadmin can update teacher")
	}

	teacher, err := h.teacherRepo.FindTeacherByID(ctx, cmd.TeacherID)
	if err != nil {
		return nil, apperror.New(apperror.ErrNotFound, "teacher not found")
	}

	if cmd.RequesterRole != authdomain.RoleSuperadmin {
		if cmd.RequesterSchoolID == nil || teacher.SchoolID != *cmd.RequesterSchoolID {
			return nil, apperror.New(apperror.ErrForbidden, "access denied to this teacher")
		}
	}

	user, err := h.userUpdater.FindByID(ctx, teacher.UserID)
	if err != nil {
		return nil, apperror.New(apperror.ErrNotFound, "user not found")
	}

	// Check email uniqueness if updating
	if cmd.Email != nil && user.Email != *cmd.Email {
		exists, err := h.userUpdater.ExistsByEmail(ctx, *cmd.Email)
		if err != nil {
			return nil, err
		}
		if exists {
			return nil, apperror.New(apperror.ErrConflict, "email already in use")
		}
		user.Email = *cmd.Email
		teacher.Email = *cmd.Email
	}

	// Check username uniqueness if updating
	if cmd.Username != nil && user.Username != *cmd.Username {
		exists, err := h.userUpdater.ExistsByUsername(ctx, *cmd.Username)
		if err != nil {
			return nil, err
		}
		if exists {
			return nil, apperror.New(apperror.ErrConflict, "username already in use")
		}
		user.Username = *cmd.Username
		teacher.Username = *cmd.Username
	}

	// Update user account fields
	if cmd.Name != nil {
		user.Name = *cmd.Name
		teacher.Name = *cmd.Name
	}

	// Update teacher profile fields
	if cmd.PhoneNumber != nil {
		teacher.PhoneNumber = cmd.PhoneNumber
	}
	if cmd.Address != nil {
		teacher.Address = cmd.Address
	}
	if cmd.Gender != nil {
		teacher.Gender = cmd.Gender
	}
	if cmd.Religion != nil {
		teacher.Religion = cmd.Religion
	}
	if cmd.BirthPlace != nil {
		teacher.BirthPlace = cmd.BirthPlace
	}
	if cmd.BirthDate != nil && *cmd.BirthDate != "" {
		t, err := time.Parse("2006-01-02", *cmd.BirthDate)
		if err != nil {
			return nil, apperror.New(apperror.ErrBadRequest, "birth_date must be YYYY-MM-DD")
		}
		teacher.BirthDate = &t
	}
	if cmd.NIK != nil {
		teacher.NIK = cmd.NIK
	}
	if cmd.KTPImagePath != nil {
		teacher.KTPImagePath = cmd.KTPImagePath
	}
	if cmd.NIP != nil {
		teacher.NIP = cmd.NIP
	}
	if cmd.NUPTK != nil {
		teacher.NUPTK = cmd.NUPTK
	}
	if cmd.Kewarganegaraan != nil {
		teacher.Kewarganegaraan = cmd.Kewarganegaraan
	}
	if cmd.GolonganDarah != nil {
		teacher.GolonganDarah = cmd.GolonganDarah
	}
	if cmd.BeratBadan != nil {
		teacher.BeratBadan = cmd.BeratBadan
	}
	if cmd.TinggiBadan != nil {
		teacher.TinggiBadan = cmd.TinggiBadan
	}
	if cmd.PenyakitYangSeringKambuh != nil {
		teacher.PenyakitYangSeringKambuh = cmd.PenyakitYangSeringKambuh
	}
	if cmd.KelainanJasmani != nil {
		teacher.KelainanJasmani = cmd.KelainanJasmani
	}
	if cmd.PenyakitKronisYangPernahDiderita != nil {
		teacher.PenyakitKronisYangPernahDiderita = cmd.PenyakitKronisYangPernahDiderita
	}
	if cmd.RTRW != nil {
		teacher.RTRW = cmd.RTRW
	}
	if cmd.KodePos != nil {
		teacher.KodePos = cmd.KodePos
	}
	if cmd.PendidikanTerakhir != nil {
		teacher.PendidikanTerakhir = cmd.PendidikanTerakhir
	}
	if cmd.Jurusan != nil {
		teacher.Jurusan = cmd.Jurusan
	}
	if cmd.TahunLulus != nil {
		teacher.TahunLulus = cmd.TahunLulus
	}
	if cmd.TahunMasuk != nil {
		teacher.TahunMasuk = cmd.TahunMasuk
	}

	now := time.Now()
	user.UpdatedAt = now
	teacher.UpdatedAt = now

	if err := h.userUpdater.Update(ctx, user); err != nil {
		return nil, err
	}
	if err := h.teacherRepo.UpdateTeacherProfile(ctx, teacher); err != nil {
		return nil, err
	}

	return teacher, nil
}
