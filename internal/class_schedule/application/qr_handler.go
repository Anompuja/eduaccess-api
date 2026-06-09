package application

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/eduaccess/eduaccess-api/internal/class_schedule/domain"
	notificationApp "github.com/eduaccess/eduaccess-api/internal/notification/application"
	"github.com/eduaccess/eduaccess-api/internal/shared/apperror"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// qrClaims is the JWT payload embedded in every rotating QR code token.
type qrClaims struct {
	ClassScheduleID string `json:"cid"`
	SchoolID        string `json:"sid"`
	jwt.RegisteredClaims
}

// attendanceSecret is loaded once from ATTENDANCE_SECRET env var.
// If absent, a per-process random secret is generated (30-second tokens won't
// survive a restart, which is acceptable for this use case).
var attendanceSecret []byte

// lateTolerance is the grace period (minutes) after class start before a scan
// is considered "late". Loaded from LATE_TOLERANCE_MINUTES (default 15).
var lateTolerance int

func init() {
	secret := os.Getenv("ATTENDANCE_SECRET")
	if secret == "" {
		secret = uuid.New().String() + uuid.New().String()
	}
	attendanceSecret = []byte(secret)

	tol, err := strconv.Atoi(os.Getenv("LATE_TOLERANCE_MINUTES"))
	if err != nil || tol <= 0 {
		tol = 15
	}
	lateTolerance = tol
}

// ── Generate QR ──────────────────────────────────────────────────────────────

// GenerateQRCommand requests a fresh 30-second QR token for an ongoing class.
type GenerateQRCommand struct {
	RequesterSchoolID *uuid.UUID
	RequesterRole     string
	ScheduleID        uuid.UUID
}

// GenerateQRResult holds the signed token and its TTL in seconds.
type GenerateQRResult struct {
	Token     string `json:"token"`
	ExpiresIn int    `json:"expires_in"`
}

// GenerateQRHandler produces a short-lived HS256 JWT that encodes the class
// schedule and school. Teachers display this as a QR code; students scan it.
type GenerateQRHandler struct{ repo domain.ClassScheduleRepository }

func NewGenerateQRHandler(repo domain.ClassScheduleRepository) *GenerateQRHandler {
	return &GenerateQRHandler{repo: repo}
}

func (h *GenerateQRHandler) Handle(ctx context.Context, cmd GenerateQRCommand) (*GenerateQRResult, error) {
	if cmd.RequesterRole == "siswa" || cmd.RequesterRole == "orangtua" {
		return nil, apperror.New(apperror.ErrForbidden, "only teachers and staff can generate QR codes")
	}
	cs, err := h.repo.FindClassScheduleByID(ctx, cmd.ScheduleID)
	if err != nil {
		return nil, err
	}
	if err := guardSchoolMatch(cmd.RequesterRole, cmd.RequesterSchoolID, cs.SchoolID); err != nil {
		return nil, err
	}
	if cs.Status != "ongoing" {
		return nil, apperror.New(apperror.ErrBadRequest, "class session must be ongoing to generate a QR code")
	}

	now := time.Now()
	claims := qrClaims{
		ClassScheduleID: cs.ID.String(),
		SchoolID:        cs.SchoolID.String(),
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(30 * time.Second)),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString(attendanceSecret)
	if err != nil {
		return nil, err
	}
	return &GenerateQRResult{Token: signed, ExpiresIn: 30}, nil
}

// ── Scan QR ──────────────────────────────────────────────────────────────────

// ScanQRCommand is submitted by a student after their camera decodes the QR.
type ScanQRCommand struct {
	RequesterSchoolID *uuid.UUID
	RequesterRole     string
	StudentUserID     uuid.UUID // auth.users UUID (JWT sub claim)
	Token             string    // raw QR-JWT string
}

// ScanQRResult reports the recorded attendance status and a friendly message.
type ScanQRResult struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}

// ScanQRHandler verifies the QR token and marks the student's attendance as
// "present" or "late" depending on how far past the class start time they are.
type ScanQRHandler struct {
	repo          domain.ClassScheduleRepository
	notifyParents *notificationApp.NotifyAttendanceParentsHandler
}

