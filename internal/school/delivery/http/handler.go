package http

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/eduaccess/eduaccess-api/internal/school/application"
	"github.com/eduaccess/eduaccess-api/internal/school/domain"
	"github.com/eduaccess/eduaccess-api/internal/shared/apperror"
	authmw "github.com/eduaccess/eduaccess-api/internal/shared/middleware"
	"github.com/eduaccess/eduaccess-api/internal/shared/response"
	"github.com/eduaccess/eduaccess-api/internal/shared/validator"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

// Handler wires school use-cases to HTTP endpoints.
type Handler struct {
	createSchool      *application.CreateSchoolHandler
	listSchools       *application.ListSchoolsHandler
	getSchool         *application.GetSchoolHandler
	updateSchool      *application.UpdateSchoolHandler
	deactivateSchool  *application.DeactivateSchoolHandler
	listRules         *application.ListRulesHandler
	upsertRules       *application.UpsertRulesHandler
	getSubscription   *application.GetSubscriptionHandler
}

// NewHandler registers school routes and returns the handler.
func NewHandler(
	v1 *echo.Group,
	createSchool *application.CreateSchoolHandler,
	listSchools *application.ListSchoolsHandler,
	getSchool *application.GetSchoolHandler,
	updateSchool *application.UpdateSchoolHandler,
	deactivateSchool *application.DeactivateSchoolHandler,
	listRules *application.ListRulesHandler,
	upsertRules *application.UpsertRulesHandler,
	getSubscription *application.GetSubscriptionHandler,
) *Handler {
	h := &Handler{
		createSchool:     createSchool,
		listSchools:      listSchools,
		getSchool:        getSchool,
		updateSchool:     updateSchool,
		deactivateSchool: deactivateSchool,
		listRules:        listRules,
		upsertRules:      upsertRules,
		getSubscription:  getSubscription,
	}

	schools := v1.Group("/schools", authmw.RequireAuth)
	schools.POST("", h.CreateSchool)
	schools.GET("", h.ListSchools)
	schools.GET("/:id", h.GetSchool)
	schools.PUT("/:id", h.UpdateSchool)
	schools.DELETE("/:id", h.DeactivateSchool)
	schools.GET("/:id/rules", h.ListRules)
	schools.PUT("/:id/rules", h.UpsertRules)
	schools.GET("/:id/subscription", h.GetSubscription)

	return h
}

// CreateSchool godoc
//
//	@Summary      Create school
//	@Description  Creates a new school tenant. Superadmin only.
//	@Tags         schools
//	@Accept       json
//	@Produce      json
//	@Security     BearerAuth
//	@Param        body  body      CreateSchoolRequest  true  "School data"
//	@Success      201   {object}  response.Response{data=SchoolResponse}
//	@Failure      400   {object}  response.Response
//	@Failure      403   {object}  response.Response
//	@Failure      409   {object}  response.Response
//	@Router       /schools [post]
func (h *Handler) CreateSchool(c echo.Context) error {
	var req CreateSchoolRequest
	if err := validator.BindAndValidate(c, &req); err != nil {
		return err
	}

	school, err := h.createSchool.Handle(c.Request().Context(), application.CreateSchoolCommand{
		RequesterRole: authmw.GetRole(c),
		Name:          req.Name,
		Address:       req.Address,
		Phone:         req.Phone,
		Email:         req.Email,
		Description:   req.Description,
		ImagePath:     req.ImagePath,
		TimeZone:      req.TimeZone,
	})
	if err != nil {
		return handleAppError(c, err)
	}

	return c.JSON(http.StatusCreated, response.Response{
		Success: true,
		Message: "school created",
		Data:    toSchoolResponse(school),
	})
}

// ListSchools godoc
//
//	@Summary      List schools
//	@Description  Returns paginated schools. Superadmin sees all; others see their own.
//	@Tags         schools
//	@Produce      json
//	@Security     BearerAuth
//	@Param        search   query  string  false  "Search by name or email"
//	@Param        status   query  string  false  "Filter by status (active|nonactive)"
//	@Param        page     query  int     false  "Page number (default 1)"
//	@Param        per_page query  int     false  "Page size (default 20)"
//	@Success      200  {object}  response.PaginatedResponse{data=[]SchoolResponse}
//	@Failure      401  {object}  response.Response
//	@Router       /schools [get]
func (h *Handler) ListSchools(c echo.Context) error {
	page, _ := strconv.Atoi(c.QueryParam("page"))
	perPage, _ := strconv.Atoi(c.QueryParam("per_page"))

	result, err := h.listSchools.Handle(c.Request().Context(), application.ListSchoolsQuery{
		RequesterSchoolID: authmw.GetSchoolID(c),
		RequesterRole:     authmw.GetRole(c),
		Search:            c.QueryParam("search"),
		Status:            c.QueryParam("status"),
		Page:              page,
		PerPage:           perPage,
	})
	if err != nil {
		return handleAppError(c, err)
	}

	dtos := make([]SchoolResponse, 0, len(result.Schools))
	for _, s := range result.Schools {
		dtos = append(dtos, toSchoolResponse(s))
	}

	return response.Paginated(c, "schools retrieved", dtos, result.Page, result.PerPage, result.Total)
}

