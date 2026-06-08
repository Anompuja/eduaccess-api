package application

import (
	"context"
	"errors"
	"strings"
	"time"

	academicdomain "github.com/eduaccess/eduaccess-api/internal/academic/domain"
	authdomain "github.com/eduaccess/eduaccess-api/internal/auth/domain"
	"github.com/eduaccess/eduaccess-api/internal/shared/apperror"
	"github.com/google/uuid"
)

// BulkStudentRow represents one data row parsed from the uploaded Excel file.
type BulkStudentRow struct {
	RowNumber         int
	Name              string
	Email             string
	NIS               string
	NISN              string
	Gender            string
	Religion          string
	BirthPlace        string
	BirthDate         string // YYYY-MM-DD string
	PhoneNumber       string
	Address           string
	TahunMasuk        string
	JalurMasukSekolah string
	EducationLevel    string // human-readable name, e.g. "SMA"
	ClassName         string // human-readable name, e.g. "Kelas 10"
	SubClassName      string // human-readable name, e.g. "10A"
}

// BulkRowError describes a failed row.
type BulkRowError struct {
	Row     int    `json:"row"`
	Email   string `json:"email"`
	Message string `json:"message"`
}

// BulkCreateResult summarises the outcome of a bulk import.
type BulkCreateResult struct {
	Total   int            `json:"total"`
	Created int            `json:"created"`
	Failed  int            `json:"failed"`
	Errors  []BulkRowError `json:"errors"`
}

// BulkCreateStudentCommand carries the uploaded rows and requester context.
type BulkCreateStudentCommand struct {
	RequesterSchoolID *uuid.UUID
	RequesterRole     string
	// SchoolID is the explicitly supplied school (required for superadmin via ?school_id param).
	// When set, it takes precedence over RequesterSchoolID for academic map lookups.
	SchoolID *uuid.UUID
	Rows     []BulkStudentRow
}

// BulkCreateStudentHandler processes Excel-sourced student rows one by one,
// resolving human-readable class names to UUIDs before delegating to the
// single-create handler. Failed rows are collected; the rest proceed.
type BulkCreateStudentHandler struct {
	single   *CreateStudentHandler
	academic academicdomain.AcademicRepository
}

func NewBulkCreateStudentHandler(
	single *CreateStudentHandler,
	academic academicdomain.AcademicRepository,
) *BulkCreateStudentHandler {
	return &BulkCreateStudentHandler{single: single, academic: academic}
}

