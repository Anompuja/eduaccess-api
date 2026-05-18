package infrastructure

import (
	"context"
	"errors"
	"fmt"
	"math"
	"time"

	"github.com/eduaccess/eduaccess-api/internal/dashboard/domain"
	"github.com/eduaccess/eduaccess-api/internal/shared/apperror"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// GormDashboardRepository aggregates dashboard metrics via SQL.
type GormDashboardRepository struct {
	db *gorm.DB
}

// NewGormDashboardRepository creates a dashboard repository.
func NewGormDashboardRepository(db *gorm.DB) *GormDashboardRepository {
	return &GormDashboardRepository{db: db}
}

type schoolSnapshotRow struct {
	ID       uuid.UUID `gorm:"column:id"`
	Name     string    `gorm:"column:name"`
	Status   string    `gorm:"column:status"`
	TimeZone string    `gorm:"column:time_zone"`
}

type subscriptionSnapshotRow struct {
	PlanName string     `gorm:"column:plan_name"`
	Status   string     `gorm:"column:status"`
	Cycle    string     `gorm:"column:cycle"`
	Price    int64      `gorm:"column:price"`
	EndsAt   *time.Time `gorm:"column:ends_at"`
}

// GetStats loads the aggregated dashboard payload.
// schoolID = nil → aggregate across all schools (superadmin "Semua Sekolah" view).
// schoolID = non-nil → scoped to that school.
func (r *GormDashboardRepository) GetStats(ctx context.Context, schoolID *uuid.UUID) (*domain.Stats, error) {
	stats := &domain.Stats{}

	// ── School summary ────────────────────────────────────────────────────────
	if schoolID != nil {
		var school schoolSnapshotRow
		if err := r.db.WithContext(ctx).
			Raw(`SELECT id, name, status, time_zone FROM schools WHERE id = ? AND deleted_at IS NULL LIMIT 1`, *schoolID).
			Scan(&school).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, apperror.New(apperror.ErrNotFound, "school not found")
			}
			return nil, err
		}
		if school.ID == uuid.Nil {
			return nil, apperror.New(apperror.ErrNotFound, "school not found")
		}
		stats.School = domain.SchoolSummary{
			ID:       school.ID,
			Name:     school.Name,
			Status:   school.Status,
			TimeZone: school.TimeZone,
		}
	} else {
		stats.School = domain.SchoolSummary{
			ID:       uuid.Nil,
			Name:     "Semua Sekolah",
			Status:   "active",
			TimeZone: "",
		}
	}

	// ── Helper: appends `AND <alias>.school_id = ?` only when filtering one school ─
	scopeFilter := func(alias string) (string, []any) {
		if schoolID == nil {
			return "", nil
		}
		return fmt.Sprintf(" AND %s.school_id = ?", alias), []any{*schoolID}
	}

	count := func(dest any, query string, args ...any) error {
		return r.db.WithContext(ctx).Raw(query, args...).Scan(dest).Error
	}

	// ── Counts ────────────────────────────────────────────────────────────────
	if clause, args := scopeFilter("su"); true {
		if err := count(&stats.Counts.SchoolUsers,
			`SELECT COUNT(DISTINCT su.user_id) FROM school_users su WHERE su.deleted_at IS NULL`+clause, args...); err != nil {
			return nil, err
		}
	}
	if clause, args := scopeFilter("su"); true {
		if err := count(&stats.Counts.Admins,
			`SELECT COUNT(DISTINCT u.id)
FROM users u
JOIN school_users su ON su.user_id = u.id AND su.deleted_at IS NULL
JOIN model_has_roles mhr ON mhr.user_id = u.id
JOIN roles r ON r.id = mhr.role_id
WHERE u.deleted_at IS NULL AND r.name = 'admin_sekolah'`+clause, args...); err != nil {
			return nil, err
		}
	}
	if clause, args := scopeFilter("hp"); true {
		if err := count(&stats.Counts.Headmasters,
			`SELECT COUNT(DISTINCT hp.user_id) FROM headmaster_profiles hp WHERE hp.deleted_at IS NULL`+clause, args...); err != nil {
			return nil, err
		}
	}
	if clause, args := scopeFilter("tp"); true {
		if err := count(&stats.Counts.Teachers,
			`SELECT COUNT(DISTINCT tp.user_id) FROM teacher_profiles tp WHERE tp.deleted_at IS NULL`+clause, args...); err != nil {
			return nil, err
		}
	}
	if clause, args := scopeFilter("sp"); true {
		if err := count(&stats.Counts.Staff,
			`SELECT COUNT(DISTINCT sp.user_id) FROM staff_profiles sp WHERE sp.deleted_at IS NULL`+clause, args...); err != nil {
			return nil, err
		}
	}
	if clause, args := scopeFilter("pp"); true {
		if err := count(&stats.Counts.Parents,
			`SELECT COUNT(DISTINCT pp.user_id) FROM parent_profiles pp WHERE pp.deleted_at IS NULL`+clause, args...); err != nil {
			return nil, err
		}
	}
	if clause, args := scopeFilter("sp"); true {
		if err := count(&stats.Counts.Students,
			`SELECT COUNT(DISTINCT sp.user_id) FROM student_profiles sp WHERE sp.deleted_at IS NULL`+clause, args...); err != nil {
			return nil, err
		}
	}
	if clause, args := scopeFilter("ss"); true {
		if err := count(&stats.Counts.ActiveStudents,
			`SELECT COUNT(DISTINCT ss.student_id) FROM student_studies ss WHERE ss.deleted_at IS NULL AND ss.status = 'active'`+clause, args...); err != nil {
			return nil, err
		}
	}
	if clause, args := scopeFilter("ss"); true {
		if err := count(&stats.Counts.Enrollments,
			`SELECT COUNT(*) FROM student_studies ss WHERE ss.deleted_at IS NULL`+clause, args...); err != nil {
			return nil, err
		}
	}
	if clause, args := scopeFilter("ss"); true {
		if err := count(&stats.Counts.ActiveEnrollments,
			`SELECT COUNT(*) FROM student_studies ss WHERE ss.deleted_at IS NULL AND ss.status = 'active'`+clause, args...); err != nil {
			return nil, err
		}
	}
	if clause, args := scopeFilter("say"); true {
		if err := count(&stats.Counts.AcademicYears,
			`SELECT COUNT(*) FROM school_academic_years say WHERE say.deleted_at IS NULL`+clause, args...); err != nil {
			return nil, err
		}
	}
	if clause, args := scopeFilter("sel"); true {
		if err := count(&stats.Counts.EducationLevels,
			`SELECT COUNT(*) FROM school_education_levels sel WHERE sel.deleted_at IS NULL`+clause, args...); err != nil {
			return nil, err
		}
	}
	if clause, args := scopeFilter("sc"); true {
		if err := count(&stats.Counts.Classes,
			`SELECT COUNT(*) FROM school_classes sc WHERE sc.deleted_at IS NULL`+clause, args...); err != nil {
			return nil, err
		}
	}
	if clause, args := scopeFilter("ssc"); true {
		if err := count(&stats.Counts.SubClasses,
			`SELECT COUNT(*) FROM school_sub_classes ssc WHERE ssc.deleted_at IS NULL`+clause, args...); err != nil {
			return nil, err
		}
	}
	if clause, args := scopeFilter("sc"); true {
		if err := count(&stats.Counts.Classrooms,
			`SELECT COUNT(*) FROM school_classrooms sc WHERE sc.deleted_at IS NULL`+clause, args...); err != nil {
			return nil, err
		}
	}
	if clause, args := scopeFilter("ss"); true {
		if err := count(&stats.Counts.Subjects,
			`SELECT COUNT(*) FROM school_subjects ss WHERE ss.deleted_at IS NULL`+clause, args...); err != nil {
			return nil, err
		}
	}
	if clause, args := scopeFilter("cs"); true {
		if err := count(&stats.Counts.Schedules,
			`SELECT COUNT(*) FROM class_schedules cs WHERE cs.deleted_at IS NULL`+clause, args...); err != nil {
			return nil, err
		}
	}

	// ── Today's attendance ────────────────────────────────────────────────────
	var attendance struct {
		Present int64 `gorm:"column:present"`
		Late    int64 `gorm:"column:late"`
		Absent  int64 `gorm:"column:absent"`
		Excused int64 `gorm:"column:excused"`
		Total   int64 `gorm:"column:total"`
	}
	attClause, attArgs := "", []any{}
	if schoolID != nil {
		attClause = " AND school_id = ?"
		attArgs = []any{*schoolID}
	}
	if err := r.db.WithContext(ctx).Raw(`
SELECT
	COUNT(*) FILTER (WHERE status = 'present') AS present,
	COUNT(*) FILTER (WHERE status = 'late') AS late,
	COUNT(*) FILTER (WHERE status = 'absent') AS absent,
	COUNT(*) FILTER (WHERE status = 'excused') AS excused,
	COUNT(*) AS total
FROM school_attendances
WHERE deleted_at IS NULL AND date = CURRENT_DATE`+attClause, attArgs...).Scan(&attendance).Error; err != nil {
		return nil, err
	}
	stats.Attendance = domain.AttendanceSummary{
		Present: attendance.Present,
		Late:    attendance.Late,
		Absent:  attendance.Absent,
		Excused: attendance.Excused,
		Total:   attendance.Total,
	}
	if attendance.Total > 0 {
		stats.Attendance.Rate = math.Round(((float64(attendance.Present+attendance.Late)/float64(attendance.Total))*100.0)*100) / 100
	}

	// ── Subscription ──────────────────────────────────────────────────────────
	// Only meaningful for a specific school; skip in aggregate view.
	if schoolID != nil {
		var sub subscriptionSnapshotRow
		if err := r.db.WithContext(ctx).Raw(`
SELECT p.name AS plan_name, sub.status, sub.cycle, sub.price, sub.ends_at
FROM subscriptions sub
JOIN plans p ON p.id = sub.plan_id
WHERE sub.school_id = ? AND sub.status IN ('active','trial')
ORDER BY sub.created_at DESC
LIMIT 1`, *schoolID).Scan(&sub).Error; err != nil {
			return nil, err
		}
		if sub.PlanName != "" || sub.Status != "" {
			stats.Subscription = &domain.SubscriptionSummary{
				PlanName: sub.PlanName,
				Status:   sub.Status,
				Cycle:    sub.Cycle,
				Price:    sub.Price,
				EndsAt:   sub.EndsAt,
			}
		}
	}

	return stats, nil
}
