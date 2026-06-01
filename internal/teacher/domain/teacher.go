package domain

import (
	"time"

	"github.com/google/uuid"
)

// TeacherProfile represents a teacher's extended profile in the school system.
type TeacherProfile struct {
	ID                                  uuid.UUID
	UserID                              uuid.UUID
	SchoolID                            uuid.UUID
	NIP                                 *string
	NUPTK                               *string
	Name                                string
	Email                               string
	Username                            string
	Avatar                              string
	PhoneNumber                         *string
	Address                             *string
	Gender                              *string
	Religion                            *string
	BirthPlace                          *string
	BirthDate                           *time.Time
	NIK                                 *string
	KTPImagePath                        *string
	Kewarganegaraan                     *string
	GolonganDarah                       *string
	BeratBadan                          *string
	TinggiBadan                         *string
	PenyakitYangSeringKambuh            *string
	KelainanJasmani                     *string
	PenyakitKronisYangPernahDiderita    *string
	RTRW                                *string
	KodePos                             *string
	PendidikanTerakhir                  *string
	Jurusan                             *string
	TahunLulus                          *string
	TahunMasuk                          *string
	DeletedAt                           *time.Time
	CreatedAt                           time.Time
	UpdatedAt                           time.Time
}

// IsActive returns true if the teacher profile has not been soft-deleted.
func (t *TeacherProfile) IsActive() bool {
	return t.DeletedAt == nil
}
