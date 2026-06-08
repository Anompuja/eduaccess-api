package http

import (
	"bytes"
	"net/http"
	"strings"

	authdomain "github.com/eduaccess/eduaccess-api/internal/auth/domain"
	authmw "github.com/eduaccess/eduaccess-api/internal/shared/middleware"
	"github.com/eduaccess/eduaccess-api/internal/shared/response"
	"github.com/eduaccess/eduaccess-api/internal/student/application"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/xuri/excelize/v2"
)

// Column indices (0-based) in the template sheet.
const (
	colName              = 0
	colEmail             = 1
	colNIS               = 2
	colNISN              = 3
	colGender            = 4
	colReligion          = 5
	colBirthPlace        = 6
	colBirthDate         = 7
	colPhoneNumber       = 8
	colAddress           = 9
	colTahunMasuk        = 10
	colJalurMasuk        = 11
	colEducationLevel    = 12
	colClassName         = 13
	colSubClassName      = 14
)

var templateHeaders = []string{
	"name", "email", "nis", "nisn", "gender", "religion",
	"birth_place", "birth_date", "phone_number", "address",
	"tahun_masuk", "jalur_masuk",
	"jenjang_pendidikan", "kelas", "sub_kelas",
}

var templateExample = []string{
	"Budi Santoso", "budi@example.com", "12345", "9876543210",
	"L", "Islam", "Jakarta", "2007-05-20",
	"081234567890", "Jl. Contoh No. 1",
	"2022", "reguler",
	"SMA", "Kelas 10", "10A",
}

// DownloadBulkTemplate godoc
//
//	@Summary      Download student bulk import template
//	@Description  Returns an XLSX template with required columns for bulk student import. The last three columns (jenjang_pendidikan, kelas, sub_kelas) accept the names configured in your school's academic settings.
//	@Tags         students
//	@Produce      application/vnd.openxmlformats-officedocument.spreadsheetml.sheet
//	@Security     BearerAuth
//	@Success      200  {file}    binary
//	@Router       /students/bulk/template [get]
func (h *Handler) DownloadBulkTemplate(c echo.Context) error {
	f := excelize.NewFile()
	defer f.Close()

	const sheet = "Students"
	f.SetSheetName("Sheet1", sheet)

	// Header row (bold style)
	style, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{Bold: true},
		Fill: excelize.Fill{Type: "pattern", Color: []string{"#D9E1F2"}, Pattern: 1},
	})
	for i, h := range templateHeaders {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		f.SetCellValue(sheet, cell, h)
		f.SetCellStyle(sheet, cell, cell, style)
	}

	// Example data row
	for i, val := range templateExample {
		cell, _ := excelize.CoordinatesToCellName(i+1, 2)
		f.SetCellValue(sheet, cell, val)
	}

	// Set column widths for readability
	widths := map[int]float64{
		1: 20, 2: 28, 3: 12, 4: 14, 5: 8, 6: 12,
		7: 16, 8: 14, 9: 16, 10: 30,
		11: 12, 12: 14, 13: 20, 14: 16, 15: 14,
	}
	for col, w := range widths {
		name, _ := excelize.ColumnNumberToName(col)
		f.SetColWidth(sheet, name, name, w)
	}

	var buf bytes.Buffer
	if err := f.Write(&buf); err != nil {
		return response.InternalError(c, "failed to generate template")
	}

	c.Response().Header().Set("Content-Disposition", `attachment; filename="student_template.xlsx"`)
	return c.Blob(http.StatusOK,
		"application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
		buf.Bytes())
}

