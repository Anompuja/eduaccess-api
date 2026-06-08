package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"time"

	"github.com/joho/godotenv"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"

	_ "github.com/eduaccess/eduaccess-api/docs"
	academicApp "github.com/eduaccess/eduaccess-api/internal/academic/application"
	academicHTTP "github.com/eduaccess/eduaccess-api/internal/academic/delivery/http"
	academicInfra "github.com/eduaccess/eduaccess-api/internal/academic/infrastructure"
	adminApp "github.com/eduaccess/eduaccess-api/internal/admin/application"
	adminHTTP "github.com/eduaccess/eduaccess-api/internal/admin/delivery/http"
	adminInfra "github.com/eduaccess/eduaccess-api/internal/admin/infrastructure"
	authApp "github.com/eduaccess/eduaccess-api/internal/auth/application"
	authHTTP "github.com/eduaccess/eduaccess-api/internal/auth/delivery/http"
	authInfra "github.com/eduaccess/eduaccess-api/internal/auth/infrastructure"
	billingApp "github.com/eduaccess/eduaccess-api/internal/billing/application"
	billingHTTP "github.com/eduaccess/eduaccess-api/internal/billing/delivery/http"
	billingInfra "github.com/eduaccess/eduaccess-api/internal/billing/infrastructure"
	classScheduleApp "github.com/eduaccess/eduaccess-api/internal/class_schedule/application"
	classScheduleHTTP "github.com/eduaccess/eduaccess-api/internal/class_schedule/delivery/http"
	classScheduleInfra "github.com/eduaccess/eduaccess-api/internal/class_schedule/infrastructure"
	dashboardApp "github.com/eduaccess/eduaccess-api/internal/dashboard/application"
	dashboardHTTP "github.com/eduaccess/eduaccess-api/internal/dashboard/delivery/http"
	dashboardInfra "github.com/eduaccess/eduaccess-api/internal/dashboard/infrastructure"
	headmasterApp "github.com/eduaccess/eduaccess-api/internal/headmaster/application"
	headmasterHTTP "github.com/eduaccess/eduaccess-api/internal/headmaster/delivery/http"
	headmasterInfra "github.com/eduaccess/eduaccess-api/internal/headmaster/infrastructure"
	parentApp "github.com/eduaccess/eduaccess-api/internal/parent/application"
	parentHTTP "github.com/eduaccess/eduaccess-api/internal/parent/delivery/http"
	parentInfra "github.com/eduaccess/eduaccess-api/internal/parent/infrastructure"
	schoolApp "github.com/eduaccess/eduaccess-api/internal/school/application"
	schoolHTTP "github.com/eduaccess/eduaccess-api/internal/school/delivery/http"
	schoolInfra "github.com/eduaccess/eduaccess-api/internal/school/infrastructure"
	appvalidator "github.com/eduaccess/eduaccess-api/internal/shared/validator"
	staffApp "github.com/eduaccess/eduaccess-api/internal/staff/application"
	staffHTTP "github.com/eduaccess/eduaccess-api/internal/staff/delivery/http"
	staffInfra "github.com/eduaccess/eduaccess-api/internal/staff/infrastructure"
	storageHTTP "github.com/eduaccess/eduaccess-api/internal/storage/delivery/http"
	studentApp "github.com/eduaccess/eduaccess-api/internal/student/application"
	studentHTTP "github.com/eduaccess/eduaccess-api/internal/student/delivery/http"
	studentInfra "github.com/eduaccess/eduaccess-api/internal/student/infrastructure"
	studentPromotionApp "github.com/eduaccess/eduaccess-api/internal/student_promotion/application"
	studentPromotionHTTP "github.com/eduaccess/eduaccess-api/internal/student_promotion/delivery/http"
	studentPromotionInfra "github.com/eduaccess/eduaccess-api/internal/student_promotion/infrastructure"
	studentTrackingApp "github.com/eduaccess/eduaccess-api/internal/student_tracking/application"
	studentTrackingHTTP "github.com/eduaccess/eduaccess-api/internal/student_tracking/delivery/http"
	studentTrackingInfra "github.com/eduaccess/eduaccess-api/internal/student_tracking/infrastructure"
	teacherApp "github.com/eduaccess/eduaccess-api/internal/teacher/application"
	teacherHTTP "github.com/eduaccess/eduaccess-api/internal/teacher/delivery/http"
	teacherInfra "github.com/eduaccess/eduaccess-api/internal/teacher/infrastructure"
	userApp "github.com/eduaccess/eduaccess-api/internal/user/application"
	userHTTP "github.com/eduaccess/eduaccess-api/internal/user/delivery/http"
	userInfra "github.com/eduaccess/eduaccess-api/internal/user/infrastructure"
	"github.com/eduaccess/eduaccess-api/pkg/database"
	supabasePkg "github.com/eduaccess/eduaccess-api/pkg/supabase"
	echoSwagger "github.com/swaggo/echo-swagger"
)