// GetSchool godoc
//
//	@Summary      Get school by ID
//	@Description  Returns a single school. Others can only fetch their own school.
//	@Tags         schools
//	@Produce      json
//	@Security     BearerAuth
//	@Param        id   path      string  true  "School UUID"
//	@Success      200  {object}  response.Response{data=SchoolResponse}
//	@Failure      403  {object}  response.Response
//	@Failure      404  {object}  response.Response
//	@Router       /schools/{id} [get]
func (h *Handler) GetSchool(c echo.Context) error {
	id, err := parseUUID(c, "id")
	if err != nil {
		return err
	}

	school, err := h.getSchool.Handle(c.Request().Context(), application.GetSchoolQuery{
		RequesterSchoolID: authmw.GetSchoolID(c),
		RequesterRole:     authmw.GetRole(c),
		SchoolID:          id,
	})
	if err != nil {
		return handleAppError(c, err)
	}

	return response.OK(c, "school retrieved", toSchoolResponse(school))
}

// UpdateSchool godoc
//
//	@Summary      Update school
//	@Description  Updates school fields. admin_sekolah can update their own school; superadmin can update any and change status.
//	@Tags         schools
//	@Accept       json
//	@Produce      json
//	@Security     BearerAuth
//	@Param        id    path      string             true  "School UUID"
//	@Param        body  body      UpdateSchoolRequest true  "Fields to update"
//	@Success      200   {object}  response.Response{data=SchoolResponse}
//	@Failure      400   {object}  response.Response
//	@Failure      403   {object}  response.Response
//	@Failure      404   {object}  response.Response
//	@Router       /schools/{id} [put]
func (h *Handler) UpdateSchool(c echo.Context) error {
	id, err := parseUUID(c, "id")
	if err != nil {
		return err
	}

	var req UpdateSchoolRequest
	if err := validator.BindAndValidate(c, &req); err != nil {
		return err
	}

	school, err := h.updateSchool.Handle(c.Request().Context(), application.UpdateSchoolCommand{
		RequesterSchoolID: authmw.GetSchoolID(c),
		RequesterRole:     authmw.GetRole(c),
		SchoolID:          id,
		Name:              req.Name,
		Address:           req.Address,
		Phone:             req.Phone,
		Email:             req.Email,
		Description:       req.Description,
		ImagePath:         req.ImagePath,
		TimeZone:          req.TimeZone,
		Status:            req.Status,
	})
	if err != nil {
		return handleAppError(c, err)
	}

	return response.OK(c, "school updated", toSchoolResponse(school))
}

// DeactivateSchool godoc
//
//	@Summary      Deactivate school
//	@Description  Soft-deletes a school. Superadmin only.
//	@Tags         schools
//	@Produce      json
//	@Security     BearerAuth
//	@Param        id   path      string  true  "School UUID"
//	@Success      200  {object}  response.Response
//	@Failure      403  {object}  response.Response
//	@Failure      404  {object}  response.Response
//	@Router       /schools/{id} [delete]
func (h *Handler) DeactivateSchool(c echo.Context) error {
	id, err := parseUUID(c, "id")
	if err != nil {
		return err
	}

	if err := h.deactivateSchool.Handle(c.Request().Context(), application.DeactivateSchoolCommand{
		RequesterRole: authmw.GetRole(c),
		SchoolID:      id,
	}); err != nil {
		return handleAppError(c, err)
	}

	return response.OK(c, "school deactivated", nil)
}

// ListRules godoc
//
//	@Summary      List school rules
//	@Description  Returns all key-value rules for a school.
//	@Tags         schools
//	@Produce      json
//	@Security     BearerAuth
//	@Param        id   path      string  true  "School UUID"
//	@Success      200  {object}  response.Response{data=[]SchoolRuleResponse}
//	@Failure      403  {object}  response.Response
//	@Failure      404  {object}  response.Response
//	@Router       /schools/{id}/rules [get]
func (h *Handler) ListRules(c echo.Context) error {
	id, err := parseUUID(c, "id")
	if err != nil {
		return err
	}

	rules, err := h.listRules.Handle(c.Request().Context(), application.ListRulesQuery{
		RequesterSchoolID: authmw.GetSchoolID(c),
		RequesterRole:     authmw.GetRole(c),
		SchoolID:          id,
	})
	if err != nil {
		return handleAppError(c, err)
	}

	dtos := make([]SchoolRuleResponse, 0, len(rules))
	for _, r := range rules {
		dtos = append(dtos, toRuleResponse(r))
	}

	return response.OK(c, "rules retrieved", dtos)
}

