package application

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/eduaccess/eduaccess-api/internal/notification/domain"
	"github.com/google/uuid"
)

// NotificationBroadcaster is satisfied by the WebSocket hub, allowing the
// application layer to push to connected clients without importing delivery code.
type NotificationBroadcaster interface {
	Broadcast(userID uuid.UUID, payload []byte)
}

type NotifyAttendanceParentsCommand struct {
	StudentUserID    uuid.UUID
	SchoolID         uuid.UUID
	ClassScheduleID  uuid.UUID
	SubjectName      string
	ClassroomName    string
	AttendanceStatus string // "present" or "late"
	AttendanceTime   time.Time
	StudentName      string
}

type NotifyAttendanceParentsHandler struct {
	repo      domain.NotificationRepository
	broadcast NotificationBroadcaster
}

func NewNotifyAttendanceParentsHandler(repo domain.NotificationRepository, broadcast NotificationBroadcaster) *NotifyAttendanceParentsHandler {
	return &NotifyAttendanceParentsHandler{repo: repo, broadcast: broadcast}
}

// Handle is designed to be called in a goroutine — it logs errors rather than returning them.
func (h *NotifyAttendanceParentsHandler) Handle(ctx context.Context, cmd NotifyAttendanceParentsCommand) {
	parentUserIDs, err := h.repo.FindParentUserIDsByStudentUserID(ctx, cmd.StudentUserID)
	if err != nil {
		log.Printf("[notification] find parents for student %s: %v", cmd.StudentUserID, err)
		return
	}
	if len(parentUserIDs) == 0 {
		return
	}

	timeStr := cmd.AttendanceTime.Format("15:04")
	title := fmt.Sprintf("Kehadiran %s", cmd.StudentName)

	var body string
	if cmd.AttendanceStatus == "late" {
		body = fmt.Sprintf("%s terlambat masuk kelas %s (%s) pukul %s", cmd.StudentName, cmd.SubjectName, cmd.ClassroomName, timeStr)
	} else {
		body = fmt.Sprintf("%s telah hadir di kelas %s (%s) pukul %s", cmd.StudentName, cmd.SubjectName, cmd.ClassroomName, timeStr)
	}

	data := map[string]any{
		"student_name":      cmd.StudentName,
		"subject_name":      cmd.SubjectName,
		"classroom_name":    cmd.ClassroomName,
		"schedule_id":       cmd.ClassScheduleID.String(),
		"attendance_status": cmd.AttendanceStatus,
		"attendance_time":   cmd.AttendanceTime.Format(time.RFC3339),
	}

	for _, parentUserID := range parentUserIDs {
		n := &domain.Notification{
			SchoolID: &cmd.SchoolID,
			UserID:   parentUserID,
			Type:     "attendance",
			Title:    title,
			Body:     body,
			Data:     data,
		}
		if err := h.repo.Create(ctx, n); err != nil {
			log.Printf("[notification] create notification for parent %s: %v", parentUserID, err)
			continue
		}

		payload, err := json.Marshal(map[string]any{
			"id":        n.ID.String(),
			"type":      n.Type,
			"title":     n.Title,
			"body":      n.Body,
			"data":      n.Data,
			"createdAt": n.CreatedAt.Format(time.RFC3339),
		})
		if err == nil {
			h.broadcast.Broadcast(parentUserID, payload)
		}
	}
}
