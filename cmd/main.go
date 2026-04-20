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
	adminApp "github.com/eduaccess/eduaccess-api/internal/admin/application"
	adminHTTP "github.com/eduaccess/eduaccess-api/internal/admin/delivery/http"
	adminInfra "github.com/eduaccess/eduaccess-api/internal/admin/infrastructure"
	authApp "github.com/eduaccess/eduaccess-api/internal/auth/application"
	authHTTP "github.com/eduaccess/eduaccess-api/internal/auth/delivery/http"
	authInfra "github.com/eduaccess/eduaccess-api/internal/auth/infrastructure"
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
	studentApp "github.com/eduaccess/eduaccess-api/internal/student/application"
	studentHTTP "github.com/eduaccess/eduaccess-api/internal/student/delivery/http"
	studentInfra "github.com/eduaccess/eduaccess-api/internal/student/infrastructure"
	userApp "github.com/eduaccess/eduaccess-api/internal/user/application"
	userHTTP "github.com/eduaccess/eduaccess-api/internal/user/delivery/http"
	userInfra "github.com/eduaccess/eduaccess-api/internal/user/infrastructure"
	"github.com/eduaccess/eduaccess-api/pkg/database"
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
// @description Type "Bearer" followed by a space and JWT token.
func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, reading from environment")
	}

	db, err := database.Connect()
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}

	e := echo.New()
	e.HideBanner = true
	e.Validator = appvalidator.New()

	// Middleware
	e.Use(middleware.Recover())
	e.Use(middleware.Logger())
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: getAllowedOrigins(),
		AllowMethods: []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete, http.MethodOptions},
		AllowHeaders: []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAccept, echo.HeaderAuthorization, echo.HeaderXRequestedWith},
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

	// ── Auth module ───────────────────────────────────────────────────────────
	userRepo := authInfra.NewGormUserRepository(db)
	rtRepo := authInfra.NewGormRefreshTokenRepository(db)
	authHTTP.NewHandler(
		v1,
		authApp.NewRegisterHandler(userRepo),
		authApp.NewLoginHandler(userRepo, rtRepo),
		authApp.NewRefreshHandler(userRepo, rtRepo),
		authApp.NewLogoutHandler(rtRepo),
	)

	// ── User management module ────────────────────────────────────────────────
	userMgmtRepo := userInfra.NewGormUserManagementRepository(db)
	userHTTP.NewHandler(
		v1,
		userApp.NewListUsersHandler(userMgmtRepo),
		userApp.NewGetUserHandler(userMgmtRepo),
		userApp.NewUpdateUserHandler(userMgmtRepo),
		userApp.NewDeactivateUserHandler(userMgmtRepo),
		userApp.NewChangePasswordHandler(userMgmtRepo),
	)

	// ── School setup module ───────────────────────────────────────────────────
	schoolRepo := schoolInfra.NewGormSchoolRepository(db)
	schoolHTTP.NewHandler(
		v1,
		schoolApp.NewCreateSchoolHandler(schoolRepo),
		schoolApp.NewListSchoolsHandler(schoolRepo),
		schoolApp.NewGetSchoolHandler(schoolRepo),
		schoolApp.NewUpdateSchoolHandler(schoolRepo),
		schoolApp.NewDeactivateSchoolHandler(schoolRepo),
		schoolApp.NewListRulesHandler(schoolRepo),
		schoolApp.NewUpsertRulesHandler(schoolRepo),
		schoolApp.NewGetSubscriptionHandler(schoolRepo),
	)

	// ── Headmaster module ─────────────────────────────────────────────────────
	headmasterRepo := headmasterInfra.NewGormHeadmasterRepository(db)
	headmasterHTTP.NewHandler(
		v1,
		headmasterApp.NewCreateHeadmasterHandler(userRepo, headmasterRepo, schoolRepo),
		headmasterApp.NewListHeadmastersHandler(headmasterRepo),
		headmasterApp.NewGetHeadmasterHandler(headmasterRepo),
		headmasterApp.NewUpdateHeadmasterHandler(headmasterRepo),
		headmasterApp.NewDeactivateHeadmasterHandler(headmasterRepo),
	)

	// ── Student module ────────────────────────────────────────────────────────
	studentRepo := studentInfra.NewGormStudentRepository(db)
	academicRepo := studentInfra.NewGormAcademicRepository(db)
	studentHTTP.NewHandler(
		v1,
		studentApp.NewCreateStudentHandler(userRepo, studentRepo, academicRepo),
		studentApp.NewListStudentsHandler(studentRepo),
		studentApp.NewGetStudentHandler(studentRepo),
		studentApp.NewUpdateStudentHandler(studentRepo),
		studentApp.NewDeactivateStudentHandler(studentRepo),
		studentApp.NewLinkParentHandler(studentRepo),
		studentApp.NewUnlinkParentHandler(studentRepo),
		studentApp.NewCreateParentHandler(userRepo, studentRepo),
		studentApp.NewListParentsHandler(studentRepo),
		studentApp.NewGetParentHandler(studentRepo),
		studentApp.NewUpdateParentHandler(studentRepo),
		studentApp.NewDeactivateParentHandler(studentRepo),
		studentApp.NewCreateLevelHandler(academicRepo),
		studentApp.NewListLevelsHandler(academicRepo),
		studentApp.NewUpdateLevelHandler(academicRepo),
		studentApp.NewDeleteLevelHandler(academicRepo),
		studentApp.NewCreateClassHandler(academicRepo),
		studentApp.NewListClassesHandler(academicRepo),
		studentApp.NewUpdateClassHandler(academicRepo),
		studentApp.NewDeleteClassHandler(academicRepo),
		studentApp.NewCreateSubClassHandler(academicRepo),
		studentApp.NewListSubClassesHandler(academicRepo),
		studentApp.NewUpdateSubClassHandler(academicRepo),
		studentApp.NewDeleteSubClassHandler(academicRepo),
	)

	// Parent module
	parentRepo := parentInfra.NewGormParentRepository(db)
	parentHTTP.NewHandler(
		v1,
		parentApp.NewListParentsHandler(parentRepo),
	)

	// ── Admin module ──────────────────────────────────────────────────────────
	adminRepo := adminInfra.NewGormAdminRepository(db)
	adminHTTP.NewHandler(
		v1,
		adminApp.NewCreateAdminHandler(userRepo, adminRepo),
		adminApp.NewGetAdminHandler(adminRepo),
		adminApp.NewListAdminsHandler(adminRepo),
		adminApp.NewUpdateAdminHandler(userMgmtRepo, adminRepo),
		adminApp.NewDeactivateAdminHandler(adminRepo),
	)

	port := os.Getenv("APP_PORT")
	if port == "" {
		port = "8080"
	}

	// Graceful shutdown
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	go func() {
		addr := fmt.Sprintf(":%s", port)
		log.Printf(" Server is running at http://localhost:%s", port)
		log.Printf(" Swagger documentation at http://localhost:%s/swagger/index.html", port)
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
