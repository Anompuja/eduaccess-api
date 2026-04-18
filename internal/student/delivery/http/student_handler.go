package http

import (
	"net/http"
	"strconv"
	"time"

	authmw "github.com/eduaccess/eduaccess-api/internal/shared/middleware"
	"github.com/eduaccess/eduaccess-api/internal/shared/response"
	"github.com/eduaccess/eduaccess-api/internal/shared/validator"
	"github.com/eduaccess/eduaccess-api/internal/student/application"
	"github.com/eduaccess/eduaccess-api/internal/student/domain"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

func (h *Handler) registerStudentRoutes(v1 *echo.Group, auth echo.MiddlewareFunc) {
	students := v1.Group("/students", auth)
	students.POST("", h.CreateStudent)
	students.GET("", h.ListStudents)
	students.GET("/:id", h.GetStudent)
	students.PUT("/:id", h.UpdateStudent)
	students.DELETE("/:id", h.DeactivateStudent)
	// students.POST("/:id/parents", h.LinkParent)
	// students.DELETE("/:id/parents/:parent_id", h.UnlinkParent)
}

// ── Students ──────────────────────────────────────────────────────────────────

// CreateStudent godoc
//
//	@Summary      Create student
//	@Description  Creates a user account (role=siswa) and student profile atomically.
//	@Tags         students
//	@Accept       json
//	@Produce      json
//	@Security     BearerAuth
//	@Param        body  body      CreateStudentRequest  true  "Student data"
//	@Success      201   {object}  response.Response{data=StudentResponse}
//	@Failure      400   {object}  response.Response
//	@Failure      403   {object}  response.Response
//	@Failure      409   {object}  response.Response
//	@Router       /students [post]
func (h *Handler) CreateStudent(c echo.Context) error {
	var req CreateStudentRequest
	if err := validator.BindAndValidate(c, &req); err != nil {
		return err
	}

	cmd := application.CreateStudentCommand{
		RequesterSchoolID: authmw.GetSchoolID(c),
		RequesterRole:     authmw.GetRole(c),
		Name:              req.Name,
		Email:             req.Email,
		Username:          req.Username,
		Password:          req.Password,
		NIS:               req.NIS,
		NISN:              req.NISN,
		PhoneNumber:       req.PhoneNumber,
		Address:           req.Address,
		Gender:            req.Gender,
		Religion:          req.Religion,
		BirthPlace:        req.BirthPlace,
		TahunMasuk:        req.TahunMasuk,
		JalurMasukSekolah: req.JalurMasukSekolah,
	}
	if err := parseDateField(req.BirthDate, &cmd.BirthDate); err != nil {
		return response.BadRequest(c, "birth_date must be YYYY-MM-DD")
	}
	if err := parseUUIDField(req.EducationLevelID, &cmd.EducationLevelID); err != nil {
		return response.BadRequest(c, "invalid education_level_id")
	}
	if err := parseUUIDField(req.ClassID, &cmd.ClassID); err != nil {
		return response.BadRequest(c, "invalid class_id")
	}
	if err := parseUUIDField(req.SubClassID, &cmd.SubClassID); err != nil {
		return response.BadRequest(c, "invalid sub_class_id")
	}

	student, err := h.createStudent.Handle(c.Request().Context(), cmd)
	if err != nil {
		return handleAppError(c, err)
	}
	return c.JSON(http.StatusCreated, response.Response{
		Success: true,
		Message: "student created",
		Data:    toStudentResponse(student),
	})
}

// ListStudents godoc
//
//	@Summary      List students
//	@Description  Returns a paginated list of students. Tenant-scoped.
//	@Tags         students
//	@Produce      json
//	@Security     BearerAuth
//	@Param        search            query  string  false  "Search by name, email, NIS or NISN"
//	@Param        education_level_id query string  false  "Filter by education level UUID"
//	@Param        class_id          query  string  false  "Filter by class UUID"
//	@Param        sub_class_id      query  string  false  "Filter by sub-class UUID"
//	@Param        page              query  int     false  "Page number (default 1)"
//	@Param        per_page          query  int     false  "Page size (default 20)"
//	@Success      200  {object}  response.PaginatedResponse{data=[]StudentResponse}
//	@Router       /students [get]
func (h *Handler) ListStudents(c echo.Context) error {
	page, _ := strconv.Atoi(c.QueryParam("page"))
	perPage, _ := strconv.Atoi(c.QueryParam("per_page"))

	q := application.ListStudentsQuery{
		RequesterSchoolID: authmw.GetSchoolID(c),
		RequesterRole:     authmw.GetRole(c),
		Search:            c.QueryParam("search"),
		Page:              page,
		PerPage:           perPage,
	}
	if raw := c.QueryParam("education_level_id"); raw != "" {
		if id, err := uuid.Parse(raw); err == nil {
			q.EducationLevelID = &id
		}
	}
	if raw := c.QueryParam("class_id"); raw != "" {
		if id, err := uuid.Parse(raw); err == nil {
			q.ClassID = &id
		}
	}
	if raw := c.QueryParam("sub_class_id"); raw != "" {
		if id, err := uuid.Parse(raw); err == nil {
			q.SubClassID = &id
		}
	}

	result, err := h.listStudents.Handle(c.Request().Context(), q)
	if err != nil {
		return handleAppError(c, err)
	}
	dtos := make([]StudentResponse, 0, len(result.Students))
	for _, s := range result.Students {
		dtos = append(dtos, toStudentResponse(s))
	}
	return response.Paginated(c, "students retrieved", dtos, result.Page, result.PerPage, result.Total)
}

// GetStudent godoc
//
//	@Summary      Get student by ID
//	@Description  Returns a student with parent links. Tenant-scoped.
//	@Tags         students
//	@Produce      json
//	@Security     BearerAuth
//	@Param        id   path      string  true  "Student profile UUID"
//	@Success      200  {object}  response.Response{data=StudentResponse}
//	@Failure      403  {object}  response.Response
//	@Failure      404  {object}  response.Response
//	@Router       /students/{id} [get]
func (h *Handler) GetStudent(c echo.Context) error {
	id, err := parseUUID(c, "id")
	if err != nil {
		return err
	}
	student, err := h.getStudent.Handle(c.Request().Context(), application.GetStudentQuery{
		RequesterSchoolID: authmw.GetSchoolID(c),
		RequesterRole:     authmw.GetRole(c),
		RequesterUserID:   authmw.GetUserID(c),
		StudentID:         id,
	})
	if err != nil {
		return handleAppError(c, err)
	}
	return response.OK(c, "student retrieved", toStudentResponse(student))
}

// UpdateStudent godoc
//
//	@Summary      Update student
//	@Description  Updates student profile fields.
//	@Tags         students
//	@Accept       json
//	@Produce      json
//	@Security     BearerAuth
//	@Param        id    path      string              true  "Student profile UUID"
//	@Param        body  body      UpdateStudentRequest true  "Fields to update"
//	@Success      200   {object}  response.Response{data=StudentResponse}
//	@Failure      400   {object}  response.Response
//	@Failure      403   {object}  response.Response
//	@Failure      404   {object}  response.Response
//	@Router       /students/{id} [put]
func (h *Handler) UpdateStudent(c echo.Context) error {
	id, err := parseUUID(c, "id")
	if err != nil {
		return err
	}
	var req UpdateStudentRequest
	if err := validator.BindAndValidate(c, &req); err != nil {
		return err
	}

	cmd := application.UpdateStudentCommand{
		RequesterSchoolID: authmw.GetSchoolID(c),
		RequesterRole:     authmw.GetRole(c),
		StudentID:         id,
		NIS:               req.NIS,
		NISN:              req.NISN,
		PhoneNumber:       req.PhoneNumber,
		Address:           req.Address,
		Gender:            req.Gender,
		Religion:          req.Religion,
		BirthPlace:        req.BirthPlace,
		TahunMasuk:        req.TahunMasuk,
		JalurMasukSekolah: req.JalurMasukSekolah,
	}
	if req.BirthDate != nil {
		if err := parseDateField(req.BirthDate, &cmd.BirthDate); err != nil {
			return response.BadRequest(c, "birth_date must be YYYY-MM-DD")
		}
	}
	if err := parseUUIDField(req.EducationLevelID, &cmd.EducationLevelID); err != nil {
		return response.BadRequest(c, "invalid education_level_id")
	}
	if err := parseUUIDField(req.ClassID, &cmd.ClassID); err != nil {
		return response.BadRequest(c, "invalid class_id")
	}
	if err := parseUUIDField(req.SubClassID, &cmd.SubClassID); err != nil {
		return response.BadRequest(c, "invalid sub_class_id")
	}

	student, err := h.updateStudent.Handle(c.Request().Context(), cmd)
	if err != nil {
		return handleAppError(c, err)
	}
	return response.OK(c, "student updated", toStudentResponse(student))
}

// DeactivateStudent godoc
//
//	@Summary      Deactivate student
//	@Description  Soft-deletes a student profile.
//	@Tags         students
//	@Produce      json
//	@Security     BearerAuth
//	@Param        id   path      string  true  "Student profile UUID"
//	@Success      200  {object}  response.Response
//	@Failure      403  {object}  response.Response
//	@Failure      404  {object}  response.Response
//	@Router       /students/{id} [delete]
func (h *Handler) DeactivateStudent(c echo.Context) error {
	id, err := parseUUID(c, "id")
	if err != nil {
		return err
	}
	if err := h.deactivateStudent.Handle(c.Request().Context(), application.DeactivateStudentCommand{
		RequesterSchoolID: authmw.GetSchoolID(c),
		RequesterRole:     authmw.GetRole(c),
		StudentID:         id,
	}); err != nil {
		return handleAppError(c, err)
	}
	return response.OK(c, "student deactivated", nil)
}

func toStudentResponse(s *domain.StudentProfile) StudentResponse {
	dto := StudentResponse{
		ID:                s.ID.String(),
		UserID:            s.UserID.String(),
		SchoolID:          s.SchoolID.String(),
		Name:              s.Name,
		Email:             s.Email,
		Username:          s.Username,
		Avatar:            s.Avatar,
		NIS:               s.NIS,
		NISN:              s.NISN,
		PhoneNumber:       s.PhoneNumber,
		Address:           s.Address,
		Gender:            s.Gender,
		Religion:          s.Religion,
		BirthPlace:        s.BirthPlace,
		BirthDate:         s.BirthDate,
		TahunMasuk:        s.TahunMasuk,
		JalurMasukSekolah: s.JalurMasukSekolah,
		CreatedAt:         s.CreatedAt,
		UpdatedAt:         s.UpdatedAt,
	}
	if s.EducationLevelID != nil {
		str := s.EducationLevelID.String()
		dto.EducationLevelID = &str
	}
	if s.ClassID != nil {
		str := s.ClassID.String()
		dto.ClassID = &str
	}
	if s.SubClassID != nil {
		str := s.SubClassID.String()
		dto.SubClassID = &str
	}
	if len(s.Parents) > 0 {
		links := make([]ParentLinkResponse, 0, len(s.Parents))
		for _, pl := range s.Parents {
			lr := ParentLinkResponse{
				ID:           pl.ID.String(),
				ParentID:     pl.ParentID.String(),
				Relationship: pl.Relationship,
				IsPrimary:    pl.IsPrimary,
			}
			if pl.Parent != nil {
				pr := toParentResponse(pl.Parent)
				lr.Parent = &pr
			}
			links = append(links, lr)
		}
		dto.Parents = links
	}
	return dto
}

// parseDateField parses an optional *string "YYYY-MM-DD" into *time.Time.
func parseDateField(src *string, dst **time.Time) error {
	if src == nil || *src == "" {
		return nil
	}
	t, err := time.Parse("2006-01-02", *src)
	if err != nil {
		return err
	}
	*dst = &t
	return nil
}

// parseUUIDField parses an optional *string UUID into *uuid.UUID.
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