// UpsertRules godoc
//
//	@Summary      Upsert school rules
//	@Description  Creates or updates multiple key-value rules for a school.
//	@Tags         schools
//	@Accept       json
//	@Produce      json
//	@Security     BearerAuth
//	@Param        id    path      string             true  "School UUID"
//	@Param        body  body      UpsertRulesRequest  true  "Rules to upsert"
//	@Success      200   {object}  response.Response{data=[]SchoolRuleResponse}
//	@Failure      400   {object}  response.Response
//	@Failure      403   {object}  response.Response
//	@Router       /schools/{id}/rules [put]
func (h *Handler) UpsertRules(c echo.Context) error {
	id, err := parseUUID(c, "id")
	if err != nil {
		return err
	}

	var req UpsertRulesRequest
	if err := validator.BindAndValidate(c, &req); err != nil {
		return err
	}

	appRules := make([]application.RuleInput, 0, len(req.Rules))
	for _, ri := range req.Rules {
		appRules = append(appRules, application.RuleInput{
			Key:   ri.Key,
			Value: ri.Value,
			Note:  ri.Note,
		})
	}

	rules, err := h.upsertRules.Handle(c.Request().Context(), application.UpsertRulesCommand{
		RequesterSchoolID: authmw.GetSchoolID(c),
		RequesterRole:     authmw.GetRole(c),
		SchoolID:          id,
		Rules:             appRules,
	})
	if err != nil {
		return handleAppError(c, err)
	}

	dtos := make([]SchoolRuleResponse, 0, len(rules))
	for _, r := range rules {
		dtos = append(dtos, toRuleResponse(r))
	}

	return response.OK(c, "rules updated", dtos)
}

// GetSubscription godoc
//
//	@Summary      Get school subscription
//	@Description  Returns the active subscription for a school.
//	@Tags         schools
//	@Produce      json
//	@Security     BearerAuth
//	@Param        id   path      string  true  "School UUID"
//	@Success      200  {object}  response.Response{data=SubscriptionResponse}
//	@Failure      403  {object}  response.Response
//	@Failure      404  {object}  response.Response
//	@Router       /schools/{id}/subscription [get]
func (h *Handler) GetSubscription(c echo.Context) error {
	id, err := parseUUID(c, "id")
	if err != nil {
		return err
	}

	sub, err := h.getSubscription.Handle(c.Request().Context(), application.GetSubscriptionQuery{
		RequesterSchoolID: authmw.GetSchoolID(c),
		RequesterRole:     authmw.GetRole(c),
		SchoolID:          id,
	})
	if err != nil {
		return handleAppError(c, err)
	}

	return response.OK(c, "subscription retrieved", toSubscriptionResponse(sub))
}

// ── helpers ───────────────────────────────────────────────────────────────────

func toSchoolResponse(s *domain.School) SchoolResponse {
	dto := SchoolResponse{
		ID:          s.ID.String(),
		Name:        s.Name,
		Address:     s.Address,
		Phone:       s.Phone,
		Email:       s.Email,
		Description: s.Description,
		ImagePath:   s.ImagePath,
		TimeZone:    s.TimeZone,
		Status:      s.Status,
		CreatedAt:   s.CreatedAt,
		UpdatedAt:   s.UpdatedAt,
	}
	if s.HeadmasterID != nil {
		str := s.HeadmasterID.String()
		dto.HeadmasterID = &str
	}
	if s.Subscription != nil {
		sub := toSubscriptionResponse(s.Subscription)
		dto.Subscription = &sub
	}
	return dto
}

func toRuleResponse(r *domain.SchoolRule) SchoolRuleResponse {
	return SchoolRuleResponse{
		ID:        r.ID.String(),
		SchoolID:  r.SchoolID.String(),
		Key:       r.Key,
		Value:     r.Value,
		Note:      r.Note,
		UpdatedAt: r.UpdatedAt,
	}
}

func toSubscriptionResponse(s *domain.Subscription) SubscriptionResponse {
	dto := SubscriptionResponse{
		ID:        s.ID.String(),
		SchoolID:  s.SchoolID.String(),
		Status:    s.Status,
		Cycle:     s.Cycle,
		Quantity:  s.Quantity,
		Price:     s.Price,
		EndsAt:    s.EndsAt,
		CreatedAt: s.CreatedAt,
		UpdatedAt: s.UpdatedAt,
	}
	if s.Plan != nil {
		features := s.Plan.Features
		if features == nil {
			features = []string{}
		}
		plan := PlanResponse{
			ID:           s.Plan.ID.String(),
			Name:         s.Plan.Name,
			Description:  s.Plan.Description,
			Features:     features,
			MonthlyPrice: s.Plan.MonthlyPrice,
			YearlyPrice:  s.Plan.YearlyPrice,
		}
		dto.Plan = &plan
	}
	return dto
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
