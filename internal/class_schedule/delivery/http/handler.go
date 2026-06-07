package http

import (
	"errors"
	"net/http"
	"time"

	"github.com/eduaccess/eduaccess-api/internal/class_schedule/application"
	"github.com/eduaccess/eduaccess-api/internal/class_schedule/domain"
	"github.com/eduaccess/eduaccess-api/internal/shared/apperror"
	"github.com/eduaccess/eduaccess-api/internal/shared/httpcache"
	authmw "github.com/eduaccess/eduaccess-api/internal/shared/middleware"
	"github.com/eduaccess/eduaccess-api/internal/shared/response"
	"github.com/eduaccess/eduaccess-api/internal/shared/validator"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

type Handler struct {
	createSchedule   *application.CreateClassScheduleHandler
	listSchedules    *application.ListClassSchedulesHandler
	getSchedule      *application.GetClassScheduleHandler
	updateSchedule   *application.UpdateClassScheduleHandler
	deleteSchedule   *application.DeleteClassScheduleHandler
	startSchedule    *application.StartClassScheduleHandler
	completeSchedule *application.CompleteClassScheduleHandler
	cancelSchedule   *application.CancelClassScheduleHandler
	syncStudents     *application.SyncStudentsHandler
	listAttendances  *application.ListAttendancesHandler
	updateAttendance *application.UpdateAttendanceHandler
	generateQR       *application.GenerateQRHandler
	scanQR           *application.ScanQRHandler
}

func NewHandler(
	v1 *echo.Group,
	createSchedule *application.CreateClassScheduleHandler,
	listSchedules *application.ListClassSchedulesHandler,
	getSchedule *application.GetClassScheduleHandler,
	updateSchedule *application.UpdateClassScheduleHandler,
	deleteSchedule *application.DeleteClassScheduleHandler,
	startSchedule *application.StartClassScheduleHandler,
	completeSchedule *application.CompleteClassScheduleHandler,
	cancelSchedule *application.CancelClassScheduleHandler,
	syncStudents *application.SyncStudentsHandler,
	listAttendances *application.ListAttendancesHandler,
	updateAttendance *application.UpdateAttendanceHandler,
	generateQR *application.GenerateQRHandler,
	scanQR *application.ScanQRHandler,
) *Handler {
	h := &Handler{
		createSchedule: createSchedule, listSchedules: listSchedules,
		getSchedule: getSchedule, updateSchedule: updateSchedule,
		deleteSchedule: deleteSchedule, startSchedule: startSchedule,
		completeSchedule: completeSchedule, cancelSchedule: cancelSchedule,
		syncStudents: syncStudents, listAttendances: listAttendances,
		updateAttendance: updateAttendance,
		generateQR:       generateQR,
		scanQR:           scanQR,
	}

	auth := authmw.RequireAuth
	cs := v1.Group("/class-schedules", auth, httpcache.Middleware(httpcache.AlwaysRevalidate))
	cs.POST("", h.CreateClassSchedule)
	cs.GET("", h.ListClassSchedules)
	// QR routes registered before /:id to prevent "qr" matching as an ID param.
	cs.GET("/:id/qr", h.GetQRToken)
	cs.GET("/:id/qr/image", h.GetQRImage)
	cs.GET("/:id", h.GetClassSchedule)
	cs.PUT("/:id", h.UpdateClassSchedule)
	cs.DELETE("/:id", h.DeleteClassSchedule)
	cs.PATCH("/:id/start", h.StartClassSchedule)
	cs.PATCH("/:id/complete", h.CompleteClassSchedule)
	cs.PATCH("/:id/cancel", h.CancelClassSchedule)
	cs.PATCH("/:id/sync-students", h.SyncStudents)
	cs.GET("/:id/attendances", h.ListAttendances)
	cs.PUT("/:id/attendances/:student_id", h.UpdateAttendance)

	// Scan endpoint is at /api/v1/attendance/scan (student-facing, not under /class-schedules).
	v1.POST("/attendance/scan", h.ScanQR, auth)

	return h
}

func (h *Handler) CreateClassSchedule(c echo.Context) error {
	var req CreateClassScheduleRequest
	if err := validator.BindAndValidate(c, &req); err != nil {
		return err
	}
	classroomID, _ := uuid.Parse(req.ClassroomID)
	subjectID, _ := uuid.Parse(req.SubjectID)
	teacherID, _ := uuid.Parse(req.TeacherID)
	date, err := time.Parse("2006-01-02", req.Date)
	if err != nil {
		return response.BadRequest(c, "invalid date format, use YYYY-MM-DD")
	}
	cs, err := h.createSchedule.Handle(c.Request().Context(), application.CreateClassScheduleCommand{
		RequesterSchoolID: getSchoolID(c), RequesterRole: authmw.GetRole(c),
		ClassroomID:   classroomID,
		SubjectID:     subjectID,
		TeacherID:     teacherID,
		StartPeriodID: parseOptionalUUID(req.StartPeriodID),
		EndPeriodID:   parseOptionalUUID(req.EndPeriodID),
		Date:          date, StartTime: req.StartTime, EndTime: req.EndTime,
	})
	if err != nil {
		return handleAppError(c, err)
	}
	return c.JSON(http.StatusCreated, response.Response{Success: true, Message: "class schedule created", Data: toScheduleResponse(cs)})
}

// ListClassSchedules godoc
//
//	@Summary      List class schedules
//	@Description  Returns class schedules filtered by the request school context. Superadmin may provide school_id to target one school.
//	@Tags         class-schedules
//	@Produce      json
//	@Security     BearerAuth
//	@Param        school_id    query  string  false  "School UUID (superadmin only)"
//	@Param        classroom_id query  string  false  "Filter by classroom UUID"
//	@Param        teacher_id   query  string  false  "Filter by teacher UUID"
//	@Param        subject_id   query  string  false  "Filter by subject UUID"
//	@Param        date         query  string  false  "Filter by date (YYYY-MM-DD)"
//	@Param        status       query  string  false  "Filter by schedule status"
//	@Success      200          {object}  response.Response{data=[]ClassScheduleResponse}
//	@Failure      400          {object}  response.Response
//	@Failure      403          {object}  response.Response
//	@Router       /class-schedules [get]
func (h *Handler) ListClassSchedules(c echo.Context) error {
	role := authmw.GetRole(c)
	q := application.ListClassSchedulesQuery{
		RequesterSchoolID: getSchoolID(c), RequesterRole: role,
	}
	// Teachers always see only their own schedules.
	if role == "guru" {
		userID := authmw.GetUserID(c)
		q.TeacherID = &userID
	}
	if raw := c.QueryParam("classroom_id"); raw != "" {
		if id, err := uuid.Parse(raw); err == nil {
			q.ClassroomID = &id
		}
	}
	// Only non-guru roles may override teacher_id via query param.
	if role != "guru" {
		if raw := c.QueryParam("teacher_id"); raw != "" {
			if id, err := uuid.Parse(raw); err == nil {
				q.TeacherID = &id
			}
		}
	}
	if raw := c.QueryParam("subject_id"); raw != "" {
		if id, err := uuid.Parse(raw); err == nil {
			q.SubjectID = &id
		}
	}
	if raw := c.QueryParam("date"); raw != "" {
		if d, err := time.Parse("2006-01-02", raw); err == nil {
			q.Date = &d
		}
	}
	if raw := c.QueryParam("status"); raw != "" {
		q.Status = &raw
	}
	list, err := h.listSchedules.Handle(c.Request().Context(), q)
	if err != nil {
		return handleAppError(c, err)
	}
	dtos := make([]ClassScheduleResponse, 0, len(list))
	for _, cs := range list {
		dtos = append(dtos, toScheduleResponse(cs))
	}
	return response.OK(c, "class schedules retrieved", dtos)
}

func (h *Handler) GetClassSchedule(c echo.Context) error {
	id, err := parseUUID(c, "id")
	if err != nil {
		return handleAppError(c, err)
	}
	cs, err := h.getSchedule.Handle(c.Request().Context(), application.GetClassScheduleQuery{
		RequesterSchoolID: getSchoolID(c), RequesterRole: authmw.GetRole(c), ScheduleID: id,
	})
	if err != nil {
		return handleAppError(c, err)
	}
	return response.OK(c, "class schedule retrieved", toScheduleResponse(cs))
}

func (h *Handler) UpdateClassSchedule(c echo.Context) error {
	id, err := parseUUID(c, "id")
	if err != nil {
		return handleAppError(c, err)
	}
	var req UpdateClassScheduleRequest
	if err := validator.BindAndValidate(c, &req); err != nil {
		return err
	}
	classroomID, _ := uuid.Parse(req.ClassroomID)
	subjectID, _ := uuid.Parse(req.SubjectID)
	teacherID, _ := uuid.Parse(req.TeacherID)
	date, err := time.Parse("2006-01-02", req.Date)
	if err != nil {
		return response.BadRequest(c, "invalid date format, use YYYY-MM-DD")
	}
	cs, err := h.updateSchedule.Handle(c.Request().Context(), application.UpdateClassScheduleCommand{
		RequesterSchoolID: getSchoolID(c), RequesterRole: authmw.GetRole(c), ScheduleID: id,
		ClassroomID:   classroomID,
		SubjectID:     subjectID,
		TeacherID:     teacherID,
		StartPeriodID: parseOptionalUUID(req.StartPeriodID),
		EndPeriodID:   parseOptionalUUID(req.EndPeriodID),
		Date:          date, StartTime: req.StartTime, EndTime: req.EndTime,
	})
	if err != nil {
		return handleAppError(c, err)
	}
	return response.OK(c, "class schedule updated", toScheduleResponse(cs))
}

func (h *Handler) DeleteClassSchedule(c echo.Context) error {
	id, err := parseUUID(c, "id")
	if err != nil {
		return handleAppError(c, err)
	}
	if err := h.deleteSchedule.Handle(c.Request().Context(), application.DeleteClassScheduleCommand{
		RequesterSchoolID: getSchoolID(c), RequesterRole: authmw.GetRole(c), ScheduleID: id,
	}); err != nil {
		return handleAppError(c, err)
	}
	return response.OK(c, "class schedule deleted", nil)
}

func (h *Handler) StartClassSchedule(c echo.Context) error {
	id, err := parseUUID(c, "id")
	if err != nil {
		return handleAppError(c, err)
	}
	if err := h.startSchedule.Handle(c.Request().Context(), application.StartClassScheduleCommand{
		RequesterSchoolID: getSchoolID(c), RequesterRole: authmw.GetRole(c), ScheduleID: id,
	}); err != nil {
		return handleAppError(c, err)
	}
	return response.OK(c, "class started", nil)
}

func (h *Handler) CompleteClassSchedule(c echo.Context) error {
	id, err := parseUUID(c, "id")
	if err != nil {
		return handleAppError(c, err)
	}
	if err := h.completeSchedule.Handle(c.Request().Context(), application.CompleteClassScheduleCommand{
		RequesterSchoolID: getSchoolID(c), RequesterRole: authmw.GetRole(c), ScheduleID: id,
	}); err != nil {
		return handleAppError(c, err)
	}
	return response.OK(c, "class completed", nil)
}

func (h *Handler) CancelClassSchedule(c echo.Context) error {
	id, err := parseUUID(c, "id")
	if err != nil {
		return handleAppError(c, err)
	}
	if err := h.cancelSchedule.Handle(c.Request().Context(), application.CancelClassScheduleCommand{
		RequesterSchoolID: getSchoolID(c), RequesterRole: authmw.GetRole(c), ScheduleID: id,
	}); err != nil {
		return handleAppError(c, err)
	}
	return response.OK(c, "class cancelled", nil)
}

func (h *Handler) SyncStudents(c echo.Context) error {
	id, err := parseUUID(c, "id")
	if err != nil {
		return handleAppError(c, err)
	}
	if err := h.syncStudents.Handle(c.Request().Context(), application.SyncStudentsCommand{
		RequesterSchoolID: getSchoolID(c), RequesterRole: authmw.GetRole(c), ScheduleID: id,
	}); err != nil {
		return handleAppError(c, err)
	}
	return response.OK(c, "students synced", nil)
}

func (h *Handler) ListAttendances(c echo.Context) error {
	id, err := parseUUID(c, "id")
	if err != nil {
		return handleAppError(c, err)
	}
	list, err := h.listAttendances.Handle(c.Request().Context(), application.ListAttendancesQuery{
		RequesterSchoolID: getSchoolID(c), RequesterRole: authmw.GetRole(c), ScheduleID: id,
	})
	if err != nil {
		return handleAppError(c, err)
	}
	dtos := make([]AttendanceResponse, 0, len(list))
	for _, att := range list {
		dtos = append(dtos, toAttendanceResponse(att))
	}
	return response.OK(c, "attendances retrieved", dtos)
}

func (h *Handler) UpdateAttendance(c echo.Context) error {
	scheduleID, err := parseUUID(c, "id")
	if err != nil {
		return handleAppError(c, err)
	}
	studentID, err := parseUUID(c, "student_id")
	if err != nil {
		return handleAppError(c, err)
	}
	var req UpdateAttendanceRequest
	if err := validator.BindAndValidate(c, &req); err != nil {
		return err
	}
	att, err := h.updateAttendance.Handle(c.Request().Context(), application.UpdateAttendanceCommand{
		RequesterSchoolID: getSchoolID(c), RequesterRole: authmw.GetRole(c),
		ScheduleID: scheduleID, StudentID: studentID,
		Status: req.Status, Note: req.Note, PhotoPath: req.PhotoPath,
	})
	if err != nil {
		return handleAppError(c, err)
	}
	return response.OK(c, "attendance updated", toAttendanceResponse(att))
}

// ── Conversion helpers ────────────────────────────────────────────────────────

func toScheduleResponse(cs *domain.ClassSchedule) ClassScheduleResponse {
	uuidPtrToStr := func(id *uuid.UUID) *string {
		if id == nil {
			return nil
		}
		s := id.String()
		return &s
	}
	return ClassScheduleResponse{
		ID:                    cs.ID.String(),
		SchoolID:              cs.SchoolID.String(),
		ClassroomID:           cs.ClassroomID.String(),
		ClassroomName:         cs.ClassroomName,
		SubjectID:             cs.SubjectID.String(),
		SubjectName:           cs.SubjectName,
		TeacherID:             cs.TeacherID.String(),
		TeacherName:           cs.TeacherName,
		StartPeriodID:         uuidPtrToStr(cs.StartPeriodID),
		StartPeriodNumber:     cs.StartPeriodNumber,
		StartPeriodLabel:      cs.StartPeriodLabel,
		EndPeriodID:           uuidPtrToStr(cs.EndPeriodID),
		EndPeriodNumber:       cs.EndPeriodNumber,
		Date:                  cs.Date.Format("2006-01-02"),
		StartTime:             cs.StartTime,
		EndTime:               cs.EndTime,
		TeacherAttendanceTime: cs.TeacherAttendanceTime,
		Status:                cs.Status,
		CreatedAt:             cs.CreatedAt,
		UpdatedAt:             cs.UpdatedAt,
	}
}

func toAttendanceResponse(att *domain.ClassScheduleStudent) AttendanceResponse {
	return AttendanceResponse{
		ID:                    att.ID.String(),
		ClassScheduleID:       att.ClassScheduleID.String(),
		StudentID:             att.StudentID.String(),
		StudentName:           att.StudentName,
		Status:                att.Status,
		Type:                  att.Type,
		Note:                  att.Note,
		PhotoPath:             att.PhotoPath,
		StudentAttendanceTime: att.StudentAttendanceTime,
		CreatedAt:             att.CreatedAt,
		UpdatedAt:             att.UpdatedAt,
	}
}

// ── Shared utilities ──────────────────────────────────────────────────────────

func parseOptionalUUID(s string) *uuid.UUID {
	if s == "" {
		return nil
	}
	id, err := uuid.Parse(s)
	if err != nil {
		return nil
	}
	return &id
}

func getSchoolID(c echo.Context) *uuid.UUID {
	if id := authmw.GetSchoolID(c); id != nil {
		return id
	}
	if authmw.GetRole(c) == "superadmin" {
		if raw := c.QueryParam("school_id"); raw != "" {
			if id, err := uuid.Parse(raw); err == nil {
				return &id
			}
		}
	}
	return nil
}

func parseUUID(c echo.Context, param string) (uuid.UUID, error) {
	raw := c.Param(param)
	id, err := uuid.Parse(raw)
	if err != nil {
		return uuid.UUID{}, apperror.New(apperror.ErrBadRequest, "invalid UUID: "+param)
	}
	return id, nil
}

func handleAppError(c echo.Context, err error) error {
	var appErr *apperror.AppError
	if errors.As(err, &appErr) {
		switch appErr.Err {
		case apperror.ErrNotFound:
			return response.NotFound(c, appErr.Message)
		case apperror.ErrForbidden:
			return response.Forbidden(c, appErr.Message)
		case apperror.ErrBadRequest, apperror.ErrWrongPassword:
			return response.BadRequest(c, appErr.Message)
		case apperror.ErrUnauthorized, apperror.ErrInvalidToken, apperror.ErrTokenRevoked:
			return response.Unauthorized(c, appErr.Message)
		case apperror.ErrConflict:
			return response.Conflict(c, appErr.Message)
		}
	}
	return c.JSON(http.StatusInternalServerError, response.Response{Success: false, Message: "internal server error"})
}