// BulkCreateStudents godoc
//
//	@Summary      Bulk import students from Excel
//	@Description  Accepts a filled XLSX template (field name: "file"). Creates one student per data row using name-based lookup for jenjang_pendidikan, kelas, and sub_kelas. Returns a per-row result summary; failed rows do not block successful ones.
//	@Tags         students
//	@Accept       multipart/form-data
//	@Produce      json
//	@Security     BearerAuth
//	@Param        file  formData  file  true  "Filled Excel template (.xlsx)"
//	@Success      200   {object}  response.Response{data=BulkCreateResultResponse}
//	@Failure      400   {object}  response.Response
//	@Failure      403   {object}  response.Response
//	@Router       /students/bulk [post]
func (h *Handler) BulkCreateStudents(c echo.Context) error {
	fileHeader, err := c.FormFile("file")
	if err != nil {
		return response.BadRequest(c, "file is required (multipart field: file)")
	}

	src, err := fileHeader.Open()
	if err != nil {
		return response.BadRequest(c, "cannot open uploaded file")
	}
	defer src.Close()

	f, err := excelize.OpenReader(src)
	if err != nil {
		return response.BadRequest(c, "invalid Excel file — must be .xlsx format")
	}
	defer f.Close()

	sheets := f.GetSheetList()
	if len(sheets) == 0 {
		return response.BadRequest(c, "Excel file has no sheets")
	}

	rows, err := f.GetRows(sheets[0])
	if err != nil {
		return response.BadRequest(c, "cannot read Excel rows")
	}
	// strip empty trailing rows
	for len(rows) > 0 {
		last := rows[len(rows)-1]
		empty := true
		for _, cell := range last {
			if strings.TrimSpace(cell) != "" {
				empty = false
				break
			}
		}
		if empty {
			rows = rows[:len(rows)-1]
		} else {
			break
		}
	}
	if len(rows) < 2 {
		return response.BadRequest(c, "no data rows found — file must contain a header row and at least one student row")
	}

	bulkRows := make([]application.BulkStudentRow, 0, len(rows)-1)
	for i, row := range rows[1:] { // skip header row
		bulkRows = append(bulkRows, application.BulkStudentRow{
			RowNumber:         i + 2,
			Name:              xlCell(row, colName),
			Email:             xlCell(row, colEmail),
			NIS:               xlCell(row, colNIS),
			NISN:              xlCell(row, colNISN),
			Gender:            xlCell(row, colGender),
			Religion:          xlCell(row, colReligion),
			BirthPlace:        xlCell(row, colBirthPlace),
			BirthDate:         xlCell(row, colBirthDate),
			PhoneNumber:       xlCell(row, colPhoneNumber),
			Address:           xlCell(row, colAddress),
			TahunMasuk:        xlCell(row, colTahunMasuk),
			JalurMasukSekolah: xlCell(row, colJalurMasuk),
			EducationLevel:    xlCell(row, colEducationLevel),
			ClassName:         xlCell(row, colClassName),
			SubClassName:      xlCell(row, colSubClassName),
		})
	}

	cmd := application.BulkCreateStudentCommand{
		RequesterSchoolID: authmw.GetSchoolID(c),
		RequesterRole:     authmw.GetRole(c),
		Rows:              bulkRows,
	}
	// Superadmin has no school in their JWT; require an explicit ?school_id param.
	if authmw.GetRole(c) == authdomain.RoleSuperadmin && authmw.GetSchoolID(c) == nil {
		raw := c.QueryParam("school_id")
		if raw == "" {
			return response.BadRequest(c, "school_id query param is required for superadmin")
		}
		sid, parseErr := uuid.Parse(raw)
		if parseErr != nil {
			return response.BadRequest(c, "invalid school_id")
		}
		cmd.SchoolID = &sid
	}

	result, err := h.bulkCreateStudent.Handle(c.Request().Context(), cmd)
	if err != nil {
		return handleAppError(c, err)
	}

	errs := make([]BulkRowErrorResponse, 0, len(result.Errors))
	for _, e := range result.Errors {
		errs = append(errs, BulkRowErrorResponse{Row: e.Row, Email: e.Email, Message: e.Message})
	}
	return response.OK(c, "bulk import complete", BulkCreateResultResponse{
		Total:   result.Total,
		Created: result.Created,
		Failed:  result.Failed,
		Errors:  errs,
	})
}

// xlCell safely reads a cell value from a row slice, trimming whitespace.
func xlCell(row []string, idx int) string {
	if idx >= len(row) {
		return ""
	}
	return strings.TrimSpace(row[idx])
}