// @title           EduAccess API
// @version         1.0
// @description     Multi-tenant School Management SaaS API
// @termsOfService  http://swagger.io/terms/

// @contact.name   EduAccess Support
// @contact.email  support@eduaccess.id

// @license.name  MIT

// @host      localhost:8080
// @BasePath  /api/v1

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and a Supabase JWT.
func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, reading from environment")
	}

	db, err := database.Connect()
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}

	supabase := supabasePkg.NewClient()

	e := echo.New()
	e.HideBanner = true
	e.Validator = appvalidator.New()

	// Middleware
	e.Use(middleware.Recover())
	e.Use(middleware.Logger())
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins:  getAllowedOrigins(),
		AllowMethods:  []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete, http.MethodOptions},
		AllowHeaders:  []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAccept, echo.HeaderAuthorization, echo.HeaderXRequestedWith, "If-None-Match", "If-Modified-Since"},
		ExposeHeaders: []string{"ETag", "Cache-Control"},
	}))
	e.Use(middleware.RequestID())

	// Swagger
	e.GET("/swagger/*", echoSwagger.WrapHandler)

	// Health check
	e.GET("/health", func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]string{"status": "ok"})
	})

	// API v1 group
	v1 := e.Group("/api/v1")

	// ΓöÇΓöÇ Auth module ΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇ
	// Login, register, and session management are handled by Supabase Auth SDK
	// on the frontend. The backend validates Supabase JWTs in middleware.
	userRepo := authInfra.NewSupabaseUserRepository(db, supabase)
	registerHandler := authApp.NewRegisterHandler(userRepo)
	authHTTP.NewHandler(v1, registerHandler, supabase, userRepo)

	// ΓöÇΓöÇ User management module ΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇ
	userMgmtRepo := userInfra.NewGormUserManagementRepository(db)
	userHTTP.NewHandler(
		v1,
		userApp.NewListUsersHandler(userMgmtRepo),
		userApp.NewGetUserHandler(userMgmtRepo),
		userApp.NewUpdateUserHandler(userMgmtRepo, supabase),
		userApp.NewDeactivateUserHandler(userMgmtRepo),
		userApp.NewChangePasswordHandler(userMgmtRepo, supabase),
	)

	// ΓöÇΓöÇ School setup module ΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇ
	schoolRepo := schoolInfra.NewGormSchoolRepository(db)
	studentRepo := studentInfra.NewGormStudentRepository(db)
	parentRepo := parentInfra.NewGormParentRepository(db)
	schoolHTTP.NewHandler(
		v1,
		schoolApp.NewCreateSchoolHandler(schoolRepo),
		schoolApp.NewListSchoolsHandler(schoolRepo),
		schoolApp.NewGetSchoolHandler(schoolRepo),
		schoolApp.NewUpdateSchoolHandler(schoolRepo),
		schoolApp.NewDeactivateSchoolHandler(schoolRepo),
		schoolApp.NewListPlansHandler(schoolRepo),
		schoolApp.NewListRulesHandler(schoolRepo),
		schoolApp.NewUpsertRulesHandler(schoolRepo),
		schoolApp.NewGetSubscriptionHandler(schoolRepo),
		schoolApp.NewUpdateSubscriptionHandler(schoolRepo, studentRepo),
	)

	paymentRepo := billingInfra.NewGormPaymentRepository(db)
	midtransClient := billingInfra.NewMidtransClientFromEnv()
	billingHTTP.NewHandler(
		v1,
		billingApp.NewCreateCheckoutHandler(paymentRepo, schoolRepo, studentRepo, midtransClient),
		billingApp.NewListPaymentsHandler(paymentRepo),
		billingApp.NewGetPaymentHandler(paymentRepo, schoolRepo, studentRepo, midtransClient),
		billingApp.NewHandleMidtransNotificationHandler(paymentRepo, schoolRepo, studentRepo, midtransClient),
	)

	// ΓöÇΓöÇ Dashboard module ΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇ
	dashboardRepo := dashboardInfra.NewGormDashboardRepository(db)
	dashboardHTTP.NewHandler(
		v1,
		dashboardApp.NewGetStatsHandler(dashboardRepo),
	)

	// ΓöÇΓöÇ Headmaster module ΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇ
	headmasterRepo := headmasterInfra.NewGormHeadmasterRepository(db)
	headmasterHTTP.NewHandler(
		v1,
		headmasterApp.NewCreateHeadmasterHandler(userRepo, headmasterRepo, schoolRepo),
		headmasterApp.NewListHeadmastersHandler(headmasterRepo),
		headmasterApp.NewGetHeadmasterHandler(headmasterRepo),
		headmasterApp.NewUpdateHeadmasterHandler(headmasterRepo),
		headmasterApp.NewDeactivateHeadmasterHandler(headmasterRepo),
	)

	// ΓöÇΓöÇ Student module ΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇ
	academicRepo := academicInfra.NewGormAcademicRepository(db)
	studentHTTP.NewHandler(
		v1,
		studentApp.NewCreateStudentHandler(userRepo, studentRepo, academicRepo, schoolRepo),
		studentApp.NewListStudentsHandler(studentRepo),
		studentApp.NewGetStudentHandler(studentRepo),
		studentApp.NewUpdateStudentHandler(studentRepo),
		studentApp.NewDeactivateStudentHandler(studentRepo),
		studentApp.NewLinkParentHandler(studentRepo, parentRepo),
		studentApp.NewUnlinkParentHandler(studentRepo),
	)

	// ΓöÇΓöÇ Academic module ΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇ
	academicHTTP.NewHandler(
		v1,
		academicApp.NewCreateLevelHandler(academicRepo),
		academicApp.NewListLevelsHandler(academicRepo),
		academicApp.NewUpdateLevelHandler(academicRepo),
		academicApp.NewDeleteLevelHandler(academicRepo),
		academicApp.NewCreateClassHandler(academicRepo),
		academicApp.NewListClassesHandler(academicRepo),
		academicApp.NewUpdateClassHandler(academicRepo),
		academicApp.NewDeleteClassHandler(academicRepo),
		academicApp.NewCreateSubClassHandler(academicRepo),
		academicApp.NewListSubClassesHandler(academicRepo),
		academicApp.NewUpdateSubClassHandler(academicRepo),
		academicApp.NewDeleteSubClassHandler(academicRepo),
		academicApp.NewCreateAcademicYearHandler(academicRepo),
		academicApp.NewListAcademicYearsHandler(academicRepo),
		academicApp.NewUpdateAcademicYearHandler(academicRepo),
		academicApp.NewDeleteAcademicYearHandler(academicRepo),
		academicApp.NewActivateAcademicYearHandler(academicRepo),
		academicApp.NewCreateSubjectHandler(academicRepo),
		academicApp.NewListSubjectsHandler(academicRepo),
		academicApp.NewUpdateSubjectHandler(academicRepo),
		academicApp.NewDeleteSubjectHandler(academicRepo),
		academicApp.NewCreateClassroomHandler(academicRepo),
		academicApp.NewListClassroomsHandler(academicRepo),
		academicApp.NewUpdateClassroomHandler(academicRepo),
		academicApp.NewDeleteClassroomHandler(academicRepo),
		academicApp.NewCreateScheduleHandler(academicRepo),
		academicApp.NewListSchedulesHandler(academicRepo),
		academicApp.NewUpdateScheduleHandler(academicRepo),
		academicApp.NewDeleteScheduleHandler(academicRepo),
	)

	// ΓöÇΓöÇ Parent module ΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇ
	parentHTTP.NewHandler(
		v1,
		parentApp.NewCreateParentHandler(userRepo, parentRepo),
		parentApp.NewListParentsHandler(parentRepo),
		parentApp.NewGetParentHandler(parentRepo),
		parentApp.NewUpdateParentHandler(userMgmtRepo, parentRepo),
		parentApp.NewDeactivateParentHandler(parentRepo),
	)

	// ΓöÇΓöÇ Admin module ΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇ
	adminRepo := adminInfra.NewGormAdminRepository(db)
	adminHTTP.NewHandler(
		v1,
		adminApp.NewCreateAdminHandler(userRepo, adminRepo),
		adminApp.NewGetAdminHandler(adminRepo),
		adminApp.NewListAdminsHandler(adminRepo),
		adminApp.NewUpdateAdminHandler(userMgmtRepo, adminRepo),
		adminApp.NewDeactivateAdminHandler(adminRepo),
	)

	// ΓöÇΓöÇ Teacher module ΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇ
	teacherRepo := teacherInfra.NewTeacherRepository(db)
	teacherHTTP.NewHandler(
		v1,
		teacherApp.NewCreateTeacherHandler(teacherRepo, userRepo),
		teacherApp.NewGetTeacherHandler(teacherRepo),
		teacherApp.NewListTeachersHandler(teacherRepo),
		teacherApp.NewUpdateTeacherHandler(teacherRepo, userMgmtRepo),
		teacherApp.NewDeactivateTeacherHandler(teacherRepo),
	)

	// ΓöÇΓöÇ Staff module ΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇ
	staffRepo := staffInfra.NewStaffRepository(db)
	staffHTTP.NewHandler(
		v1,
		staffApp.NewCreateStaffHandler(staffRepo, userRepo),
		staffApp.NewGetStaffHandler(staffRepo),
		staffApp.NewListStaffHandler(staffRepo),
		staffApp.NewUpdateStaffHandler(staffRepo, userMgmtRepo),
		staffApp.NewDeactivateStaffHandler(staffRepo),
	)

	// ΓöÇΓöÇ Class Schedule module ΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇ
	classScheduleRepo := classScheduleInfra.NewGormClassScheduleRepository(db)
	classScheduleHTTP.NewHandler(
		v1,
		classScheduleApp.NewCreateClassScheduleHandler(classScheduleRepo),
		classScheduleApp.NewListClassSchedulesHandler(classScheduleRepo),
		classScheduleApp.NewGetClassScheduleHandler(classScheduleRepo),
		classScheduleApp.NewUpdateClassScheduleHandler(classScheduleRepo),
		classScheduleApp.NewDeleteClassScheduleHandler(classScheduleRepo),
		classScheduleApp.NewStartClassScheduleHandler(classScheduleRepo),
		classScheduleApp.NewCompleteClassScheduleHandler(classScheduleRepo),
		classScheduleApp.NewCancelClassScheduleHandler(classScheduleRepo),
		classScheduleApp.NewSyncStudentsHandler(classScheduleRepo),
		classScheduleApp.NewListAttendancesHandler(classScheduleRepo),
		classScheduleApp.NewUpdateAttendanceHandler(classScheduleRepo),
	)

	// ΓöÇΓöÇ Student Tracking module ΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇ
	studentTrackingRepo := studentTrackingInfra.NewGormRepository(db)
	studentTrackingHTTP.NewHandler(
		v1,
		studentTrackingApp.NewListStudiesHandler(studentTrackingRepo),
		studentTrackingApp.NewGetStudentDetailHandler(studentTrackingRepo),
	)

	// ΓöÇΓöÇ Student Promotion module (kenaikan kelas) ΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇ
	studentPromotionRepo := studentPromotionInfra.NewGormRepository(db)
	studentPromotionHTTP.NewHandler(
		v1,
		studentPromotionApp.NewListPromotionsHandler(studentPromotionRepo),
		studentPromotionApp.NewPromoteHandler(studentPromotionRepo),
	)

	// ΓöÇΓöÇ Storage module ΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇΓöÇ
	storageHTTP.NewHandler(v1, supabase)

	// Resolve port ΓÇö Heroku injects $PORT at runtime
	port := os.Getenv("PORT")
	if port == "" {
		port = os.Getenv("APP_PORT")
	}
	if port == "" {
		port = "8080"
	}

	// Graceful shutdown
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	go func() {
		addr := fmt.Sprintf(":%s", port)
		log.Printf("Server is running at http://localhost:%s", port)
		log.Printf("Swagger documentation at http://localhost:%s/swagger/index.html", port)
		if err := e.Start(addr); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server error: %v", err)
		}
	}()

	<-ctx.Done()
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := e.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("server shutdown error: %v", err)
	}
	log.Println("server stopped")
}

func getAllowedOrigins() []string {
	origins := strings.TrimSpace(os.Getenv("CORS_ALLOW_ORIGINS"))
	if origins == "" {
		return []string{"*"}
	}

	parts := strings.Split(origins, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		v := strings.TrimSpace(p)
		if v != "" {
			out = append(out, v)
		}
	}
	if len(out) == 0 {
		return []string{"*"}
	}

	return out
}
