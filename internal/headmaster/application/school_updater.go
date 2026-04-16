package application

import (
	"context"

	"github.com/google/uuid"
)

// SchoolHeadmasterSetter updates schools.headmaster_id to point at the newly
// assigned headmaster user. Implemented by the school infrastructure layer.
type SchoolHeadmasterSetter interface {
	SetHeadmasterID(ctx context.Context, schoolID, headmasterUserID uuid.UUID) error
}
