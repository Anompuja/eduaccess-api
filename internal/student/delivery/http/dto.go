package http

import "time"

type StudentResponse struct {
	ID                string               `json:"id"`
	UserID            string               `json:"user_id"`
	SchoolID          string               `json:"school_id"`
	Name              string               `json:"name"`
	Email             string               `json:"email"`
	Username          string               `json:"username"`
	Avatar            string               `json:"avatar"`
	NIS               string               `json:"nis"`
	NISN              string               `json:"nisn"`
	PhoneNumber       string               `json:"phone_number"`
	Address           string               `json:"address"`
	Gender            string               `json:"gender"`
	Religion          string               `json:"religion"`
	BirthPlace        string               `json:"birth_place"`
	BirthDate         *time.Time           `json:"birth_date,omitempty"`
	TahunMasuk        string               `json:"tahun_masuk"`
	JalurMasukSekolah string               `json:"jalur_masuk_sekolah"`
	EducationLevelID  *string              `json:"education_level_id,omitempty"`
	ClassID           *string              `json:"class_id,omitempty"`
	SubClassID        *string              `json:"sub_class_id,omitempty"`
	Parents           []ParentLinkResponse `json:"parents,omitempty"`
	CreatedAt         time.Time            `json:"created_at"`
	UpdatedAt         time.Time            `json:"updated_at"`
}

type CreateStudentRequest struct {
	Name              string  `json:"name"                validate:"required,min=2,max=191"`
	Email             string  `json:"email"               validate:"required,email,max=191"`
	Username          string  `json:"username"            validate:"omitempty,min=3,max=50"`
	Password          string  `json:"password"            validate:"omitempty,min=8"`
	NIS               string  `json:"nis"                 validate:"omitempty,max=191"`
	NISN              string  `json:"nisn"                validate:"omitempty,max=191"`
	PhoneNumber       string  `json:"phone_number"        validate:"omitempty,max=50"`
	Address           string  `json:"address"             validate:"omitempty"`
	Gender            string  `json:"gender"              validate:"omitempty,oneof=L P"`
	Religion          string  `json:"religion"            validate:"omitempty,max=100"`
	BirthPlace        string  `json:"birth_place"         validate:"omitempty,max=191"`
	BirthDate         *string `json:"birth_date"          validate:"omitempty"`
	TahunMasuk        string  `json:"tahun_masuk"         validate:"omitempty,max=10"`
	JalurMasukSekolah string  `json:"jalur_masuk_sekolah" validate:"omitempty,oneof=reguler beasiswa mutasi lainnya"`
	EducationLevelID  *string `json:"education_level_id"  validate:"omitempty,uuid"`
	ClassID           *string `json:"class_id"            validate:"omitempty,uuid"`
	SubClassID        *string `json:"sub_class_id"        validate:"omitempty,uuid"`
}

type UpdateStudentRequest struct {
	NIS               *string `json:"nis"                 validate:"omitempty,max=191"`
	NISN              *string `json:"nisn"                validate:"omitempty,max=191"`
	PhoneNumber       *string `json:"phone_number"        validate:"omitempty,max=50"`
	Address           *string `json:"address"             validate:"omitempty"`
	Gender            *string `json:"gender"              validate:"omitempty,oneof=L P"`
	Religion          *string `json:"religion"            validate:"omitempty,max=100"`
	BirthPlace        *string `json:"birth_place"         validate:"omitempty,max=191"`
	BirthDate         *string `json:"birth_date"          validate:"omitempty"`
	TahunMasuk        *string `json:"tahun_masuk"         validate:"omitempty,max=10"`
	JalurMasukSekolah *string `json:"jalur_masuk_sekolah" validate:"omitempty,oneof=reguler beasiswa mutasi lainnya"`
	EducationLevelID  *string `json:"education_level_id"  validate:"omitempty,uuid"`
	ClassID           *string `json:"class_id"            validate:"omitempty,uuid"`
	SubClassID        *string `json:"sub_class_id"        validate:"omitempty,uuid"`
}

type ParentResponse struct {
	ID             string    `json:"id"`
	UserID         string    `json:"user_id"`
	SchoolID       string    `json:"school_id"`
	Name           string    `json:"name"`
	Email          string    `json:"email"`
	Username       string    `json:"username"`
	Avatar         string    `json:"avatar"`
	FatherName     string    `json:"father_name"`
	MotherName     string    `json:"mother_name"`
	FatherReligion string    `json:"father_religion"`
	MotherReligion string    `json:"mother_religion"`
	PhoneNumber    string    `json:"phone_number"`
	Address        string    `json:"address"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

type CreateParentRequest struct {
	Name           string `json:"name"            validate:"required,min=2,max=191"`
	Email          string `json:"email"           validate:"required,email,max=191"`
	Username       string `json:"username"        validate:"omitempty,min=3,max=50"`
	Password       string `json:"password"        validate:"omitempty,min=8"`
	FatherName     string `json:"father_name"     validate:"omitempty,max=191"`
	MotherName     string `json:"mother_name"     validate:"omitempty,max=191"`
	FatherReligion string `json:"father_religion" validate:"omitempty,max=100"`
	MotherReligion string `json:"mother_religion" validate:"omitempty,max=100"`
	PhoneNumber    string `json:"phone_number"    validate:"omitempty,max=50"`
	Address        string `json:"address"         validate:"omitempty"`
}

type UpdateParentRequest struct {
	FatherName     *string `json:"father_name"     validate:"omitempty,max=191"`
	MotherName     *string `json:"mother_name"     validate:"omitempty,max=191"`
	FatherReligion *string `json:"father_religion" validate:"omitempty,max=100"`
	MotherReligion *string `json:"mother_religion" validate:"omitempty,max=100"`
	PhoneNumber    *string `json:"phone_number"    validate:"omitempty,max=50"`
	Address        *string `json:"address"         validate:"omitempty"`
}

type ParentLinkResponse struct {
	ID           string          `json:"id"`
	ParentID     string          `json:"parent_id"`
	Relationship string          `json:"relationship"`
	IsPrimary    bool            `json:"is_primary"`
	Parent       *ParentResponse `json:"parent,omitempty"`
}

type LinkParentRequest struct {
	ParentID     string `json:"parent_id"    validate:"required,uuid"`
	Relationship string `json:"relationship" validate:"required,oneof=father mother guardian other"`
	IsPrimary    bool   `json:"is_primary"`
}