func (h *BulkCreateStudentHandler) Handle(ctx context.Context, cmd BulkCreateStudentCommand) (*BulkCreateResult, error) {
	role := cmd.RequesterRole
	if role != authdomain.RoleSuperadmin &&
		role != authdomain.RoleAdminSekolah &&
		role != authdomain.RoleKepalaSekolah &&
		role != authdomain.RoleStaff {
		return nil, apperror.New(apperror.ErrForbidden,
			"only admin_sekolah, kepala_sekolah, staff, or superadmin can bulk-import students")
	}

	// Resolve the effective school: SchoolID (superadmin-supplied) takes precedence.
	effectiveSchoolID := cmd.RequesterSchoolID
	if cmd.SchoolID != nil {
		effectiveSchoolID = cmd.SchoolID
	}

	// Build name→UUID lookup maps once for the whole import.
	levelMap, classMap, subClassMap, err := h.buildAcademicMaps(ctx, effectiveSchoolID)
	if err != nil {
		return nil, err
	}

	result := &BulkCreateResult{
		Total:  len(cmd.Rows),
		Errors: []BulkRowError{},
	}

	for _, row := range cmd.Rows {
		if msg := missingBulkFields(row); msg != "" {
			result.Failed++
			result.Errors = append(result.Errors, BulkRowError{Row: row.RowNumber, Email: row.Email, Message: "missing fields: " + msg})
			continue
		}

		// Resolve class names to UUIDs.
		levelID, ok := levelMap[normalize(row.EducationLevel)]
		if !ok {
			result.Failed++
			result.Errors = append(result.Errors, BulkRowError{Row: row.RowNumber, Email: row.Email, Message: "education level not found: " + row.EducationLevel})
			continue
		}
		classID, ok := classMap[normalize(row.ClassName)]
		if !ok {
			result.Failed++
			result.Errors = append(result.Errors, BulkRowError{Row: row.RowNumber, Email: row.Email, Message: "class not found: " + row.ClassName})
			continue
		}
		subClassID, ok := subClassMap[normalize(row.SubClassName)]
		if !ok {
			result.Failed++
			result.Errors = append(result.Errors, BulkRowError{Row: row.RowNumber, Email: row.Email, Message: "sub-class not found: " + row.SubClassName})
			continue
		}

		// Parse birth date.
		var birthDate *time.Time
		if row.BirthDate != "" {
			t, parseErr := time.Parse("2006-01-02", row.BirthDate)
			if parseErr != nil {
				result.Failed++
				result.Errors = append(result.Errors, BulkRowError{Row: row.RowNumber, Email: row.Email, Message: "birth_date must be YYYY-MM-DD"})
				continue
			}
			birthDate = &t
		}

		levelUUID := levelID
		classUUID := classID
		subUUID := subClassID

		singleCmd := CreateStudentCommand{
			RequesterSchoolID: effectiveSchoolID,
			RequesterRole:     authdomain.RoleAdminSekolah, // bulk already authorised above
			Name:              row.Name,
			Email:             row.Email,
			NIS:               row.NIS,
			NISN:              row.NISN,
			Gender:            row.Gender,
			Religion:          row.Religion,
			BirthPlace:        row.BirthPlace,
			BirthDate:         birthDate,
			PhoneNumber:       row.PhoneNumber,
			Address:           row.Address,
			TahunMasuk:        row.TahunMasuk,
			JalurMasukSekolah: row.JalurMasukSekolah,
			EducationLevelID:  &levelUUID,
			ClassID:           &classUUID,
			SubClassID:        &subUUID,
		}

		if _, createErr := h.single.Handle(ctx, singleCmd); createErr != nil {
			result.Failed++
			msg := createErr.Error()
			var appErr *apperror.AppError
			if errors.As(createErr, &appErr) {
				msg = appErr.Message
			}
			result.Errors = append(result.Errors, BulkRowError{Row: row.RowNumber, Email: row.Email, Message: msg})
			continue
		}

		result.Created++
	}

	return result, nil
}

// buildAcademicMaps loads all education levels, classes, and sub-classes for the
// requester's school, returning normalised-name → UUID maps for fast lookup.
func (h *BulkCreateStudentHandler) buildAcademicMaps(
	ctx context.Context,
	schoolID *uuid.UUID,
) (levels, classes, subClasses map[string]uuid.UUID, err error) {
	levelList, err := h.academic.ListLevels(ctx, schoolID)
	if err != nil {
		return nil, nil, nil, err
	}
	classList, err := h.academic.ListClasses(ctx, schoolID, nil)
	if err != nil {
		return nil, nil, nil, err
	}
	subList, err := h.academic.ListSubClasses(ctx, schoolID, nil)
	if err != nil {
		return nil, nil, nil, err
	}

	levels = make(map[string]uuid.UUID, len(levelList))
	for _, l := range levelList {
		levels[normalize(l.Name)] = l.ID
	}
	classes = make(map[string]uuid.UUID, len(classList))
	for _, c := range classList {
		classes[normalize(c.Name)] = c.ID
	}
	subClasses = make(map[string]uuid.UUID, len(subList))
	for _, s := range subList {
		subClasses[normalize(s.Name)] = s.ID
	}
	return levels, classes, subClasses, nil
}

func missingBulkFields(row BulkStudentRow) string {
	var missing []string
	check := func(v, name string) {
		if strings.TrimSpace(v) == "" {
			missing = append(missing, name)
		}
	}
	check(row.Name, "name")
	check(row.Email, "email")
	check(row.NIS, "nis")
	check(row.NISN, "nisn")
	check(row.Gender, "gender")
	check(row.Religion, "religion")
	check(row.BirthPlace, "birth_place")
	check(row.BirthDate, "birth_date")
	check(row.PhoneNumber, "phone_number")
	check(row.Address, "address")
	check(row.TahunMasuk, "tahun_masuk")
	check(row.JalurMasukSekolah, "jalur_masuk")
	check(row.EducationLevel, "jenjang_pendidikan")
	check(row.ClassName, "kelas")
	check(row.SubClassName, "sub_kelas")
	return strings.Join(missing, ", ")
}

// normalize trims whitespace and lowercases for case-insensitive name matching.
func normalize(s string) string {
	return strings.ToLower(strings.TrimSpace(s))
}
