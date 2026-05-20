package http

import "time"

// CreateTeacherRequest represents the request body for creating a teacher.
type CreateTeacherRequest struct {
	Name                            string  `json:"name" validate:"required,min=1,max=191"`
	Email                           string  `json:"email" validate:"required,email"`
	Username                        string  `json:"username" validate:"required,min=3,max=191"`
	Password                        string  `json:"password" validate:"required,min=8"`
	SchoolID                        *string `json:"school_id" validate:"omitempty,uuid"`
	NIP                             *string `json:"nip"`
	NUPTK                           *string `json:"nuptk"`
	PhoneNumber                     *string `json:"phone_number"`
	Address                         *string `json:"address"`
	Gender                          *string `json:"gender" validate:"omitempty,oneof=male female other"`
	Religion                        *string `json:"religion"`
	BirthPlace                      *string `json:"birth_place"`
	BirthDate                       *string `json:"birth_date"`
	NIK                             *string `json:"nik"`
	KTPImagePath                    *string `json:"ktp_image_path"`
	Kewarganegaraan                 *string `json:"kewarganegaraan"`
	GolonganDarah                   *string `json:"golongan_darah"`
	BeratBadan                      *string `json:"berat_badan"`
	TinggiBadan                     *string `json:"tinggi_badan"`
	PenyakitYangSeringKambuh        *string `json:"penyakit_yang_sering_kambuh"`
	KelainanJasmani                 *string `json:"kelainan_jasmani"`
	PenyakitKronisYangPernahDiderita *string `json:"penyakit_kronis_yang_pernah_diderita"`
	RTRW                            *string `json:"rt_rw"`
	KodePos                         *string `json:"kode_pos"`
	PendidikanTerakhir              *string `json:"pendidikan_terakhir"`
	Jurusan                         *string `json:"jurusan"`
	TahunLulus                      *string `json:"tahun_lulus"`
	TahunMasuk                      *string `json:"tahun_masuk"`
}

// UpdateTeacherRequest represents the request body for updating a teacher.
type UpdateTeacherRequest struct {
	Name                            *string `json:"name" validate:"omitempty,min=1,max=191"`
	Email                           *string `json:"email" validate:"omitempty,email"`
	Username                        *string `json:"username" validate:"omitempty,min=3,max=191"`
	PhoneNumber                     *string `json:"phone_number"`
	Address                         *string `json:"address"`
	Gender                          *string `json:"gender" validate:"omitempty,oneof=male female other"`
	Religion                        *string `json:"religion"`
	BirthPlace                      *string `json:"birth_place"`
	BirthDate                       *string `json:"birth_date"`
	NIK                             *string `json:"nik"`
	KTPImagePath                    *string `json:"ktp_image_path"`
	NIP                             *string `json:"nip"`
	NUPTK                           *string `json:"nuptk"`
	Kewarganegaraan                 *string `json:"kewarganegaraan"`
	GolonganDarah                   *string `json:"golongan_darah"`
	BeratBadan                      *string `json:"berat_badan"`
	TinggiBadan                     *string `json:"tinggi_badan"`
	PenyakitYangSeringKambuh        *string `json:"penyakit_yang_sering_kambuh"`
	KelainanJasmani                 *string `json:"kelainan_jasmani"`
	PenyakitKronisYangPernahDiderita *string `json:"penyakit_kronis_yang_pernah_diderita"`
	RTRW                            *string `json:"rt_rw"`
	KodePos                         *string `json:"kode_pos"`
	PendidikanTerakhir              *string `json:"pendidikan_terakhir"`
	Jurusan                         *string `json:"jurusan"`
	TahunLulus                      *string `json:"tahun_lulus"`
	TahunMasuk                      *string `json:"tahun_masuk"`
}

// TeacherResponse represents the response body for a teacher.
type TeacherResponse struct {
	ID                              string     `json:"id"`
	UserID                          string     `json:"user_id"`
	SchoolID                        string     `json:"school_id"`
	Name                            string     `json:"name"`
	Email                           string     `json:"email"`
	Username                        string     `json:"username"`
	Avatar                          string     `json:"avatar"`
	NIP                             *string    `json:"nip"`
	NUPTK                           *string    `json:"nuptk"`
	PhoneNumber                     *string    `json:"phone_number"`
	Address                         *string    `json:"address"`
	Gender                          *string    `json:"gender"`
	Religion                        *string    `json:"religion"`
	BirthPlace                      *string    `json:"birth_place"`
	BirthDate                       *time.Time `json:"birth_date"`
	NIK                             *string    `json:"nik"`
	KTPImagePath                    *string    `json:"ktp_image_path"`
	Kewarganegaraan                 *string    `json:"kewarganegaraan"`
	GolonganDarah                   *string    `json:"golongan_darah"`
	BeratBadan                      *string    `json:"berat_badan"`
	TinggiBadan                     *string    `json:"tinggi_badan"`
	PenyakitYangSeringKambuh        *string    `json:"penyakit_yang_sering_kambuh"`
	KelainanJasmani                 *string    `json:"kelainan_jasmani"`
	PenyakitKronisYangPernahDiderita *string    `json:"penyakit_kronis_yang_pernah_diderita"`
	RTRW                            *string    `json:"rt_rw"`
	KodePos                         *string    `json:"kode_pos"`
	PendidikanTerakhir              *string    `json:"pendidikan_terakhir"`
	Jurusan                         *string    `json:"jurusan"`
	TahunLulus                      *string    `json:"tahun_lulus"`
	TahunMasuk                      *string    `json:"tahun_masuk"`
	CreatedAt                       time.Time  `json:"created_at"`
	UpdatedAt                       time.Time  `json:"updated_at"`
}
