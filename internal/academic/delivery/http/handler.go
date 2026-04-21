package http

import (
	"errors"
	"net/http"

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
	createLevel    *application.CreateLevelHandler
	listLevels     *application.ListLevelsHandler
	updateLevel    *application.UpdateLevelHandler
	deleteLevel    *application.DeleteLevelHandler
	createClass    *application.CreateClassHandler
	listClasses    *application.ListClassesHandler
	updateClass    *application.UpdateClassHandler
	deleteClass    *application.DeleteClassHandler
	createSubClass *application.CreateSubClassHandler
	listSubClasses *application.ListSubClassesHandler
	updateSubClass *application.UpdateSubClassHandler
	deleteSubClass *application.DeleteSubClassHandler
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
) *Handler {
	h := &Handler{
		createLevel:    createLevel,
		listLevels:     listLevels,
		updateLevel:    updateLevel,
		deleteLevel:    deleteLevel,
		createClass:    createClass,
		listClasses:    listClasses,
		updateClass:    updateClass,
		deleteClass:    deleteClass,
		createSubClass: createSubClass,
		listSubClasses: listSubClasses,
		updateSubClass: updateSubClass,
		deleteSubClass: deleteSubClass,
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
		RequesterSchoolID: authmw.GetSchoolID(c),
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
		RequesterSchoolID: authmw.GetSchoolID(c),
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
		RequesterSchoolID: authmw.GetSchoolID(c),
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
		RequesterSchoolID: authmw.GetSchoolID(c),
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
		RequesterSchoolID: authmw.GetSchoolID(c),
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
		RequesterSchoolID: authmw.GetSchoolID(c),
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
		RequesterSchoolID: authmw.GetSchoolID(c),
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
		RequesterSchoolID: authmw.GetSchoolID(c),
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
		RequesterSchoolID: authmw.GetSchoolID(c),
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
		RequesterSchoolID: authmw.GetSchoolID(c),
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
		RequesterSchoolID: authmw.GetSchoolID(c),
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
		RequesterSchoolID: authmw.GetSchoolID(c),
		RequesterRole:     authmw.GetRole(c),
		SubClassID:        id,
	}); err != nil {
		return handleAppError(c, err)
	}
	return response.OK(c, "sub-class deleted", nil)
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
