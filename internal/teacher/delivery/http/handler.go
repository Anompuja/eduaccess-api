package http

import (
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/eduaccess/eduaccess-api/internal/teacher/application"
	"github.com/eduaccess/eduaccess-api/internal/teacher/domain"
	"github.com/eduaccess/eduaccess-api/internal/shared/apperror"
	authmw "github.com/eduaccess/eduaccess-api/internal/shared/middleware"
	"github.com/eduaccess/eduaccess-api/internal/shared/response"
	"github.com/eduaccess/eduaccess-api/internal/shared/validator"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

// Handler wires teacher use-cases to HTTP endpoints.
type Handler struct {
	createTeacher      *application.CreateTeacherHandler
	getTeacher         *application.GetTeacherHandler
	listTeachers       *application.ListTeachersHandler
	updateTeacher      *application.UpdateTeacherHandler
	deactivateTeacher  *application.DeactivateTeacherHandler
}

// NewHandler registers teacher routes and returns the handler.
func NewHandler(
	v1 *echo.Group,
	createTeacher *application.CreateTeacherHandler,
	getTeacher *application.GetTeacherHandler,
	listTeachers *application.ListTeachersHandler,
	updateTeacher *application.UpdateTeacherHandler,
	deactivateTeacher *application.DeactivateTeacherHandler,
) *Handler {
	h := &Handler{
		createTeacher:     createTeacher,
		getTeacher:        getTeacher,
		listTeachers:      listTeachers,
		updateTeacher:     updateTeacher,
		deactivateTeacher: deactivateTeacher,
	}

	teachers := v1.Group("/teachers", authmw.RequireAuth)
	teachers.POST("", h.CreateTeacher)
	teachers.GET("", h.ListTeachers)
	teachers.GET("/:id", h.GetTeacher)
	teachers.PUT("/:id", h.UpdateTeacher)
	teachers.DELETE("/:id", h.DeactivateTeacher)

	return h
}

// GetTeacher godoc
//
//	@Summary      Get teacher by ID
//	@Tags         teachers
//	@Produce      json
//	@Security     BearerAuth
//	@Param        id   path      string  true  "Teacher profile UUID"
//	@Success      200  {object}  response.Response{data=TeacherResponse}
//	@Failure      403  {object}  response.Response
//	@Failure      404  {object}  response.Response
//	@Router       /teachers/{id} [get]
func (h *Handler) GetTeacher(c echo.Context) error {
	teacherID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return response.BadRequest(c, "invalid id")
	}

	teacher, err := h.getTeacher.Handle(c.Request().Context(), application.GetTeacherQuery{
		RequesterSchoolID: authmw.GetSchoolID(c),
		RequesterRole:     authmw.GetRole(c),
		TeacherID:         teacherID,
	})
	if err != nil {
		return handleAppError(c, err)
	}

	return response.OK(c, "teacher retrieved", toTeacherResponse(teacher))
}

// ListTeachers godoc
//
//	@Summary      List teachers
//	@Description  Returns a paginated list of teachers. Tenant-scoped.
//	@Tags         teachers
//	@Produce      json
//	@Security     BearerAuth
//	@Param        search   query  string  false  "Search by name, email or username"
//	@Param        page     query  int     false  "Page number (default 1)"
//	@Param        per_page query  int     false  "Page size (default 20)"
//	@Success      200  {object}  response.PaginatedResponse{data=[]TeacherResponse}
//	@Router       /teachers [get]
func (h *Handler) ListTeachers(c echo.Context) error {
	page, _ := strconv.Atoi(c.QueryParam("page"))
	perPage, _ := strconv.Atoi(c.QueryParam("per_page"))

	var schoolID *uuid.UUID
	if rawSchoolID := c.QueryParam("school_id"); rawSchoolID != "" {
		parsedSchoolID, err := uuid.Parse(rawSchoolID)
		if err != nil {
			return response.BadRequest(c, "invalid school_id")
		}
		schoolID = &parsedSchoolID
	}

	q := application.ListTeachersQuery{
		RequesterSchoolID: authmw.GetSchoolID(c),
		RequesterRole:     authmw.GetRole(c),
		SchoolID:          schoolID,
		Search:            c.QueryParam("search"),
		Page:              page,
		PerPage:           perPage,
	}

	result, err := h.listTeachers.Handle(c.Request().Context(), q)
	if err != nil {
		return handleAppError(c, err)
	}

	dtos := make([]TeacherResponse, 0, len(result.Teachers))
	for _, t := range result.Teachers {
		dtos = append(dtos, toTeacherResponse(t))
	}
	return response.Paginated(c, "teachers retrieved", dtos, result.Page, result.PerPage, result.Total)
}