func NewScanQRHandler(repo domain.ClassScheduleRepository, notifyParents *notificationApp.NotifyAttendanceParentsHandler) *ScanQRHandler {
	return &ScanQRHandler{repo: repo, notifyParents: notifyParents}
}

func (h *ScanQRHandler) Handle(ctx context.Context, cmd ScanQRCommand) (*ScanQRResult, error) {
	if cmd.RequesterRole != "siswa" {
		return nil, apperror.New(apperror.ErrForbidden, "only students can scan QR codes")
	}
	if cmd.RequesterSchoolID == nil {
		return nil, apperror.New(apperror.ErrForbidden, "student must be enrolled in a school")
	}

	// Parse and verify QR token (HS256 with ATTENDANCE_SECRET).
	var claims qrClaims
	parsed, err := jwt.ParseWithClaims(cmd.Token, &claims, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return attendanceSecret, nil
	})
	if err != nil || !parsed.Valid {
		return nil, apperror.New(apperror.ErrUnauthorized, "QR code has expired or is invalid — ask your teacher to refresh the QR")
	}

	scheduleID, err := uuid.Parse(claims.ClassScheduleID)
	if err != nil {
		return nil, apperror.New(apperror.ErrBadRequest, "malformed QR token")
	}
	qrSchoolID, err := uuid.Parse(claims.SchoolID)
	if err != nil {
		return nil, apperror.New(apperror.ErrBadRequest, "malformed QR token")
	}

	// Prevent cross-school attendance fraud.
	if *cmd.RequesterSchoolID != qrSchoolID {
		return nil, apperror.New(apperror.ErrForbidden, "QR code belongs to a different school")
	}

	cs, err := h.repo.FindClassScheduleByID(ctx, scheduleID)
	if err != nil {
		return nil, err
	}
	if cs.Status != "ongoing" {
		return nil, apperror.New(apperror.ErrBadRequest, "class session is not ongoing")
	}

	// student_id in class_schedule_students is the auth.users UUID.
	att, err := h.repo.FindAttendance(ctx, scheduleID, cmd.StudentUserID)
	if err != nil {
		var appErr *apperror.AppError
		if errors.As(err, &appErr) && appErr.Err == apperror.ErrNotFound {
			return nil, apperror.New(apperror.ErrForbidden, "you are not enrolled in this class session")
		}
		return nil, err
	}

	switch att.Status {
	case "present", "late":
		return nil, apperror.New(apperror.ErrConflict, "attendance already recorded for this session")
	case "permission", "sick":
		return nil, apperror.New(apperror.ErrConflict, "your attendance status was set manually and cannot be overridden by QR scan")
	}

	// Determine present vs late.
	now := time.Now()
	newStatus := resolveAttendanceStatus(cs, now)

	att.Status = newStatus
	att.StudentAttendanceTime = &now
	att.UpdatedAt = now
	if err := h.repo.UpdateAttendance(ctx, att); err != nil {
		return nil, err
	}

	if h.notifyParents != nil {
		go h.notifyParents.Handle(context.Background(), notificationApp.NotifyAttendanceParentsCommand{
			StudentUserID:    cmd.StudentUserID,
			SchoolID:         *cmd.RequesterSchoolID,
			ClassScheduleID:  cs.ID,
			SubjectName:      cs.SubjectName,
			ClassroomName:    cs.ClassroomName,
			AttendanceStatus: newStatus,
			AttendanceTime:   now,
			StudentName:      att.StudentName,
		})
	}

	msg := "Berhasil hadir!"
	if newStatus == "late" {
		msg = "Hadir terlambat"
	}
	return &ScanQRResult{Status: newStatus, Message: msg}, nil
}

// resolveAttendanceStatus returns "present" or "late" based on how far past
// the class start time + lateTolerance minutes the scan occurred.
func resolveAttendanceStatus(cs *domain.ClassSchedule, now time.Time) string {
	var startHour, startMin int
	fmt.Sscanf(cs.StartTime, "%d:%d", &startHour, &startMin)

	deadline := time.Date(
		cs.Date.Year(), cs.Date.Month(), cs.Date.Day(),
		startHour, startMin, 0, 0, cs.Date.Location(),
	).Add(time.Duration(lateTolerance) * time.Minute)

	if now.After(deadline) {
		return "late"
	}
	return "present"
}
