package infrastructure

import (
	"context"
	"errors"
	"time"

	"github.com/eduaccess/eduaccess-api/internal/teacher/domain"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// teacherProfileModel represents the ORM model for teacher profiles.
type teacherProfileModel struct {
	ID                                  uuid.UUID  `gorm:"column:id;type:uuid;primaryKey"`
	UserID                              uuid.UUID  `gorm:"column:user_id;type:uuid;index"`
	SchoolID                            uuid.UUID  `gorm:"column:school_id;type:uuid;index"`
	NIP                                 *string    `gorm:"column:nip;type:varchar(191)"`
	NUPTK                               *string    `gorm:"column:nuptk;type:varchar(191)"`
	PhoneNumber                         *string    `gorm:"column:phone_number;type:varchar(50)"`
	Address                             *string    `gorm:"column:address;type:text"`
	Gender                              *string    `gorm:"column:gender;type:varchar(50)"`
	Religion                            *string    `gorm:"column:religion;type:varchar(100)"`
	BirthPlace                          *string    `gorm:"column:birth_place;type:varchar(191)"`
	BirthDate                           *time.Time `gorm:"column:birth_date;type:date"`
	NIK                                 *string    `gorm:"column:nik;type:varchar(50)"`
	KTPImagePath                        *string    `gorm:"column:ktp_image_path;type:varchar(191)"`
	Kewarganegaraan                     *string    `gorm:"column:kewarganegaraan;type:varchar(100)"`
	GolonganDarah                       *string    `gorm:"column:golongan_darah;type:varchar(10)"`
	BeratBadan                          *string    `gorm:"column:berat_badan;type:varchar(20)"`
	TinggiBadan                         *string    `gorm:"column:tinggi_badan;type:varchar(20)"`
	PenyakitYangSeringKambuh            *string    `gorm:"column:penyakit_yang_sering_kambuh;type:text"`
	KelainanJasmani                     *string    `gorm:"column:kelainan_jasmani;type:text"`
	PenyakitKronisYangPernahDiderita    *string    `gorm:"column:penyakit_kronis_yang_pernah_diderita;type:text"`
	RTRW                                *string    `gorm:"column:rt_rw;type:varchar(50)"`
	KodePos                             *string    `gorm:"column:kode_pos;type:varchar(20)"`
	PendidikanTerakhir                  *string    `gorm:"column:pendidikan_terakhir;type:varchar(100)"`
	Jurusan                             *string    `gorm:"column:jurusan;type:varchar(191)"`
	TahunLulus                          *string    `gorm:"column:tahun_lulus;type:varchar(10)"`
	TahunMasuk                          *string    `gorm:"column:tahun_masuk;type:varchar(10)"`
	DeletedAt                           *time.Time `gorm:"column:deleted_at;type:timestamptz;index"`
	CreatedAt                           time.Time  `gorm:"column:created_at;type:timestamptz;autoCreateTime"`
	UpdatedAt                           time.Time  `gorm:"column:updated_at;type:timestamptz;autoUpdateTime"`
}

// TableName specifies the table name for the ORM.
func (t *teacherProfileModel) TableName() string {
	return "teacher_profiles"
}

// teacherWithUser is used for JOIN queries to include user data.
type teacherWithUser struct {
	teacherProfileModel
	Name     string
	Email    string
	Username string
	Avatar   string
}

// TeacherRepository implements the domain repository interface for teachers.
type TeacherRepository struct {
	db *gorm.DB
}

// NewTeacherRepository creates a new TeacherRepository.
func NewTeacherRepository(db *gorm.DB) *TeacherRepository {
	return &TeacherRepository{db: db}
}

// CreateTeacherProfile creates a new teacher profile in the database.
func (r *TeacherRepository) CreateTeacherProfile(ctx context.Context, teacher *domain.TeacherProfile) error {
	model := &teacherProfileModel{
		ID:                                  teacher.ID,
		UserID:                              teacher.UserID,
		SchoolID:                            teacher.SchoolID,
		NIP:                                 teacher.NIP,
		NUPTK:                               teacher.NUPTK,
		PhoneNumber:                         teacher.PhoneNumber,
		Address:                             teacher.Address,
		Gender:                              teacher.Gender,
		Religion:                            teacher.Religion,
		BirthPlace:                          teacher.BirthPlace,
		BirthDate:                           teacher.BirthDate,
		NIK:                                 teacher.NIK,
		KTPImagePath:                        teacher.KTPImagePath,
		Kewarganegaraan:                     teacher.Kewarganegaraan,
		GolonganDarah:                       teacher.GolonganDarah,
		BeratBadan:                          teacher.BeratBadan,
		TinggiBadan:                         teacher.TinggiBadan,
		PenyakitYangSeringKambuh:            teacher.PenyakitYangSeringKambuh,
		KelainanJasmani:                     teacher.KelainanJasmani,
		PenyakitKronisYangPernahDiderita:    teacher.PenyakitKronisYangPernahDiderita,
		RTRW:                                teacher.RTRW,
		KodePos:                             teacher.KodePos,
		PendidikanTerakhir:                  teacher.PendidikanTerakhir,
		Jurusan:                             teacher.Jurusan,
		TahunLulus:                          teacher.TahunLulus,
		TahunMasuk:                          teacher.TahunMasuk,
	}

	return r.db.WithContext(ctx).Create(model).Error
}

// FindTeacherByID retrieves a teacher by ID.
func (r *TeacherRepository) FindTeacherByID(ctx context.Context, id uuid.UUID) (*domain.TeacherProfile, error) {
	var model teacherProfileModel
	result := r.db.WithContext(ctx).Where("id = ? AND deleted_at IS NULL", id).First(&model)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, gorm.ErrRecordNotFound
		}
		return nil, result.Error
	}

	// Fetch associated user data
	var user struct {
		Name     string
		Email    string
		Username string
		Avatar   string
	}
	if err := r.db.WithContext(ctx).Table("users").Where("id = ?", model.UserID).
		Select("name, email, username, avatar").First(&user).Error; err != nil {
		return nil, err
	}

	return r.modelToDomain(&model, user.Name, user.Email, user.Username, user.Avatar), nil
}