// CreateTeacher godoc
//
//	@Summary      Create teacher
//	@Description  Creates a user account (role=teacher) and teacher profile atomically.
//	@Tags         teachers
//	@Accept       json
//	@Produce      json
//	@Security     BearerAuth
//	@Param        body  body      CreateTeacherRequest  true  "Teacher data"
//	@Success      201   {object}  response.Response{data=TeacherResponse}
//	@Failure      400   {object}  response.Response
//	@Failure      403   {object}  response.Response
//	@Failure      409   {object}  response.Response
//	@Router       /teachers [post]
func (h *Handler) CreateTeacher(c echo.Context) error {
	var req CreateTeacherRequest
	if err := validator.BindAndValidate(c, &req); err != nil {
		return err
	}

	cmd := application.CreateTeacherCommand{
		RequesterSchoolID:                   authmw.GetSchoolID(c),
		RequesterRole:                       authmw.GetRole(c),
		Name:                                req.Name,
		Email:                               req.Email,
		Username:                            req.Username,
		Password:                            req.Password,
		NIP:                                 req.NIP,
		NUPTK:                               req.NUPTK,
		PhoneNumber:                         req.PhoneNumber,
		Address:                             req.Address,
		Gender:                              req.Gender,
		Religion:                            req.Religion,
		BirthPlace:                          req.BirthPlace,
		NIK:                                 req.NIK,
		KTPImagePath:                        req.KTPImagePath,
		Kewarganegaraan:                     req.Kewarganegaraan,
		GolonganDarah:                       req.GolonganDarah,
		BeratBadan:                          req.BeratBadan,
		TinggiBadan:                         req.TinggiBadan,
		PenyakitYangSeringKambuh:            req.PenyakitYangSeringKambuh,
		KelainanJasmani:                     req.KelainanJasmani,
		PenyakitKronisYangPernahDiderita:    req.PenyakitKronisYangPernahDiderita,
		RTRW:                                req.RTRW,
		KodePos:                             req.KodePos,
		PendidikanTerakhir:                  req.PendidikanTerakhir,
		Jurusan:                             req.Jurusan,
		TahunLulus:                          req.TahunLulus,
		TahunMasuk:                          req.TahunMasuk,
	}

	if err := parseDateField(req.BirthDate, &cmd.BirthDate); err != nil {
		return response.BadRequest(c, "birth_date must be YYYY-MM-DD")
	}
	if err := parseUUIDField(req.SchoolID, &cmd.SchoolID); err != nil {
		return response.BadRequest(c, "invalid school_id")
	}

	teacher, err := h.createTeacher.Handle(c.Request().Context(), cmd)
	if err != nil {
		return handleAppError(c, err)
	}

	return c.JSON(http.StatusCreated, response.Response{
		Success: true,
		Message: "teacher created",
		Data:    toTeacherResponse(teacher),
	})
}

// UpdateTeacher godoc
//
//	@Summary      Update teacher
//	@Description  Updates mutable fields in users and teacher_profiles.
//	@Tags         teachers
//	@Accept       json
//	@Produce      json
//	@Security     BearerAuth
//	@Param        id    path      string              true  "Teacher profile UUID"
//	@Param        body  body      UpdateTeacherRequest  true  "Teacher fields to update"
//	@Success      200   {object}  response.Response{data=TeacherResponse}
//	@Failure      400   {object}  response.Response
//	@Failure      403   {object}  response.Response
//	@Failure      404   {object}  response.Response
//	@Failure      409   {object}  response.Response
//	@Router       /teachers/{id} [put]
func (h *Handler) UpdateTeacher(c echo.Context) error {
	teacherID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return response.BadRequest(c, "invalid id")
	}

	var req UpdateTeacherRequest
	if err := validator.BindAndValidate(c, &req); err != nil {
		return err
	}

	cmd := application.UpdateTeacherCommand{
		RequesterSchoolID:                   authmw.GetSchoolID(c),
		RequesterRole:                       authmw.GetRole(c),
		TeacherID:                           teacherID,
		Name:                                req.Name,
		Email:                               req.Email,
		Username:                            req.Username,
		PhoneNumber:                         req.PhoneNumber,
		Address:                             req.Address,
		Gender:                              req.Gender,
		Religion:                            req.Religion,
		BirthPlace:                          req.BirthPlace,
		NIK:                                 req.NIK,
		KTPImagePath:                        req.KTPImagePath,
		NIP:                                 req.NIP,
		NUPTK:                               req.NUPTK,
		Kewarganegaraan:                     req.Kewarganegaraan,
		GolonganDarah:                       req.GolonganDarah,
		BeratBadan:                          req.BeratBadan,
		TinggiBadan:                         req.TinggiBadan,
		PenyakitYangSeringKambuh:            req.PenyakitYangSeringKambuh,
		KelainanJasmani:                     req.KelainanJasmani,
		PenyakitKronisYangPernahDiderita:    req.PenyakitKronisYangPernahDiderita,
		RTRW:                                req.RTRW,
		KodePos:                             req.KodePos,
		PendidikanTerakhir:                  req.PendidikanTerakhir,
		Jurusan:                             req.Jurusan,
		TahunLulus:                          req.TahunLulus,
		TahunMasuk:                          req.TahunMasuk,
	}

	if err := parseDateField(req.BirthDate, &cmd.BirthDate); err != nil {
		return response.BadRequest(c, "birth_date must be YYYY-MM-DD")
	}

	teacher, err := h.updateTeacher.Handle(c.Request().Context(), cmd)
	if err != nil {
		return handleAppError(c, err)
	}

	return response.OK(c, "teacher updated", toTeacherResponse(teacher))
}

