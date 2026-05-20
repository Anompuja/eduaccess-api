package http

import (
	"errors"
	"net/http"
	"time"

	"github.com/eduaccess/eduaccess-api/internal/academic/application"
	"github.com/eduaccess/eduaccess-api/internal/academic/domain"
	"github.com/eduaccess/eduaccess-api/internal/shared/apperror"
	authmw "github.com/eduaccess/eduaccess-api/internal/shared/middleware"
	"github.com/eduaccess/eduaccess-api/internal/shared/response"
	"github.com/eduaccess/eduaccess-api/internal/shared/validator"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

// Handler wires academic use-cases to HTTP endpoints.
type Handler struct {
	createLevel          *application.CreateLevelHandler
	listLevels           *application.ListLevelsHandler
	updateLevel          *application.UpdateLevelHandler
	deleteLevel          *application.DeleteLevelHandler
	createClass          *application.CreateClassHandler
	listClasses          *application.ListClassesHandler
	updateClass          *application.UpdateClassHandler
	deleteClass          *application.DeleteClassHandler
	createSubClass       *application.CreateSubClassHandler
	listSubClasses       *application.ListSubClassesHandler
	updateSubClass       *application.UpdateSubClassHandler
	deleteSubClass       *application.DeleteSubClassHandler
	createAcademicYear   *application.CreateAcademicYearHandler
	listAcademicYears    *application.ListAcademicYearsHandler
	updateAcademicYear   *application.UpdateAcademicYearHandler
	deleteAcademicYear   *application.DeleteAcademicYearHandler
	activateAcademicYear *application.ActivateAcademicYearHandler
	createSubject        *application.CreateSubjectHandler
	listSubjects         *application.ListSubjectsHandler
	updateSubject        *application.UpdateSubjectHandler
	deleteSubject        *application.DeleteSubjectHandler
	createClassroom      *application.CreateClassroomHandler
	listClassrooms       *application.ListClassroomsHandler
	updateClassroom      *application.UpdateClassroomHandler
	deleteClassroom      *application.DeleteClassroomHandler
	createSchedule       *application.CreateScheduleHandler
	listSchedules        *application.ListSchedulesHandler
	updateSchedule       *application.UpdateScheduleHandler
	deleteSchedule       *application.DeleteScheduleHandler
}

// NewHandler registers academic routes and returns the handler.
func NewHandler(
	v1 *echo.Group,
	createLevel *application.CreateLevelHandler,
	listLevels *application.ListLevelsHandler,
	updateLevel *application.UpdateLevelHandler,
	deleteLevel *application.DeleteLevelHandler,
	createClass *application.CreateClassHandler,
	listClasses *application.ListClassesHandler,
	updateClass *application.UpdateClassHandler,
	deleteClass *application.DeleteClassHandler,
	createSubClass *application.CreateSubClassHandler,
	listSubClasses *application.ListSubClassesHandler,
	updateSubClass *application.UpdateSubClassHandler,
	deleteSubClass *application.DeleteSubClassHandler,
	createAcademicYear *application.CreateAcademicYearHandler,
	listAcademicYears *application.ListAcademicYearsHandler,
	updateAcademicYear *application.UpdateAcademicYearHandler,
	deleteAcademicYear *application.DeleteAcademicYearHandler,
	activateAcademicYear *application.ActivateAcademicYearHandler,
	createSubject *application.CreateSubjectHandler,
	listSubjects *application.ListSubjectsHandler,
	updateSubject *application.UpdateSubjectHandler,
	deleteSubject *application.DeleteSubjectHandler,
	createClassroom *application.CreateClassroomHandler,
	listClassrooms *application.ListClassroomsHandler,
	updateClassroom *application.UpdateClassroomHandler,
	deleteClassroom *application.DeleteClassroomHandler,
	createSchedule *application.CreateScheduleHandler,
	listSchedules *application.ListSchedulesHandler,
	updateSchedule *application.UpdateScheduleHandler,
	deleteSchedule *application.DeleteScheduleHandler,
) *Handler {
	h := &Handler{
		createLevel:          createLevel,
		listLevels:           listLevels,
		updateLevel:          updateLevel,
		deleteLevel:          deleteLevel,
		createClass:          createClass,
		listClasses:          listClasses,
		updateClass:          updateClass,
		deleteClass:          deleteClass,
		createSubClass:       createSubClass,
		listSubClasses:       listSubClasses,
		updateSubClass:       updateSubClass,
		deleteSubClass:       deleteSubClass,
		createAcademicYear:   createAcademicYear,
		listAcademicYears:    listAcademicYears,
		updateAcademicYear:   updateAcademicYear,
		deleteAcademicYear:   deleteAcademicYear,
		activateAcademicYear: activateAcademicYear,
		createSubject:        createSubject,
		listSubjects:         listSubjects,
		updateSubject:        updateSubject,
		deleteSubject:        deleteSubject,
		createClassroom:      createClassroom,
		listClassrooms:       listClassrooms,
		updateClassroom:      updateClassroom,
		deleteClassroom:      deleteClassroom,
		createSchedule:       createSchedule,
		listSchedules:        listSchedules,
		updateSchedule:       updateSchedule,
		deleteSchedule:       deleteSchedule,
	}

	auth := authmw.RequireAuth

	levels := v1.Group("/academic/levels", auth)
	levels.POST("", h.CreateLevel)
	levels.GET("", h.ListLevels)
	levels.PUT("/:id", h.UpdateLevel)
	levels.DELETE("/:id", h.DeleteLevel)

	classes := v1.Group("/academic/classes", auth)
	classes.POST("", h.CreateClass)
	classes.GET("", h.ListClasses)
	classes.PUT("/:id", h.UpdateClass)
	classes.DELETE("/:id", h.DeleteClass)

	subClasses := v1.Group("/academic/sub-classes", auth)
	subClasses.POST("", h.CreateSubClass)
	subClasses.GET("", h.ListSubClasses)
	subClasses.PUT("/:id", h.UpdateSubClass)
	subClasses.DELETE("/:id", h.DeleteSubClass)

	academicYears := v1.Group("/academic/academic-years", auth)
	academicYears.POST("", h.CreateAcademicYear)
	academicYears.GET("", h.ListAcademicYears)
	academicYears.PUT("/:id", h.UpdateAcademicYear)
	academicYears.DELETE("/:id", h.DeleteAcademicYear)
	academicYears.PATCH("/:id/activate", h.ActivateAcademicYear)

	subjects := v1.Group("/academic/subjects", auth)
	subjects.POST("", h.CreateSubject)
	subjects.GET("", h.ListSubjects)
	subjects.PUT("/:id", h.UpdateSubject)
	subjects.DELETE("/:id", h.DeleteSubject)

	classrooms := v1.Group("/academic/classrooms", auth)
	classrooms.POST("", h.CreateClassroom)
	classrooms.GET("", h.ListClassrooms)
	classrooms.PUT("/:id", h.UpdateClassroom)
	classrooms.DELETE("/:id", h.DeleteClassroom)

	schedules := v1.Group("/academic/schedules", auth)
	schedules.POST("", h.CreateSchedule)
	schedules.GET("", h.ListSchedules)
	schedules.PUT("/:id", h.UpdateSchedule)
	schedules.DELETE("/:id", h.DeleteSchedule)

	return h
}

// CreateLevel godoc
//
//	@Summary      Create education level
//	@Tags         academic
//	@Accept       json
//	@Produce      json
//	@Security     BearerAuth
//	@Param        body  body      AcademicNameRequest  true  "Level name"
//	@Success      201   {object}  response.Response{data=EducationLevelResponse}
//	@Router       /academic/levels [post]
func (h *Handler) CreateLevel(c echo.Context) error {
	var req AcademicNameRequest
	if err := validator.BindAndValidate(c, &req); err != nil {
		return err
	}
	level, err := h.createLevel.Handle(c.Request().Context(), application.CreateLevelCommand{
		RequesterSchoolID: getSchoolID(c),
		RequesterRole:     authmw.GetRole(c),
		Name:              req.Name,
	})
	if err != nil {
		return handleAppError(c, err)
	}
	return c.JSON(http.StatusCreated, response.Response{Success: true, Message: "level created", Data: toLevelResponse(level)})
}

// ListLevels godoc
//
//	@Summary      List education levels
//	@Tags         academic
//	@Produce      json
//	@Security     BearerAuth
//	@Success      200  {object}  response.Response{data=[]EducationLevelResponse}
//	@Router       /academic/levels [get]
func (h *Handler) ListLevels(c echo.Context) error {
	levels, err := h.listLevels.Handle(c.Request().Context(), application.ListLevelsQuery{
		RequesterSchoolID: getSchoolID(c),
		RequesterRole:     authmw.GetRole(c),
	})
	if err != nil {
		return handleAppError(c, err)
	}
	dtos := make([]EducationLevelResponse, 0, len(levels))
	for _, l := range levels {
		dtos = append(dtos, toLevelResponse(l))
	}
	return response.OK(c, "levels retrieved", dtos)
}

// UpdateLevel godoc
//
//	@Summary      Update education level
//	@Tags         academic
//	@Accept       json
//	@Produce      json
//	@Security     BearerAuth
//	@Param        id    path      string              true  "Level UUID"
//	@Param        body  body      AcademicNameRequest  true  "Level name"
//	@Success      200   {object}  response.Response{data=EducationLevelResponse}
//	@Router       /academic/levels/{id} [put]
func (h *Handler) UpdateLevel(c echo.Context) error {
	id, err := parseUUID(c, "id")
	if err != nil {
		return err
	}
	var req AcademicNameRequest
	if err := validator.BindAndValidate(c, &req); err != nil {
		return err
	}
	level, err := h.updateLevel.Handle(c.Request().Context(), application.UpdateLevelCommand{
		RequesterSchoolID: getSchoolID(c),
		RequesterRole:     authmw.GetRole(c),
		LevelID:           id,
		Name:              req.Name,
	})
	if err != nil {
		return handleAppError(c, err)
	}
	return response.OK(c, "level updated", toLevelResponse(level))
}

// DeleteLevel godoc
//
//	@Summary      Delete education level
//	@Tags         academic
//	@Produce      json
//	@Security     BearerAuth
//	@Param        id   path      string  true  "Level UUID"
//	@Success      200  {object}  response.Response
//	@Router       /academic/levels/{id} [delete]
func (h *Handler) DeleteLevel(c echo.Context) error {
	id, err := parseUUID(c, "id")
	if err != nil {
		return err
	}
	if err := h.deleteLevel.Handle(c.Request().Context(), application.DeleteLevelCommand{
		RequesterSchoolID: getSchoolID(c),
		RequesterRole:     authmw.GetRole(c),
		LevelID:           id,
	}); err != nil {
		return handleAppError(c, err)
	}
	return response.OK(c, "level deleted", nil)
}

// CreateClass godoc
//
//	@Summary      Create class
//	@Tags         academic
//	@Accept       json
//	@Produce      json
//	@Security     BearerAuth
//	@Param        body  body      CreateClassRequest  true  "Class data"
//	@Success      201   {object}  response.Response{data=ClassResponse}
//	@Router       /academic/classes [post]
func (h *Handler) CreateClass(c echo.Context) error {
	var req CreateClassRequest
	if err := validator.BindAndValidate(c, &req); err != nil {
		return err
	}
	levelID, err := uuid.Parse(req.LevelID)
	if err != nil {
		return response.BadRequest(c, "invalid level_id")
	}
	class, err := h.createClass.Handle(c.Request().Context(), application.CreateClassCommand{
		RequesterSchoolID: getSchoolID(c),
		RequesterRole:     authmw.GetRole(c),
		LevelID:           levelID,
		Name:              req.Name,
	})
	if err != nil {
		return handleAppError(c, err)
	}
	return c.JSON(http.StatusCreated, response.Response{Success: true, Message: "class created", Data: toClassResponse(class)})
}

// ListClasses godoc
//
//	@Summary      List classes
//	@Tags         academic
//	@Produce      json
//	@Security     BearerAuth
//	@Param        level_id  query  string  false  "Filter by education level UUID"
//	@Success      200  {object}  response.Response{data=[]ClassResponse}
//	@Router       /academic/classes [get]
func (h *Handler) ListClasses(c echo.Context) error {
	q := application.ListClassesQuery{
		RequesterSchoolID: getSchoolID(c),
		RequesterRole:     authmw.GetRole(c),
	}
	if raw := c.QueryParam("level_id"); raw != "" {
		if id, err := uuid.Parse(raw); err == nil {
			q.LevelID = &id
		}
	}
	classes, err := h.listClasses.Handle(c.Request().Context(), q)
	if err != nil {
		return handleAppError(c, err)
	}
	dtos := make([]ClassResponse, 0, len(classes))
	for _, cl := range classes {
		dtos = append(dtos, toClassResponse(cl))
	}
	return response.OK(c, "classes retrieved", dtos)
}

// UpdateClass godoc
//
//	@Summary      Update class
//	@Tags         academic
//	@Accept       json
//	@Produce      json
//	@Security     BearerAuth
//	@Param        id    path      string              true  "Class UUID"
//	@Param        body  body      AcademicNameRequest  true  "Class name"
//	@Success      200   {object}  response.Response{data=ClassResponse}
//	@Router       /academic/classes/{id} [put]
func (h *Handler) UpdateClass(c echo.Context) error {
	id, err := parseUUID(c, "id")
	if err != nil {
		return err
	}
	var req AcademicNameRequest
	if err := validator.BindAndValidate(c, &req); err != nil {
		return err
	}
	class, err := h.updateClass.Handle(c.Request().Context(), application.UpdateClassCommand{
		RequesterSchoolID: getSchoolID(c),
		RequesterRole:     authmw.GetRole(c),
		ClassID:           id,
		Name:              req.Name,
	})
	if err != nil {
		return handleAppError(c, err)
	}
	return response.OK(c, "class updated", toClassResponse(class))
}

// DeleteClass godoc
//
//	@Summary      Delete class
//	@Tags         academic
//	@Produce      json
//	@Security     BearerAuth
//	@Param        id   path      string  true  "Class UUID"
//	@Success      200  {object}  response.Response
//	@Router       /academic/classes/{id} [delete]
func (h *Handler) DeleteClass(c echo.Context) error {
	id, err := parseUUID(c, "id")
	if err != nil {
		return err
	}
	if err := h.deleteClass.Handle(c.Request().Context(), application.DeleteClassCommand{
		RequesterSchoolID: getSchoolID(c),
		RequesterRole:     authmw.GetRole(c),
		ClassID:           id,
	}); err != nil {
		return handleAppError(c, err)
	}
	return response.OK(c, "class deleted", nil)
}

// CreateSubClass godoc
//
//	@Summary      Create sub-class
//	@Tags         academic
//	@Accept       json
//	@Produce      json
//	@Security     BearerAuth
//	@Param        body  body      CreateSubClassRequest  true  "Sub-class data"
//	@Success      201   {object}  response.Response{data=SubClassResponse}
//	@Router       /academic/sub-classes [post]
func (h *Handler) CreateSubClass(c echo.Context) error {
	var req CreateSubClassRequest
	if err := validator.BindAndValidate(c, &req); err != nil {
		return err
	}
	classID, err := uuid.Parse(req.ClassID)
	if err != nil {
		return response.BadRequest(c, "invalid class_id")
	}
	sub, err := h.createSubClass.Handle(c.Request().Context(), application.CreateSubClassCommand{
		RequesterSchoolID: getSchoolID(c),
		RequesterRole:     authmw.GetRole(c),
		ClassID:           classID,
		Name:              req.Name,
	})
	if err != nil {
		return handleAppError(c, err)
	}
	return c.JSON(http.StatusCreated, response.Response{Success: true, Message: "sub-class created", Data: toSubClassResponse(sub)})
}

// ListSubClasses godoc
//
//	@Summary      List sub-classes
//	@Tags         academic
//	@Produce      json
//	@Security     BearerAuth
//	@Param        class_id  query  string  false  "Filter by class UUID"
//	@Success      200  {object}  response.Response{data=[]SubClassResponse}
//	@Router       /academic/sub-classes [get]
func (h *Handler) ListSubClasses(c echo.Context) error {
	q := application.ListSubClassesQuery{
		RequesterSchoolID: getSchoolID(c),
		RequesterRole:     authmw.GetRole(c),
	}
	if raw := c.QueryParam("class_id"); raw != "" {
		if id, err := uuid.Parse(raw); err == nil {
			q.ClassID = &id
		}
	}
	subs, err := h.listSubClasses.Handle(c.Request().Context(), q)
	if err != nil {
		return handleAppError(c, err)
	}
	dtos := make([]SubClassResponse, 0, len(subs))
	for _, s := range subs {
		dtos = append(dtos, toSubClassResponse(s))
	}
	return response.OK(c, "sub-classes retrieved", dtos)
}

// UpdateSubClass godoc
//
//	@Summary      Update sub-class
//	@Tags         academic
//	@Accept       json
//	@Produce      json
//	@Security     BearerAuth
//	@Param        id    path      string              true  "Sub-class UUID"
//	@Param        body  body      AcademicNameRequest  true  "Sub-class name"
//	@Success      200   {object}  response.Response{data=SubClassResponse}
//	@Router       /academic/sub-classes/{id} [put]
func (h *Handler) UpdateSubClass(c echo.Context) error {
	id, err := parseUUID(c, "id")
	if err != nil {
		return err
	}
	var req AcademicNameRequest
	if err := validator.BindAndValidate(c, &req); err != nil {
		return err
	}
	sub, err := h.updateSubClass.Handle(c.Request().Context(), application.UpdateSubClassCommand{
		RequesterSchoolID: getSchoolID(c),
		RequesterRole:     authmw.GetRole(c),
		SubClassID:        id,
		Name:              req.Name,
	})
	if err != nil {
		return handleAppError(c, err)
	}
	return response.OK(c, "sub-class updated", toSubClassResponse(sub))
}

// DeleteSubClass godoc
//
//	@Summary      Delete sub-class
//	@Tags         academic
//	@Produce      json
//	@Security     BearerAuth
//	@Param        id   path      string  true  "Sub-class UUID"
//	@Success      200  {object}  response.Response
//	@Router       /academic/sub-classes/{id} [delete]
func (h *Handler) DeleteSubClass(c echo.Context) error {
	id, err := parseUUID(c, "id")
	if err != nil {
		return err
	}
	if err := h.deleteSubClass.Handle(c.Request().Context(), application.DeleteSubClassCommand{
		RequesterSchoolID: getSchoolID(c),
		RequesterRole:     authmw.GetRole(c),
		SubClassID:        id,
	}); err != nil {
		return handleAppError(c, err)
	}
	return response.OK(c, "sub-class deleted", nil)
}

// ── Academic Year endpoints ───────────────────────────────────────────────────

func (h *Handler) CreateAcademicYear(c echo.Context) error {
	var req CreateAcademicYearRequest
	if err := validator.BindAndValidate(c, &req); err != nil {
		return err
	}
	start, err := time.Parse("2006-01-02", req.StartDate)
	if err != nil {
		return response.BadRequest(c, "invalid start_date format (YYYY-MM-DD)")
	}
	end, err := time.Parse("2006-01-02", req.EndDate)
	if err != nil {
		return response.BadRequest(c, "invalid end_date format (YYYY-MM-DD)")
	}
	ay, err := h.createAcademicYear.Handle(c.Request().Context(), application.CreateAcademicYearCommand{
		RequesterSchoolID: getSchoolID(c), RequesterRole: authmw.GetRole(c),
		Name: req.Name, StartDate: start, EndDate: end, Description: req.Description,
	})
	if err != nil {
		return handleAppError(c, err)
	}
	return c.JSON(http.StatusCreated, response.Response{Success: true, Message: "academic year created", Data: toAcademicYearResponse(ay)})
}

func (h *Handler) ListAcademicYears(c echo.Context) error {
	list, err := h.listAcademicYears.Handle(c.Request().Context(), application.ListAcademicYearsQuery{
		RequesterSchoolID: getSchoolID(c), RequesterRole: authmw.GetRole(c),
	})
	if err != nil {
		return handleAppError(c, err)
	}
	dtos := make([]AcademicYearResponse, 0, len(list))
	for _, ay := range list {
		dtos = append(dtos, toAcademicYearResponse(ay))
	}
	return response.OK(c, "academic years retrieved", dtos)
}

func (h *Handler) UpdateAcademicYear(c echo.Context) error {
	id, err := parseUUID(c, "id")
	if err != nil {
		return err
	}
	var req UpdateAcademicYearRequest
	if err := validator.BindAndValidate(c, &req); err != nil {
		return err
	}
	start, err := time.Parse("2006-01-02", req.StartDate)
	if err != nil {
		return response.BadRequest(c, "invalid start_date format (YYYY-MM-DD)")
	}
	end, err := time.Parse("2006-01-02", req.EndDate)
	if err != nil {
		return response.BadRequest(c, "invalid end_date format (YYYY-MM-DD)")
	}
	ay, err := h.updateAcademicYear.Handle(c.Request().Context(), application.UpdateAcademicYearCommand{
		RequesterSchoolID: getSchoolID(c), RequesterRole: authmw.GetRole(c),
		AcademicYearID: id, Name: req.Name, StartDate: start, EndDate: end, Description: req.Description,
	})
	if err != nil {
		return handleAppError(c, err)
	}
	return response.OK(c, "academic year updated", toAcademicYearResponse(ay))
}

func (h *Handler) DeleteAcademicYear(c echo.Context) error {
	id, err := parseUUID(c, "id")
	if err != nil {
		return err
	}
	if err := h.deleteAcademicYear.Handle(c.Request().Context(), application.DeleteAcademicYearCommand{
		RequesterSchoolID: getSchoolID(c), RequesterRole: authmw.GetRole(c), AcademicYearID: id,
	}); err != nil {
		return handleAppError(c, err)
	}
	return response.OK(c, "academic year deleted", nil)
}

func (h *Handler) ActivateAcademicYear(c echo.Context) error {
	id, err := parseUUID(c, "id")
	if err != nil {
		return err
	}
	if err := h.activateAcademicYear.Handle(c.Request().Context(), application.ActivateAcademicYearCommand{
		RequesterSchoolID: getSchoolID(c), RequesterRole: authmw.GetRole(c), AcademicYearID: id,
	}); err != nil {
		return handleAppError(c, err)
	}
	return response.OK(c, "academic year activated", nil)
}

// ── Subject endpoints ─────────────────────────────────────────────────────────

func (h *Handler) CreateSubject(c echo.Context) error {
	var req CreateSubjectRequest
	if err := validator.BindAndValidate(c, &req); err != nil {
		return err
	}
	s, err := h.createSubject.Handle(c.Request().Context(), application.CreateSubjectCommand{
		RequesterSchoolID: getSchoolID(c), RequesterRole: authmw.GetRole(c), Name: req.Name, Category: req.Category,
	})
	if err != nil {
		return handleAppError(c, err)
	}
	return c.JSON(http.StatusCreated, response.Response{Success: true, Message: "subject created", Data: toSubjectResponse(s)})
}

func (h *Handler) ListSubjects(c echo.Context) error {
	list, err := h.listSubjects.Handle(c.Request().Context(), application.ListSubjectsQuery{
		RequesterSchoolID: getSchoolID(c), RequesterRole: authmw.GetRole(c),
	})
	if err != nil {
		return handleAppError(c, err)
	}
	dtos := make([]SubjectResponse, 0, len(list))
	for _, s := range list {
		dtos = append(dtos, toSubjectResponse(s))
	}
	return response.OK(c, "subjects retrieved", dtos)
}

func (h *Handler) UpdateSubject(c echo.Context) error {
	id, err := parseUUID(c, "id")
	if err != nil {
		return err
	}
	var req UpdateSubjectRequest
	if err := validator.BindAndValidate(c, &req); err != nil {
		return err
	}
	s, err := h.updateSubject.Handle(c.Request().Context(), application.UpdateSubjectCommand{
		RequesterSchoolID: getSchoolID(c), RequesterRole: authmw.GetRole(c), SubjectID: id, Name: req.Name, Category: req.Category,
	})
	if err != nil {
		return handleAppError(c, err)
	}
	return response.OK(c, "subject updated", toSubjectResponse(s))
}

func (h *Handler) DeleteSubject(c echo.Context) error {
	id, err := parseUUID(c, "id")
	if err != nil {
		return err
	}
	if err := h.deleteSubject.Handle(c.Request().Context(), application.DeleteSubjectCommand{
		RequesterSchoolID: getSchoolID(c), RequesterRole: authmw.GetRole(c), SubjectID: id,
	}); err != nil {
		return handleAppError(c, err)
	}
	return response.OK(c, "subject deleted", nil)
}

// ── Classroom endpoints ───────────────────────────────────────────────────────

func (h *Handler) CreateClassroom(c echo.Context) error {
	var req CreateClassroomRequest
	if err := validator.BindAndValidate(c, &req); err != nil {
		return err
	}
	cr, err := h.createClassroom.Handle(c.Request().Context(), application.CreateClassroomCommand{
		RequesterSchoolID: getSchoolID(c), RequesterRole: authmw.GetRole(c),
		Name: req.Name, Capacity: req.Capacity, Floor: req.Floor, Building: req.Building,
		RoomType: req.RoomType, Status: req.Status, Facilities: req.Facilities,
	})
	if err != nil {
		return handleAppError(c, err)
	}
	return c.JSON(http.StatusCreated, response.Response{Success: true, Message: "classroom created", Data: toClassroomResponse(cr)})
}

func (h *Handler) ListClassrooms(c echo.Context) error {
	list, err := h.listClassrooms.Handle(c.Request().Context(), application.ListClassroomsQuery{
		RequesterSchoolID: getSchoolID(c), RequesterRole: authmw.GetRole(c),
	})
	if err != nil {
		return handleAppError(c, err)
	}
	dtos := make([]ClassroomResponse, 0, len(list))
	for _, cr := range list {
		dtos = append(dtos, toClassroomResponse(cr))
	}
	return response.OK(c, "classrooms retrieved", dtos)
}

func (h *Handler) UpdateClassroom(c echo.Context) error {
	id, err := parseUUID(c, "id")
	if err != nil {
		return err
	}
	var req UpdateClassroomRequest
	if err := validator.BindAndValidate(c, &req); err != nil {
		return err
	}
	cr, err := h.updateClassroom.Handle(c.Request().Context(), application.UpdateClassroomCommand{
		RequesterSchoolID: getSchoolID(c), RequesterRole: authmw.GetRole(c), ClassroomID: id,
		Name: req.Name, Capacity: req.Capacity, Floor: req.Floor, Building: req.Building,
		RoomType: req.RoomType, Status: req.Status, Facilities: req.Facilities,
	})
	if err != nil {
		return handleAppError(c, err)
	}
	return response.OK(c, "classroom updated", toClassroomResponse(cr))
}

func (h *Handler) DeleteClassroom(c echo.Context) error {
	id, err := parseUUID(c, "id")
	if err != nil {
		return err
	}
	if err := h.deleteClassroom.Handle(c.Request().Context(), application.DeleteClassroomCommand{
		RequesterSchoolID: getSchoolID(c), RequesterRole: authmw.GetRole(c), ClassroomID: id,
	}); err != nil {
		return handleAppError(c, err)
	}
	return response.OK(c, "classroom deleted", nil)
}

// ── Schedule endpoints ────────────────────────────────────────────────────────

func (h *Handler) CreateSchedule(c echo.Context) error {
	var req CreateScheduleRequest
	if err := validator.BindAndValidate(c, &req); err != nil {
		return err
	}
	s, err := h.createSchedule.Handle(c.Request().Context(), application.CreateScheduleCommand{
		RequesterSchoolID: getSchoolID(c), RequesterRole: authmw.GetRole(c),
		ShiftType: req.ShiftType, StartTime: req.StartTime, EndTime: req.EndTime,
	})
	if err != nil {
		return handleAppError(c, err)
	}
	return c.JSON(http.StatusCreated, response.Response{Success: true, Message: "schedule created", Data: toScheduleResponse(s)})
}

func (h *Handler) ListSchedules(c echo.Context) error {
	list, err := h.listSchedules.Handle(c.Request().Context(), application.ListSchedulesQuery{
		RequesterSchoolID: getSchoolID(c), RequesterRole: authmw.GetRole(c),
	})
	if err != nil {
		return handleAppError(c, err)
	}
	dtos := make([]ScheduleResponse, 0, len(list))
	for _, s := range list {
		dtos = append(dtos, toScheduleResponse(s))
	}
	return response.OK(c, "schedules retrieved", dtos)
}

func (h *Handler) UpdateSchedule(c echo.Context) error {
	id, err := parseUUID(c, "id")
	if err != nil {
		return err
	}
	var req UpdateScheduleRequest
	if err := validator.BindAndValidate(c, &req); err != nil {
		return err
	}
	s, err := h.updateSchedule.Handle(c.Request().Context(), application.UpdateScheduleCommand{
		RequesterSchoolID: getSchoolID(c), RequesterRole: authmw.GetRole(c), ScheduleID: id,
		ShiftType: req.ShiftType, StartTime: req.StartTime, EndTime: req.EndTime,
	})
	if err != nil {
		return handleAppError(c, err)
	}
	return response.OK(c, "schedule updated", toScheduleResponse(s))
}

func (h *Handler) DeleteSchedule(c echo.Context) error {
	id, err := parseUUID(c, "id")
	if err != nil {
		return err
	}
	if err := h.deleteSchedule.Handle(c.Request().Context(), application.DeleteScheduleCommand{
		RequesterSchoolID: getSchoolID(c), RequesterRole: authmw.GetRole(c), ScheduleID: id,
	}); err != nil {
		return handleAppError(c, err)
	}
	return response.OK(c, "schedule deleted", nil)
}

// ── Conversion helpers ────────────────────────────────────────────────────────

func toAcademicYearResponse(ay *domain.AcademicYear) AcademicYearResponse {
	return AcademicYearResponse{
		ID: ay.ID.String(), SchoolID: ay.SchoolID.String(), Name: ay.Name,
		StartDate: ay.StartDate, EndDate: ay.EndDate, IsActive: ay.IsActive,
		Description: ay.Description, CreatedAt: ay.CreatedAt, UpdatedAt: ay.UpdatedAt,
	}
}

func toSubjectResponse(s *domain.Subject) SubjectResponse {
	return SubjectResponse{
		ID: s.ID.String(), SchoolID: s.SchoolID.String(), Name: s.Name,
		Category: s.Category, CreatedAt: s.CreatedAt, UpdatedAt: s.UpdatedAt,
	}
}

func toClassroomResponse(c *domain.Classroom) ClassroomResponse {
	return ClassroomResponse{
		ID: c.ID.String(), SchoolID: c.SchoolID.String(), Name: c.Name,
		Capacity: c.Capacity, Floor: c.Floor, Building: c.Building, RoomType: c.RoomType,
		Status: c.Status, Facilities: c.Facilities, CreatedAt: c.CreatedAt, UpdatedAt: c.UpdatedAt,
	}
}

func toScheduleResponse(s *domain.Schedule) ScheduleResponse {
	return ScheduleResponse{
		ID: s.ID.String(), SchoolID: s.SchoolID.String(), ShiftType: s.ShiftType,
		StartTime: s.StartTime, EndTime: s.EndTime, CreatedAt: s.CreatedAt, UpdatedAt: s.UpdatedAt,
	}
}

func toLevelResponse(l *domain.EducationLevel) EducationLevelResponse {
	return EducationLevelResponse{
		ID:        l.ID.String(),
		SchoolID:  l.SchoolID.String(),
		Name:      l.Name,
		CreatedAt: l.CreatedAt,
		UpdatedAt: l.UpdatedAt,
	}
}

func toClassResponse(c *domain.Class) ClassResponse {
	return ClassResponse{
		ID:               c.ID.String(),
		SchoolID:         c.SchoolID.String(),
		EducationLevelID: c.EducationLevelID.String(),
		Name:             c.Name,
		CreatedAt:        c.CreatedAt,
		UpdatedAt:        c.UpdatedAt,
	}
}

func toSubClassResponse(s *domain.SubClass) SubClassResponse {
	return SubClassResponse{
		ID:        s.ID.String(),
		SchoolID:  s.SchoolID.String(),
		ClassID:   s.ClassID.String(),
		Name:      s.Name,
		CreatedAt: s.CreatedAt,
		UpdatedAt: s.UpdatedAt,
	}
}

// getSchoolID returns the school UUID for the current request.
// For superadmin (whose JWT has no school_id) it falls back to the
// ?school_id=<uuid> query parameter so they can target a specific school.
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
		_ = c.JSON(http.StatusBadRequest, response.Response{
			Success: false,
			Message: "invalid UUID: " + param,
		})
		return uuid.UUID{}, echo.ErrBadRequest
	}
	return id, nil
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