// UpdateTeacherProfile updates an existing teacher profile.
func (r *TeacherRepository) UpdateTeacherProfile(ctx context.Context, teacher *domain.TeacherProfile) error {
	model := &teacherProfileModel{
		ID:                                  teacher.ID,
		UserID:                              teacher.UserID,
		SchoolID:                            teacher.SchoolID,
		NIP:                                 teacher.NIP,
		NUPTK:                               teacher.NUPTK,
		PhoneNumber:                         teacher.PhoneNumber,
		Address:                             teacher.Address,
		Gender:                              teacher.Gender,
		Religion:                            teacher.Religion,
		BirthPlace:                          teacher.BirthPlace,
		BirthDate:                           teacher.BirthDate,
		NIK:                                 teacher.NIK,
		KTPImagePath:                        teacher.KTPImagePath,
		Kewarganegaraan:                     teacher.Kewarganegaraan,
		GolonganDarah:                       teacher.GolonganDarah,
		BeratBadan:                          teacher.BeratBadan,
		TinggiBadan:                         teacher.TinggiBadan,
		PenyakitYangSeringKambuh:            teacher.PenyakitYangSeringKambuh,
		KelainanJasmani:                     teacher.KelainanJasmani,
		PenyakitKronisYangPernahDiderita:    teacher.PenyakitKronisYangPernahDiderita,
		RTRW:                                teacher.RTRW,
		KodePos:                             teacher.KodePos,
		PendidikanTerakhir:                  teacher.PendidikanTerakhir,
		Jurusan:                             teacher.Jurusan,
		TahunLulus:                          teacher.TahunLulus,
		TahunMasuk:                          teacher.TahunMasuk,
	}

	return r.db.WithContext(ctx).Model(&teacherProfileModel{}).Where("id = ?", teacher.ID).
		Updates(model).Error
}

