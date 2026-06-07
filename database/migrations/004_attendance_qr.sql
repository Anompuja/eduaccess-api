-- Add "late" status for QR-based attendance marking.
-- Also consolidate the constraint to match all statuses used in code.
ALTER TABLE class_schedule_students
  DROP CONSTRAINT IF EXISTS class_schedule_students_status_check;

ALTER TABLE class_schedule_students
  ADD CONSTRAINT class_schedule_students_status_check
  CHECK (status IN ('scheduled', 'present', 'absent', 'late', 'permission', 'sick'));
