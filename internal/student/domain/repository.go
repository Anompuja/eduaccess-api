package domain

import (
	"context"

	"github.com/google/uuid"
)

// ParentLinkRepository handles student-to-parent relationship persistence.
type ParentLinkRepository interface {
	LinkParent(ctx context.Context, link *ParentLink) error
	UnlinkParent(ctx context.Context, studentID, parentID uuid.UUID) error
	ListParentLinks(ctx context.Context, studentID uuid.UUID) ([]*ParentLink, error)
}

// StudentRepository handles student profiles and their parent links.
type StudentRepository interface {
	StudentProfileRepository
	ParentLinkRepository
}