// SoftDeleteTeacher soft-deletes a teacher profile by setting deleted_at.
func (r *TeacherRepository) SoftDeleteTeacher(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Model(&teacherProfileModel{}).Where("id = ?", id).
		Update("deleted_at", time.Now()).Error
}

// ListTeachers retrieves a paginated list of teachers with optional filtering.
func (r *TeacherRepository) ListTeachers(ctx context.Context, filter domain.TeacherFilter) ([]*domain.TeacherProfile, int64, error) {
	var models []teacherProfileModel
	var total int64

	query := r.db.WithContext(ctx).
		Where("teacher_profiles.deleted_at IS NULL")
	if filter.SchoolID != nil {
		query = query.Where("teacher_profiles.school_id = ?", *filter.SchoolID)
	}

	// Apply search filter
	if filter.Search != "" {
		query = query.Where("(u.name ILIKE ? OR u.email ILIKE ? OR u.username ILIKE ?)",
			"%"+filter.Search+"%", "%"+filter.Search+"%", "%"+filter.Search+"%")
	}

	// Get total count
	if err := query.Model(&teacherProfileModel{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Fetch paginated results with user data via JOIN
	result := query.
		Joins("LEFT JOIN users u ON teacher_profiles.user_id = u.id").
		Select("teacher_profiles.*, u.name, u.email, u.username, u.avatar").
		Offset(filter.Offset).
		Limit(filter.Limit).
		Order("teacher_profiles.created_at DESC").
		Find(&models)

	if result.Error != nil {
		return nil, 0, result.Error
	}

	// Convert models to domain entities
	teachers := make([]*domain.TeacherProfile, len(models))
	for i, model := range models {
		var user struct {
			Name     string
			Email    string
			Username string
			Avatar   string
		}
		_ = r.db.WithContext(ctx).Table("users").Where("id = ?", model.UserID).
			Select("name, email, username, avatar").First(&user).Error

		teachers[i] = r.modelToDomain(&model, user.Name, user.Email, user.Username, user.Avatar)
	}

	return teachers, total, nil
}

// modelToDomain converts a model to a domain entity.
func (r *TeacherRepository) modelToDomain(m *teacherProfileModel, name, email, username, avatar string) *domain.TeacherProfile {
	return &domain.TeacherProfile{
		ID:                                  m.ID,
		UserID:                              m.UserID,
		SchoolID:                            m.SchoolID,
		NIP:                                 m.NIP,
		NUPTK:                               m.NUPTK,
		Name:                                name,
		Email:                               email,
		Username:                            username,
		Avatar:                              avatar,
		PhoneNumber:                         m.PhoneNumber,
		Address:                             m.Address,
		Gender:                              m.Gender,
		Religion:                            m.Religion,
		BirthPlace:                          m.BirthPlace,
		BirthDate:                           m.BirthDate,
		NIK:                                 m.NIK,
		KTPImagePath:                        m.KTPImagePath,
		Kewarganegaraan:                     m.Kewarganegaraan,
		GolonganDarah:                       m.GolonganDarah,
		BeratBadan:                          m.BeratBadan,
		TinggiBadan:                         m.TinggiBadan,
		PenyakitYangSeringKambuh:            m.PenyakitYangSeringKambuh,
		KelainanJasmani:                     m.KelainanJasmani,
		PenyakitKronisYangPernahDiderita:    m.PenyakitKronisYangPernahDiderita,
		RTRW:                                m.RTRW,
		KodePos:                             m.KodePos,
		PendidikanTerakhir:                  m.PendidikanTerakhir,
		Jurusan:                             m.Jurusan,
		TahunLulus:                          m.TahunLulus,
		TahunMasuk:                          m.TahunMasuk,
		DeletedAt:                           m.DeletedAt,
		CreatedAt:                           m.CreatedAt,
		UpdatedAt:                           m.UpdatedAt,
	}
}