// DeactivateTeacher godoc
//
//	@Summary      Deactivate teacher
//	@Tags         teachers
//	@Produce      json
//	@Security     BearerAuth
//	@Param        id   path      string  true  "Teacher profile UUID"
//	@Success      200  {object}  response.Response
//	@Failure      403  {object}  response.Response
//	@Failure      404  {object}  response.Response
//	@Router       /teachers/{id} [delete]
func (h *Handler) DeactivateTeacher(c echo.Context) error {
	teacherID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return response.BadRequest(c, "invalid id")
	}

	if err := h.deactivateTeacher.Handle(c.Request().Context(), application.DeactivateTeacherCommand{
		RequesterSchoolID: authmw.GetSchoolID(c),
		RequesterRole:     authmw.GetRole(c),
		TeacherID:         teacherID,
	}); err != nil {
		return handleAppError(c, err)
	}

	return response.OK(c, "teacher deactivated", nil)
}

func toTeacherResponse(t *domain.TeacherProfile) TeacherResponse {
	var birthDate *string
	if formatted := formatDatePtr(t.BirthDate); formatted != nil {
		birthDate = formatted
	}

	return TeacherResponse{
		ID:                              t.ID.String(),
		UserID:                          t.UserID.String(),
		SchoolID:                        t.SchoolID.String(),
		Name:                            t.Name,
		Email:                           t.Email,
		Username:                        t.Username,
		Avatar:                          t.Avatar,
		NIP:                             t.NIP,
		NUPTK:                           t.NUPTK,
		PhoneNumber:                     t.PhoneNumber,
		Address:                         t.Address,
		Gender:                          t.Gender,
		Religion:                        t.Religion,
		BirthPlace:                      t.BirthPlace,
		BirthDate:                       birthDate,
		NIK:                             t.NIK,
		KTPImagePath:                    t.KTPImagePath,
		Kewarganegaraan:                 t.Kewarganegaraan,
		GolonganDarah:                   t.GolonganDarah,
		BeratBadan:                      t.BeratBadan,
		TinggiBadan:                     t.TinggiBadan,
		PenyakitYangSeringKambuh:        t.PenyakitYangSeringKambuh,
		KelainanJasmani:                 t.KelainanJasmani,
		PenyakitKronisYangPernahDiderita: t.PenyakitKronisYangPernahDiderita,
		RTRW:                            t.RTRW,
		KodePos:                         t.KodePos,
		PendidikanTerakhir:              t.PendidikanTerakhir,
		Jurusan:                         t.Jurusan,
		TahunLulus:                      t.TahunLulus,
		TahunMasuk:                      t.TahunMasuk,
		CreatedAt:                       t.CreatedAt,
		UpdatedAt:                       t.UpdatedAt,
	}
}

func formatDatePtr(t *time.Time) *string {
	if t == nil {
		return nil
	}
	if y := t.Year(); y < 1 || y > 9999 {
		return nil
	}
	formatted := t.Format("2006-01-02")
	return &formatted
}

// parseDateField parses an optional *string "YYYY-MM-DD" into *string.
func parseDateField(src *string, dst **string) error {
	if src == nil || *src == "" {
		return nil
	}
	_, err := time.Parse("2006-01-02", *src)
	if err != nil {
		return err
	}
	*dst = src
	return nil
}

func parseUUIDField(src *string, dst **uuid.UUID) error {
	if src == nil || *src == "" {
		return nil
	}
	id, err := uuid.Parse(*src)
	if err != nil {
		return err
	}
	*dst = &id
	return nil
}

func handleAppError(c echo.Context, err error) error {
	var appErr *apperror.AppError
	if errors.As(err, &appErr) {
		switch appErr.Err {
		case apperror.ErrNotFound:
			return response.NotFound(c, appErr.Message)
		case apperror.ErrUnauthorized, apperror.ErrInvalidToken, apperror.ErrTokenRevoked:
			return response.Unauthorized(c, appErr.Message)
		case apperror.ErrForbidden:
			return response.Forbidden(c, appErr.Message)
		case apperror.ErrConflict:
			return response.Conflict(c, appErr.Message)
		case apperror.ErrBadRequest, apperror.ErrWrongPassword:
			return response.BadRequest(c, appErr.Message)
		}
	}

	return c.JSON(http.StatusInternalServerError, response.Response{
		Success: false,
		Message: "internal server error",
	})
}
