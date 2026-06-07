package complaint

import (
	"database/sql"
	"errors"
	"strings"
	"fmt"
	"pengaduan_be2/pkg/db"
	"pengaduan_be2/pkg/utils"
)

type ComplaintService struct{}

func NewComplaintService() *ComplaintService {
	return &ComplaintService{}
}

// GetCategories ambil semua kategori
func (s *ComplaintService) GetCategories() ([]Category, error) {
	rows, err := db.DB.Query(`
		SELECT id, name, COALESCE(description, ''), COALESCE(icon, ''), sort_order 
		FROM categories WHERE is_active = TRUE ORDER BY sort_order`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var categories []Category
	for rows.Next() {
		var cat Category
		var desc, icon string
		rows.Scan(&cat.ID, &cat.Name, &desc, &icon, &cat.SortOrder)
		if desc != "" {
			cat.Description = &desc
		}
		if icon != "" {
			cat.Icon = &icon
		}
		categories = append(categories, cat)
	}
	return categories, nil
}

// GetGovernorInvestigations - ambil investigasi untuk governor
func (s *ComplaintService) GetGovernorInvestigations(provinceApiID int, query *GetInvestigationsQuery) ([]InvestigationResponse, int, error) {
	offset := (query.Page - 1) * query.Limit

	sqlQuery := `
		SELECT 
			c.id as complaint_id,
			c.tracking_code,
			c.description,
			c.location_detail,
			c.status,
			c.created_at,
			COALESCE(u.username, '') as user_name,
			COALESCE(u.fullname, '') as user_fullname,
			COALESCE(inv.id, 0) as investigator_id,
			COALESCE(inv.username, '') as investigator_name,
			c.assigned_investigator_at as assigned_at
		FROM complaints c
		JOIN users u ON c.user_id = u.id
		LEFT JOIN users inv ON c.assigned_investigator_id = inv.id
		WHERE c.province_api_id = ?
	`

	countQuery := `SELECT COUNT(*) FROM complaints WHERE province_api_id = ?`
	args := []interface{}{provinceApiID}
	countArgs := []interface{}{provinceApiID}

	// Filter status
	if query.Status != "" && query.Status != "all" {
		sqlQuery += " AND c.status = ?"
		countQuery += " AND status = ?"
		args = append(args, query.Status)
		countArgs = append(countArgs, query.Status)
	}

	// Filter search
	if query.Search != "" {
		sqlQuery += " AND (c.tracking_code LIKE ? OR c.description LIKE ?)"
		countQuery += " AND (tracking_code LIKE ? OR description LIKE ?)"
		searchTerm := "%" + query.Search + "%"
		args = append(args, searchTerm, searchTerm)
		countArgs = append(countArgs, searchTerm, searchTerm)
	}

	sqlQuery += " ORDER BY c.created_at DESC LIMIT ? OFFSET ?"
	args = append(args, query.Limit, offset)

	// Get total
	var total int
	err := db.DB.QueryRow(countQuery, countArgs...).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	// Get data
	rows, err := db.DB.Query(sqlQuery, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var investigations []InvestigationResponse
	for rows.Next() {
		var inv InvestigationResponse
		var investigatorID int
		var investigatorName sql.NullString
		var assignedAt sql.NullTime

		err := rows.Scan(
			&inv.ComplaintID,
			&inv.TrackingCode,
			&inv.Description,
			&inv.LocationDetail,
			&inv.Status,
			&inv.CreatedAt,
			&inv.UserName,
			&inv.UserFullname,
			&investigatorID,
			&investigatorName,
			&assignedAt,
		)
		if err != nil {
			continue
		}

		inv.ID = inv.ComplaintID
		if investigatorID > 0 {
			inv.InvestigatorID = &investigatorID
			if investigatorName.Valid {
				inv.InvestigatorName = &investigatorName.String
			}
		}
		if assignedAt.Valid {
			inv.AssignedAt = &assignedAt.Time
		}

		inv.StatusText = getStatusText(inv.Status)
		investigations = append(investigations, inv)
	}

	return investigations, total, nil
}

// GetGovernorComplaints - KHUSUS UNTUK GOVERNOR (TAMBAHKAN METHOD INI)
// GetGovernorComplaints - KHUSUS UNTUK GOVERNOR
func (s *ComplaintService) GetGovernorComplaints(provinceApiID int, query *GovernorComplaintsQuery) ([]Complaint, int, error) {
	offset := (query.Page - 1) * query.Limit

	// 🔥 SELECT dengan jumlah kolom YANG SAMA dengan SCAN
	sqlQuery := `
		SELECT 
			c.id, 
			c.tracking_code, 
			c.user_id,
			c.province_api_id,
			c.location_detail,
			c.category_id,
			c.description,
			c.photo,
			c.status,
			c.rejected_reason,
			c.assigned_investigator_id,
			c.investigation_result,
			c.investigation_evidence,
			c.created_at,
			c.updated_at
		FROM complaints c
		WHERE c.province_api_id = ?
	`

	countQuery := "SELECT COUNT(*) FROM complaints WHERE province_api_id = ?"
	args := []interface{}{provinceApiID}
	countArgs := []interface{}{provinceApiID}

	if query.Status != "" && query.Status != "all" {
		sqlQuery += " AND c.status = ?"
		countQuery += " AND status = ?"
		args = append(args, query.Status)
		countArgs = append(countArgs, query.Status)
	}

	if query.Search != "" {
		sqlQuery += " AND (c.tracking_code LIKE ? OR c.description LIKE ?)"
		countQuery += " AND (tracking_code LIKE ? OR description LIKE ?)"
		searchTerm := "%" + query.Search + "%"
		args = append(args, searchTerm, searchTerm)
		countArgs = append(countArgs, searchTerm, searchTerm)
	}

	sqlQuery += " ORDER BY c.created_at DESC LIMIT ? OFFSET ?"
	args = append(args, query.Limit, offset)

	// Get total
	var total int
	err := db.DB.QueryRow(countQuery, countArgs...).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	if total == 0 {
		return []Complaint{}, 0, nil
	}

	// Get data
	rows, err := db.DB.Query(sqlQuery, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var complaints []Complaint
	for rows.Next() {
		var c Complaint
		var photo, rejectedReason, investigationResult, investigationEvidence sql.NullString
		var assignedInvID sql.NullInt64

		err := rows.Scan(
			&c.ID,
			&c.TrackingCode,
			&c.UserID,
			&c.ProvinceID,
			&c.LocationDetail,
			&c.CategoryID,
			&c.Description,
			&photo,
			&c.Status,
			&rejectedReason,
			&assignedInvID,
			&investigationResult,
			&investigationEvidence,
			&c.CreatedAt,
			&c.UpdatedAt,
		)

		if err != nil {
			continue
		}

		// Isi data user
		c.UserName = ""
		c.UserFullname = ""
		c.ProvinceName = ""

		if photo.Valid {
			c.Photo = &photo.String
		}
		if rejectedReason.Valid {
			c.RejectedReason = &rejectedReason.String
		}
		if assignedInvID.Valid {
			id := int(assignedInvID.Int64)
			c.AssignedInvestigatorID = &id
		}
		if investigationResult.Valid {
			c.InvestigationResult = &investigationResult.String
		}
		if investigationEvidence.Valid {
			c.InvestigationEvidence = &investigationEvidence.String
		}

		c.StatusText = getStatusText(c.Status)
		complaints = append(complaints, c)
	}

	// Ambil nama user dan nama provinsi setelahnya (opsional)
	for i := range complaints {
		var userName, userFullname, provinceName string
		db.DB.QueryRow("SELECT COALESCE(username, ''), COALESCE(fullname, '') FROM users WHERE id = ?", complaints[i].UserID).Scan(&userName, &userFullname)
		db.DB.QueryRow("SELECT COALESCE(name, '') FROM provinces WHERE api_id = ?", complaints[i].ProvinceID).Scan(&provinceName)
		complaints[i].UserName = userName
		complaints[i].UserFullname = userFullname
		complaints[i].ProvinceName = provinceName
	}

	return complaints, total, nil
}

// GetGovernorReports - ambil laporan untuk governor
func (s *ComplaintService) GetGovernorReports(provinceApiID int, query *ReportQuery) ([]Complaint, *ReportStatsResponse, int, error) {
	offset := (query.Page - 1) * query.Limit

	// Build date filter
	dateFilter := ""
	switch query.Period {
	case "today":
		dateFilter = "DATE(c.created_at) = CURDATE()"
	case "week":
		dateFilter = "YEARWEEK(c.created_at) = YEARWEEK(CURDATE())"
	case "month":
		dateFilter = "MONTH(c.created_at) = MONTH(CURDATE()) AND YEAR(c.created_at) = YEAR(CURDATE())"
	case "year":
		dateFilter = "YEAR(c.created_at) = YEAR(CURDATE())"
	case "custom":
		if query.StartDate != "" && query.EndDate != "" {
			dateFilter = fmt.Sprintf("DATE(c.created_at) BETWEEN '%s' AND '%s'", query.StartDate, query.EndDate)
		}
	}

	// Base query
	baseQuery := `
		FROM complaints c
		JOIN users u ON c.user_id = u.id
		JOIN categories cat ON c.category_id = cat.id
		WHERE c.province_api_id = ?
	`
	if dateFilter != "" {
		baseQuery += " AND " + dateFilter
	}

	selectQuery := `
		SELECT c.id, c.tracking_code, c.user_id, u.username, u.fullname,
			c.province_api_id, c.location_detail, c.category_id, cat.name,
			c.description, c.photo, c.status, c.rejected_reason,
			c.assigned_investigator_id, c.created_at, c.updated_at
	`

	args := []interface{}{provinceApiID}
	countArgs := []interface{}{provinceApiID}

	if query.Status != "" && query.Status != "all" {
		baseQuery += " AND c.status = ?"
		args = append(args, query.Status)
		countArgs = append(countArgs, query.Status)
	}

	if query.Search != "" {
		baseQuery += " AND (c.tracking_code LIKE ? OR c.description LIKE ?)"
		searchTerm := "%" + query.Search + "%"
		args = append(args, searchTerm, searchTerm)
		countArgs = append(countArgs, searchTerm, searchTerm)
	}

	// Count total
	var total int
	err := db.DB.QueryRow("SELECT COUNT(*) "+baseQuery, countArgs...).Scan(&total)
	if err != nil {
		return nil, nil, 0, err
	}

	// Get data
	dataQuery := selectQuery + baseQuery + " ORDER BY c.created_at DESC LIMIT ? OFFSET ?"
	dataArgs := append(args, query.Limit, offset)
	rows, err := db.DB.Query(dataQuery, dataArgs...)
	if err != nil {
		return nil, nil, 0, err
	}
	defer rows.Close()

	var complaints []Complaint
	for rows.Next() {
		c, err := s.scanComplaint(rows)
		if err != nil {
			continue
		}
		complaints = append(complaints, c)
	}

	// Get stats
	stats, err := s.getReportStats(provinceApiID, dateFilter, query.Status, query.Search)
	if err != nil {
		return complaints, nil, total, err
	}

	return complaints, stats, total, nil
}

// getReportStats - ambil statistik laporan
func (s *ComplaintService) getReportStats(provinceApiID int, dateFilter, statusFilter, search string) (*ReportStatsResponse, error) {
	var stats ReportStatsResponse

	whereClause := "WHERE c.province_api_id = ?"
	args := []interface{}{provinceApiID}

	if dateFilter != "" {
		whereClause += " AND " + dateFilter
	}
	if statusFilter != "" && statusFilter != "all" {
		whereClause += " AND c.status = ?"
		args = append(args, statusFilter)
	}
	if search != "" {
		whereClause += " AND (c.tracking_code LIKE ? OR c.description LIKE ?)"
		searchTerm := "%" + search + "%"
		args = append(args, searchTerm, searchTerm)
	}

	// Main stats
	err := db.DB.QueryRow(`
		SELECT 
			COUNT(*) as total,
			SUM(CASE WHEN c.status = 'pending_governor' THEN 1 ELSE 0 END) as pending,
			SUM(CASE WHEN c.status = 'investigation_assigned' THEN 1 ELSE 0 END) as investigation,
			SUM(CASE WHEN c.status = 'completed' THEN 1 ELSE 0 END) as completed,
			SUM(CASE WHEN c.status = 'rejected' THEN 1 ELSE 0 END) as rejected,
			SUM(CASE WHEN MONTH(c.created_at) = MONTH(CURDATE()) AND YEAR(c.created_at) = YEAR(CURDATE()) THEN 1 ELSE 0 END) as thisMonth,
			SUM(CASE WHEN YEARWEEK(c.created_at) = YEARWEEK(CURDATE()) THEN 1 ELSE 0 END) as thisWeek
		FROM complaints c
		`+whereClause, args...).Scan(
		&stats.Total, &stats.Pending, &stats.Investigation,
		&stats.Completed, &stats.Rejected, &stats.ThisMonth, &stats.ThisWeek,
	)
	if err != nil {
		return nil, err
	}

	// Avg completion days
	db.DB.QueryRow(`
		SELECT COALESCE(AVG(DATEDIFF(c.updated_at, c.created_at)), 0)
		FROM complaints c
		`+whereClause+` AND c.status = 'completed'`, args...).Scan(&stats.AvgCompletionDays)

	// By category
	rows, _ := db.DB.Query(`
		SELECT cat.name, COUNT(*) as count
		FROM complaints c
		JOIN categories cat ON c.category_id = cat.id
		`+whereClause+`
		GROUP BY cat.id, cat.name
		ORDER BY count DESC
		LIMIT 5
	`, args...)
	if rows != nil {
		defer rows.Close()
		for rows.Next() {
			var item CategoryCount
			rows.Scan(&item.Name, &item.Count)
			stats.ByCategory = append(stats.ByCategory, item)
		}
	}

	// By status
	rows2, _ := db.DB.Query(`
		SELECT 
			CASE c.status
				WHEN 'pending_governor' THEN 'Menunggu'
				WHEN 'investigation_assigned' THEN 'Investigasi'
				WHEN 'completed' THEN 'Selesai'
				WHEN 'rejected' THEN 'Ditolak'
				ELSE c.status
			END as name,
			COUNT(*) as count
		FROM complaints c
		`+whereClause+`
		GROUP BY c.status
	`, args...)
	if rows2 != nil {
		defer rows2.Close()
		for rows2.Next() {
			var item StatusCount
			rows2.Scan(&item.Name, &item.Value)
			stats.ByStatus = append(stats.ByStatus, item)
		}
	}

	return &stats, nil
}

// GetInvestigatorComplaints - ambil complaints untuk investigator
func (s *ComplaintService) GetInvestigatorComplaints(investigatorID int, query *InvestigatorComplaintsQuery) ([]Complaint, int, error) {
	offset := (query.Page - 1) * query.Limit

	// Ambil province_api_id dari investigator
	var provinceApiID int
	err := db.DB.QueryRow(`
		SELECT COALESCE(province_api_id, 0) FROM users 
		WHERE id = ? AND role = 'investigator' AND is_active = TRUE`,
		investigatorID,
	).Scan(&provinceApiID)
	if err != nil {
		return nil, 0, errors.New("gagal mengambil data provinsi")
	}

	if provinceApiID == 0 {
		return []Complaint{}, 0, nil
	}

	// Query untuk mengambil complaints yang ditugaskan ke investigator ini
	sqlQuery := `
		SELECT 
			c.id, 
			c.tracking_code, 
			c.user_id,
			c.province_api_id,
			c.location_detail,
			c.category_id,
			c.description,
			c.photo,
			c.status,
			c.rejected_reason,
			c.assigned_investigator_id,
			c.investigation_result,
			c.investigation_evidence,
			c.created_at,
			c.updated_at
		FROM complaints c
		WHERE c.province_api_id = ? 
		AND c.assigned_investigator_id = ?
	`

	countQuery := `SELECT COUNT(*) FROM complaints 
		WHERE province_api_id = ? AND assigned_investigator_id = ?`
	
	args := []interface{}{provinceApiID, investigatorID}
	countArgs := []interface{}{provinceApiID, investigatorID}

	if query.Status != "" && query.Status != "all" {
		sqlQuery += " AND c.status = ?"
		countQuery += " AND status = ?"
		args = append(args, query.Status)
		countArgs = append(countArgs, query.Status)
	}

	if query.Search != "" {
		sqlQuery += " AND (c.tracking_code LIKE ? OR c.description LIKE ?)"
		countQuery += " AND (tracking_code LIKE ? OR description LIKE ?)"
		searchTerm := "%" + query.Search + "%"
		args = append(args, searchTerm, searchTerm)
		countArgs = append(countArgs, searchTerm, searchTerm)
	}

	sqlQuery += " ORDER BY c.created_at DESC LIMIT ? OFFSET ?"
	args = append(args, query.Limit, offset)

	// Get total
	var total int
	err = db.DB.QueryRow(countQuery, countArgs...).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	if total == 0 {
		return []Complaint{}, 0, nil
	}

	// Get data
	rows, err := db.DB.Query(sqlQuery, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var complaints []Complaint
	for rows.Next() {
		var c Complaint
		var photo, rejectedReason, investigationResult, investigationEvidence sql.NullString
		var assignedInvID sql.NullInt64

		err := rows.Scan(
			&c.ID,
			&c.TrackingCode,
			&c.UserID,
			&c.ProvinceID,
			&c.LocationDetail,
			&c.CategoryID,
			&c.Description,
			&photo,
			&c.Status,
			&rejectedReason,
			&assignedInvID,
			&investigationResult,
			&investigationEvidence,
			&c.CreatedAt,
			&c.UpdatedAt,
		)

		if err != nil {
			continue
		}

		if photo.Valid {
			c.Photo = &photo.String
		}
		if rejectedReason.Valid {
			c.RejectedReason = &rejectedReason.String
		}
		if investigationResult.Valid {
			c.InvestigationResult = &investigationResult.String
		}
		if investigationEvidence.Valid {
			c.InvestigationEvidence = &investigationEvidence.String
		}

		c.StatusText = getStatusText(c.Status)
		complaints = append(complaints, c)
	}

	// Ambil nama user untuk setiap complaint
	for i := range complaints {
		var userName, userFullname string
		db.DB.QueryRow("SELECT COALESCE(username, ''), COALESCE(fullname, '') FROM users WHERE id = ?", complaints[i].UserID).Scan(&userName, &userFullname)
		complaints[i].UserName = userName
		complaints[i].UserFullname = userFullname
	}

	return complaints, total, nil
}

// GetInvestigatorStats - statistik untuk dashboard investigator
func (s *ComplaintService) GetInvestigatorStats(investigatorID int) (*InvestigatorStatsResponse, error) {
	// Ambil province_api_id dari investigator
	var provinceApiID int
	err := db.DB.QueryRow(`
		SELECT COALESCE(province_api_id, 0) FROM users 
		WHERE id = ? AND role = 'investigator' AND is_active = TRUE`,
		investigatorID,
	).Scan(&provinceApiID)
	if err != nil {
		return nil, err
	}

	var stats InvestigatorStatsResponse
	query := `
		SELECT 
			COUNT(*) as total,
			SUM(CASE WHEN status = 'investigation_assigned' THEN 1 ELSE 0 END) as investigation_assigned,
			SUM(CASE WHEN status = 'investigation_done' THEN 1 ELSE 0 END) as investigation_done,
			SUM(CASE WHEN status = 'completed' THEN 1 ELSE 0 END) as completed
		FROM complaints
		WHERE province_api_id = ? AND assigned_investigator_id = ?
	`

	err = db.DB.QueryRow(query, provinceApiID, investigatorID).Scan(
		&stats.Total,
		&stats.InvestigationAssigned,
		&stats.InvestigationDone,
		&stats.Completed,
	)

	return &stats, err
}

// SubmitInvestigationResultWithNote - investigator kirim hasil dengan catatan
func (s *ComplaintService) SubmitInvestigationResultWithNote(complaintID int, result, evidence string, isValid bool, notes string, investigatorID int) error {
	status := "investigation_done"
	rejectedReason := ""

	if !isValid {
		status = "rejected"
		rejectedReason = "Hasil investigasi: laporan tidak terbukti"
	}

	// Update investigation_result dengan format JSON atau text
	finalResult := result
	if notes != "" {
		finalResult = result + "\n\nCatatan: " + notes
	}

	_, err := db.DB.Exec(`
		UPDATE complaints 
		SET investigation_result = ?, 
			investigation_evidence = ?, 
			status = ?, 
			rejected_reason = COALESCE(NULLIF(?, ''), rejected_reason),
			investigation_completed_at = NOW(),
			updated_at = NOW()
		WHERE id = ? AND assigned_investigator_id = ?`,
		finalResult, evidence, status, rejectedReason, complaintID, investigatorID,
	)

	return err
}

// SubmitComplaint warga submit pengaduan baru
func (s *ComplaintService) SubmitComplaint(userID int, req *SubmitComplaintRequest) (*SubmitComplaintResponse, error) {
	trackingCode := utils.GenerateTrackingCode()

	result, err := db.DB.Exec(`
		INSERT INTO complaints (
			tracking_code, user_id, province_api_id, regency_id, district_id, village_id,
			location_detail, category_id, description, photo, status
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, 'pending_governor')`,
		trackingCode, userID, req.ProvinceID, req.RegencyID, req.DistrictID,
		req.VillageID, req.LocationDetail, req.CategoryID, req.Description, req.Photo,
	)

	if err != nil {
		return nil, errors.New("gagal menyimpan pengaduan: " + err.Error())
	}

	id, _ := result.LastInsertId()

	// Insert notifikasi ke gubernur provinsi terkait
	go s.notifyGovernor(req.ProvinceID, int(id), trackingCode)

	return &SubmitComplaintResponse{
		TrackingCode: trackingCode,
		ID:           int(id),
	}, nil
}

func (s *ComplaintService) notifyGovernor(provinceID, complaintID int, trackingCode string) {
	var governorID int
	err := db.DB.QueryRow(`
		SELECT id FROM users WHERE province_api_id = ? AND role = 'governor' AND is_active = TRUE`,
		provinceID,
	).Scan(&governorID)

	if err == nil {
		db.DB.Exec(`
			INSERT INTO notifications (user_id, complaint_id, title, message, type)
			VALUES (?, ?, 'Pengaduan Baru', 
					CONCAT('Pengaduan baru dengan kode ', ?, ' menunggu tindakan Anda'),
					'info')`,
			governorID, complaintID, trackingCode,
		)
	}
}

// GetUserComplaints ambil pengaduan milik user tertentu
func (s *ComplaintService) GetUserComplaints(userID int, page, limit int) ([]Complaint, int, error) {
    offset := (page - 1) * limit

    rows, err := db.DB.Query(`
        SELECT c.id, c.tracking_code, c.user_id, 
            COALESCE(u.username, '') as user_name,
            COALESCE(u.fullname, '') as user_fullname,
            c.province_api_id, COALESCE(p.name, '') as province_name,
            c.regency_id, r.name as regency_name,
            c.district_id, d.name as district_name,
            c.village_id, v.name as village_name,
            c.location_detail, c.category_id, COALESCE(cat.name, '') as category_name,
            c.description, c.photo,
            c.status, c.rejected_reason, c.assigned_investigator_id, 
            COALESCE(inv.username, '') as investigator_name,
            c.investigation_result, c.investigation_evidence,
            c.created_at, c.updated_at
        FROM complaints c
        JOIN users u ON c.user_id = u.id
        JOIN provinces p ON c.province_api_id = p.api_id
        LEFT JOIN regencies r ON c.regency_id = r.api_id
        LEFT JOIN districts d ON c.district_id = d.api_id
        LEFT JOIN villages v ON c.village_id = v.api_id
        JOIN categories cat ON c.category_id = cat.id
        LEFT JOIN users inv ON c.assigned_investigator_id = inv.id
        WHERE c.user_id = ?
        ORDER BY c.created_at DESC LIMIT ? OFFSET ?`,
        userID, limit, offset,
    )
    if err != nil {
        return nil, 0, err
    }
    defer rows.Close()

    var complaints []Complaint
    for rows.Next() {
        var c Complaint
        var regencyID, districtID, villageID sql.NullInt64
        var regencyName, districtName, villageName, photo, rejectedReason sql.NullString
        var assignedInvID sql.NullInt64
        var assignedInvName, investigationResult, investigationEvidence sql.NullString

        err := rows.Scan(
            &c.ID, &c.TrackingCode, &c.UserID,
            &c.UserName, &c.UserFullname,
            &c.ProvinceID, &c.ProvinceName,
            &regencyID, &regencyName,
            &districtID, &districtName,
            &villageID, &villageName,
            &c.LocationDetail,
            &c.CategoryID, &c.CategoryName,
            &c.Description, &photo,
            &c.Status, &rejectedReason,
            &assignedInvID, &assignedInvName,
            &investigationResult, &investigationEvidence,
            &c.CreatedAt, &c.UpdatedAt,
        )

        if err != nil {
            fmt.Printf("Scan error in GetUserComplaints: %v\n", err)
            continue
        }

        if photo.Valid {
            c.Photo = &photo.String
        }
        if rejectedReason.Valid {
            c.RejectedReason = &rejectedReason.String
        }
        if assignedInvID.Valid {
            id := int(assignedInvID.Int64)
            c.AssignedInvestigatorID = &id
            if assignedInvName.Valid {
                c.AssignedInvestigatorName = &assignedInvName.String
            }
        }
        if investigationResult.Valid {
            c.InvestigationResult = &investigationResult.String
        }
        if investigationEvidence.Valid {
            c.InvestigationEvidence = &investigationEvidence.String
        }

        c.StatusText = getStatusText(c.Status)
        complaints = append(complaints, c)
    }

    var total int
    db.DB.QueryRow("SELECT COUNT(*) FROM complaints WHERE user_id = ?", userID).Scan(&total)

    return complaints, total, nil
}

// GetAllComplaints untuk admin/governor/investigator
func (s *ComplaintService) GetAllComplaints(role string, userID, provinceID int, query *GetComplaintsQuery) ([]Complaint, int, error) {
	offset := (query.Page - 1) * query.Limit

	baseQuery := `
		FROM complaints c
		JOIN users u ON c.user_id = u.id
		JOIN provinces p ON c.province_api_id = p.id
		LEFT JOIN regencies r ON c.regency_id = r.id
		LEFT JOIN districts d ON c.district_id = d.id
		LEFT JOIN villages v ON c.village_id = v.id
		JOIN categories cat ON c.category_id = cat.id
		LEFT JOIN users inv ON c.assigned_investigator_id = inv.id
		WHERE 1=1
	`

	selectQuery := `
		SELECT c.id, c.tracking_code, c.user_id, u.username, u.fullname,
			c.province_api_id, p.name,
			c.regency_id, r.name, c.district_id, d.name, c.village_id, v.name,
			c.location_detail, c.category_id, cat.name, c.description, c.photo,
			c.status, c.rejected_reason, c.assigned_investigator_id, 
			COALESCE(inv.username, ''), c.investigation_result, c.investigation_evidence,
			c.created_at, c.updated_at
	`

	args := []interface{}{}
	conditions := []string{}

	// Filter berdasarkan role
	if role == "governor" {
		conditions = append(conditions, "c.province_api_id = ?")
		args = append(args, provinceID)
	} else if role == "investigator" {
		conditions = append(conditions, "c.province_api_id = ? AND c.status IN ('investigation_assigned', 'investigation_done')")
		args = append(args, provinceID)
	}

	// Filter status
	if query.Status != "" && query.Status != "all" {
		conditions = append(conditions, "c.status = ?")
		args = append(args, query.Status)
	}

	// Filter province (untuk admin)
	if query.ProvinceID > 0 && role == "admin" {
		conditions = append(conditions, "c.province_api_id = ?")
		args = append(args, query.ProvinceID)
	}

	// Filter search
	if query.Search != "" {
		conditions = append(conditions, "(c.tracking_code LIKE ? OR c.description LIKE ?)")
		searchTerm := "%" + query.Search + "%"
		args = append(args, searchTerm, searchTerm)
	}

	// Build WHERE clause
	if len(conditions) > 0 {
		baseQuery += " AND " + strings.Join(conditions, " AND ")
	}

	// Count query
	countQuery := "SELECT COUNT(*) " + baseQuery
	var total int
	err := db.DB.QueryRow(countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	// Data query
	dataQuery := selectQuery + baseQuery + " ORDER BY c.created_at DESC LIMIT ? OFFSET ?"
	dataArgs := append(args, query.Limit, offset)

	rows, err := db.DB.Query(dataQuery, dataArgs...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var complaints []Complaint
	for rows.Next() {
		c, err := s.scanComplaint(rows)
		if err != nil {
			continue
		}
		complaints = append(complaints, c)
	}

	return complaints, total, nil
}

// GetComplaintByID - versi simpel sementara
func (s *ComplaintService) GetComplaintByID(id int) (*Complaint, error) {
	var c Complaint
	var photo sql.NullString
	
	query := `
		SELECT id, tracking_code, user_id, province_api_id, 
		       location_detail, category_id, description, photo, 
		       status, rejected_reason, assigned_investigator_id,
		       investigation_result, investigation_evidence,
		       created_at, updated_at
		FROM complaints 
		WHERE id = ?
	`
	
	err := db.DB.QueryRow(query, id).Scan(
		&c.ID, &c.TrackingCode, &c.UserID, &c.ProvinceID,
		&c.LocationDetail, &c.CategoryID, &c.Description, &photo,
		&c.Status, &c.RejectedReason, &c.AssignedInvestigatorID,
		&c.InvestigationResult, &c.InvestigationEvidence,
		&c.CreatedAt, &c.UpdatedAt,
	)
	
	if err != nil {
		return nil, err
	}
	
	if photo.Valid {
		c.Photo = &photo.String
	}
	
	c.StatusText = getStatusText(c.Status)
	
	// Ambil nama user
	db.DB.QueryRow("SELECT COALESCE(username, '') FROM users WHERE id = ?", c.UserID).Scan(&c.UserName)
	
	return &c, nil
}
// GetComplaintByTrackingCode cek status publik
func (s *ComplaintService) GetComplaintByTrackingCode(trackingCode string) (*Complaint, error) {
	row := db.DB.QueryRow(`
		SELECT c.id, c.tracking_code, c.user_id, u.username, u.fullname,
			c.province_api_id, p.name,
			c.regency_id, r.name, c.district_id, d.name, c.village_id, v.name,
			c.location_detail, c.category_id, cat.name, c.description, c.photo,
			c.status, c.rejected_reason, c.assigned_investigator_id, 
			COALESCE(inv.username, ''), c.investigation_result, c.investigation_evidence,
			c.created_at, c.updated_at
		FROM complaints c
		JOIN users u ON c.user_id = u.id
		JOIN provinces p ON c.province_api_id = p.id
		LEFT JOIN regencies r ON c.regency_id = r.id
		LEFT JOIN districts d ON c.district_id = d.id
		LEFT JOIN villages v ON c.village_id = v.id
		JOIN categories cat ON c.category_id = cat.id
		LEFT JOIN users inv ON c.assigned_investigator_id = inv.id
		WHERE c.tracking_code = ?`, trackingCode)

	complaint, err := s.scanComplaintRow(row)
	if err != nil {
		return nil, errors.New("kode tracking tidak ditemukan")
	}

	return complaint, nil
}

// GetGovernorStats - statistik untuk dashboard governor
func (s *ComplaintService) GetGovernorStats(provinceApiID int) (*GovernorStatsResponse, error) {
	var stats GovernorStatsResponse

	query := `
		SELECT 
			COUNT(*) as total,
			SUM(CASE WHEN status = 'pending_governor' THEN 1 ELSE 0 END) as pending_governor,
			SUM(CASE WHEN status = 'investigation_assigned' THEN 1 ELSE 0 END) as investigation_assigned,
			SUM(CASE WHEN status = 'investigation_done' THEN 1 ELSE 0 END) as investigation_done,
			SUM(CASE WHEN status = 'governor_processing' THEN 1 ELSE 0 END) as governor_processing,
			SUM(CASE WHEN status = 'completed' THEN 1 ELSE 0 END) as completed,
			SUM(CASE WHEN status = 'rejected' THEN 1 ELSE 0 END) as rejected
		FROM complaints
		WHERE province_api_id = ?
	`

	err := db.DB.QueryRow(query, provinceApiID).Scan(
		&stats.Total,
		&stats.PendingGovernor,
		&stats.InvestigationAssigned,
		&stats.InvestigationDone,
		&stats.GovernorProcessing,
		&stats.Completed,
		&stats.Rejected,
	)

	return &stats, err
}

// GetInvestigatorsByProvince - ambil daftar investigator di provinsi
func (s *ComplaintService) GetInvestigatorsByProvince(provinceApiID int) ([]Investigator, error) {
	rows, err := db.DB.Query(`
		SELECT id, username, fullname, email, province_api_id
		FROM users
		WHERE role = 'investigator' AND province_api_id = ? AND is_active = TRUE
		ORDER BY fullname ASC
	`, provinceApiID)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var investigators []Investigator
	for rows.Next() {
		var inv Investigator
		err := rows.Scan(&inv.ID, &inv.Username, &inv.Fullname, &inv.Email, &inv.ProvinceApiID)
		if err != nil {
			continue
		}
		investigators = append(investigators, inv)
	}

	return investigators, nil
}

// UpdateStatus - update status pengaduan (terima/tolak)
func (s *ComplaintService) UpdateStatus(complaintID int, status string, rejectReason *string, userID int, role string) error {
	// Cek status lama
	var oldStatus string
	var provinceID int
	err := db.DB.QueryRow("SELECT status, province_api_id FROM complaints WHERE id = ?", complaintID).
		Scan(&oldStatus, &provinceID)
	if err != nil {
		return errors.New("pengaduan tidak ditemukan")
	}

	// Validasi role governor
	if role == "governor" {
		if oldStatus != "pending_governor" {
			return errors.New("pengaduan sudah diproses, tidak dapat diubah lagi")
		}
		if status != "investigation_assigned" && status != "rejected" {
			return errors.New("status tidak valid untuk gubernur")
		}
	}

	// Update query
	var query string
	var args []interface{}

	if status == "rejected" {
		reason := ""
		if rejectReason != nil {
			reason = *rejectReason
		}
		query = `UPDATE complaints SET status = ?, rejected_reason = ?, updated_at = NOW() WHERE id = ?`
		args = []interface{}{status, reason, complaintID}
	} else if status == "investigation_assigned" {
		// 🔥 CARI INVESTIGATOR DI PROVINSI YANG SAMA
		var investigatorID int
		err := db.DB.QueryRow(`
			SELECT id FROM users 
			WHERE role = 'investigator' 
			AND province_api_id = ? 
			AND is_active = TRUE 
			LIMIT 1
		`, provinceID).Scan(&investigatorID)
		
		if err != nil {
			// Tidak ada investigator aktif, tetap update status tanpa assign
			query = `UPDATE complaints SET status = ?, governor_processed_at = NOW(), updated_at = NOW() WHERE id = ?`
			args = []interface{}{status, complaintID}
		} else {
			// Ada investigator, assign otomatis
			query = `UPDATE complaints 
				SET status = ?, 
					governor_processed_at = NOW(), 
					assigned_investigator_id = ?,
					assigned_investigator_at = NOW(),
					updated_at = NOW() 
				WHERE id = ?`
			args = []interface{}{status, investigatorID, complaintID}
			
			// Notifikasi ke investigator
			go func() {
				db.DB.Exec(`
					INSERT INTO notifications (user_id, complaint_id, title, message, type)
					VALUES (?, ?, 'Tugas Investigasi', 'Anda ditugaskan untuk menginvestigasi pengaduan ini', 'info')`,
					investigatorID, complaintID,
				)
			}()
		}
	} else {
		query = `UPDATE complaints SET status = ?, updated_at = NOW() WHERE id = ?`
		args = []interface{}{status, complaintID}
	}

	_, err = db.DB.Exec(query, args...)
	if err != nil {
		return err
	}

	// Insert log
	db.DB.Exec(`
		INSERT INTO activity_logs (user_id, action, complaint_id, old_status, new_status)
		VALUES (?, 'UPDATE_STATUS', ?, ?, ?)`,
		userID, complaintID, oldStatus, status,
	)

	return nil
}
// AssignInvestigator - governor tugaskan investigator
func (s *ComplaintService) AssignInvestigator(complaintID, investigatorID int, governorID int) error {
	// Cek investigator valid dan satu provinsi
	var investigatorProvinceID int
	err := db.DB.QueryRow(`
		SELECT province_api_id FROM users 
		WHERE id = ? AND role = 'investigator' AND is_active = TRUE`,
		investigatorID,
	).Scan(&investigatorProvinceID)
	if err != nil {
		return errors.New("investigator tidak valid")
	}

	var complaintProvinceID int
	var complaintStatus string
	err = db.DB.QueryRow("SELECT province_api_id, status FROM complaints WHERE id = ?", complaintID).
		Scan(&complaintProvinceID, &complaintStatus)
	if err != nil {
		return errors.New("pengaduan tidak ditemukan")
	}

	if complaintStatus != "pending_governor" {
		return errors.New("pengaduan sudah diproses")
	}

	if investigatorProvinceID != complaintProvinceID {
		return errors.New("investigator harus dari provinsi yang sama")
	}

	_, err = db.DB.Exec(`
		UPDATE complaints 
		SET assigned_investigator_id = ?, 
			status = 'investigation_assigned', 
			assigned_investigator_at = NOW(),
			updated_at = NOW()
		WHERE id = ?`,
		investigatorID, complaintID,
	)

	if err != nil {
		return err
	}

	// Notifikasi ke investigator
	db.DB.Exec(`
		INSERT INTO notifications (user_id, complaint_id, title, message, type)
		VALUES (?, ?, 'Tugas Investigasi', 'Anda ditugaskan untuk menginvestigasi pengaduan ini', 'info')`,
		investigatorID, complaintID,
	)

	return nil
}

// SubmitInvestigationResult - investigator kirim hasil
func (s *ComplaintService) SubmitInvestigationResult(complaintID int, result, evidence string, isValid bool, investigatorID int) error {
	status := "investigation_done"
	rejectedReason := ""

	if !isValid {
		status = "rejected"
		rejectedReason = "Hasil investigasi: laporan tidak terbukti"
	}

	_, err := db.DB.Exec(`
		UPDATE complaints 
		SET investigation_result = ?, investigation_evidence = ?, 
			status = ?, rejected_reason = COALESCE(NULLIF(?, ''), rejected_reason),
			investigation_completed_at = NOW(),
			updated_at = NOW()
		WHERE id = ? AND assigned_investigator_id = ?`,
		result, evidence, status, rejectedReason, complaintID, investigatorID,
	)

	return err
}

// SubmitProcessReport - governor kirim laporan proses ke admin
func (s *ComplaintService) SubmitProcessReport(complaintID int, governorID int, req *ProcessReportRequest) error {
	_, err := db.DB.Exec(`
		INSERT INTO process_reports (complaint_id, governor_id, process_photos, process_notes, process_date, status)
		VALUES (?, ?, ?, ?, ?, 'pending')`,
		complaintID, governorID, req.ProcessPhotos, req.ProcessNotes, req.ProcessDate,
	)

	if err != nil {
		return err
	}

	// Update status complaint
	_, err = db.DB.Exec(`
		UPDATE complaints SET status = 'process_report_submitted', updated_at = NOW() WHERE id = ?`,
		complaintID,
	)

	return err
}

// SubmitCompletionReport - governor kirim laporan akhir ke admin
func (s *ComplaintService) SubmitCompletionReport(complaintID int, governorID int, req *CompletionReportRequest) error {
	_, err := db.DB.Exec(`
		INSERT INTO completion_reports (complaint_id, governor_id, final_photos, completion_date, cost, cost_details, work_details, status)
		VALUES (?, ?, ?, ?, ?, ?, ?, 'pending')`,
		complaintID, governorID, req.FinalPhotos, req.CompletionDate, req.Cost, req.CostDetails, req.WorkDetails,
	)

	if err != nil {
		return err
	}

	// Update status complaint
	_, err = db.DB.Exec(`
		UPDATE complaints SET status = 'completion_report_submitted', updated_at = NOW() WHERE id = ?`,
		complaintID,
	)

	return err
}

// GetDashboardStats untuk admin
func (s *ComplaintService) GetDashboardStats() (*DashboardStats, error) {
	var stats DashboardStats

	db.DB.QueryRow("SELECT COUNT(*) FROM users WHERE role = 'user'").Scan(&stats.TotalWarga)
	db.DB.QueryRow("SELECT COUNT(*) FROM users WHERE role = 'investigator'").Scan(&stats.TotalInvestigator)
	db.DB.QueryRow("SELECT COUNT(*) FROM users WHERE role = 'governor'").Scan(&stats.TotalGovernor)
	db.DB.QueryRow("SELECT COUNT(*) FROM complaints").Scan(&stats.TotalComplaints)
	db.DB.QueryRow("SELECT COUNT(*) FROM complaints WHERE status IN ('pending_governor', 'investigation_assigned')").Scan(&stats.NeedAttention)
	db.DB.QueryRow("SELECT COUNT(*) FROM complaints WHERE status IN ('process_report_submitted', 'completion_report_submitted')").Scan(&stats.NeedVerification)
	db.DB.QueryRow("SELECT COUNT(*) FROM complaints WHERE status = 'completed'").Scan(&stats.Completed)
	db.DB.QueryRow("SELECT COUNT(*) FROM complaints WHERE DATE(created_at) = CURDATE()").Scan(&stats.TodayComplaints)
	db.DB.QueryRow("SELECT COUNT(*) FROM publications WHERE DATE(published_at) = CURDATE()").Scan(&stats.TodayPublications)

	return &stats, nil
}

// Helper scan complaint
func (s *ComplaintService) scanComplaint(rows *sql.Rows) (Complaint, error) {
	var c Complaint
	var regencyID, districtID, villageID sql.NullInt64
	var regencyName, districtName, villageName, photo, rejectedReason sql.NullString
	var assignedInvID sql.NullInt64
	var assignedInvName, investigationResult, investigationEvidence sql.NullString

	err := rows.Scan(
		&c.ID, &c.TrackingCode, &c.UserID, &c.UserName, &c.UserFullname,
		&c.ProvinceID, &c.ProvinceName,
		&regencyID, &regencyName,
		&districtID, &districtName,
		&villageID, &villageName,
		&c.LocationDetail,
		&c.CategoryID, &c.CategoryName,
		&c.Description, &photo,
		&c.Status, &rejectedReason,
		&assignedInvID, &assignedInvName,
		&investigationResult, &investigationEvidence,
		&c.CreatedAt, &c.UpdatedAt,
	)

	if err != nil {
		return c, err
	}

	if regencyID.Valid {
		c.RegencyID = intPtr(int(regencyID.Int64))
		if regencyName.Valid {
			c.RegencyName = &regencyName.String
		}
	}
	if districtID.Valid {
		c.DistrictID = intPtr(int(districtID.Int64))
		if districtName.Valid {
			c.DistrictName = &districtName.String
		}
	}
	if villageID.Valid {
		c.VillageID = intPtr(int(villageID.Int64))
		if villageName.Valid {
			c.VillageName = &villageName.String
		}
	}
	if photo.Valid {
		c.Photo = &photo.String
	}
	if rejectedReason.Valid {
		c.RejectedReason = &rejectedReason.String
	}
	if assignedInvID.Valid {
		c.AssignedInvestigatorID = intPtr(int(assignedInvID.Int64))
		if assignedInvName.Valid {
			c.AssignedInvestigatorName = &assignedInvName.String
		}
	}
	if investigationResult.Valid {
		c.InvestigationResult = &investigationResult.String
	}
	if investigationEvidence.Valid {
		c.InvestigationEvidence = &investigationEvidence.String
	}

	c.StatusText = getStatusText(c.Status)

	return c, nil
}

func (s *ComplaintService) scanComplaintRow(row *sql.Row) (*Complaint, error) {
	var c Complaint
	var regencyID, districtID, villageID sql.NullInt64
	var regencyName, districtName, villageName, photo, rejectedReason sql.NullString
	var assignedInvID sql.NullInt64
	var assignedInvName, investigationResult, investigationEvidence sql.NullString

	err := row.Scan(
		&c.ID, &c.TrackingCode, &c.UserID, &c.UserName, &c.UserFullname,
		&c.ProvinceID, &c.ProvinceName,
		&regencyID, &regencyName,
		&districtID, &districtName,
		&villageID, &villageName,
		&c.LocationDetail,
		&c.CategoryID, &c.CategoryName,
		&c.Description, &photo,
		&c.Status, &rejectedReason,
		&assignedInvID, &assignedInvName,
		&investigationResult, &investigationEvidence,
		&c.CreatedAt, &c.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	if regencyID.Valid {
		c.RegencyID = intPtr(int(regencyID.Int64))
		if regencyName.Valid {
			c.RegencyName = &regencyName.String
		}
	}
	if districtID.Valid {
		c.DistrictID = intPtr(int(districtID.Int64))
		if districtName.Valid {
			c.DistrictName = &districtName.String
		}
	}
	if villageID.Valid {
		c.VillageID = intPtr(int(villageID.Int64))
		if villageName.Valid {
			c.VillageName = &villageName.String
		}
	}
	if photo.Valid {
		c.Photo = &photo.String
	}
	if rejectedReason.Valid {
		c.RejectedReason = &rejectedReason.String
	}
	if assignedInvID.Valid {
		c.AssignedInvestigatorID = intPtr(int(assignedInvID.Int64))
		if assignedInvName.Valid {
			c.AssignedInvestigatorName = &assignedInvName.String
		}
	}
	if investigationResult.Valid {
		c.InvestigationResult = &investigationResult.String
	}
	if investigationEvidence.Valid {
		c.InvestigationEvidence = &investigationEvidence.String
	}

	c.StatusText = getStatusText(c.Status)

	return &c, nil
}

func getStatusText(status string) string {
	statusMap := map[string]string{
		"pending_governor":            "Menunggu Gubernur",
		"investigation_assigned":      "Investigasi Ditugaskan",
		"investigation_done":          "Investigasi Selesai",
		"governor_processing":         "Diproses Gubernur",
		"process_report_submitted":    "Laporan Proses Dikirim",
		"process_report_verified":     "Laporan Proses Diverifikasi",
		"completion_report_submitted": "Laporan Akhir Dikirim",
		"completed":                   "Selesai",
		"rejected":                    "Ditolak",
	}
	if val, ok := statusMap[status]; ok {
		return val
	}
	return status
}

// GetDashboardCharts - ambil data chart
func (s *ComplaintService) GetDashboardCharts() (*ChartData, error) {
	charts := &ChartData{
		MonthlyComplaints:    []MonthlyComplaint{},
		ComplaintsByStatus:   []StatusCount{},
		ComplaintsByCategory: []CategoryCount{},
	}
	
	// Monthly complaints - QUERY YANG DIPERBAIKI
	rows, err := db.DB.Query(`
		SELECT 
			DATE_FORMAT(created_at, '%b') as month,
			COUNT(*) as total
		FROM complaints
		GROUP BY YEAR(created_at), MONTH(created_at), DATE_FORMAT(created_at, '%b')
		ORDER BY MIN(created_at) ASC
	`)
	if err != nil {
		return charts, nil // Return empty charts if error
	}
	defer rows.Close()
	
	for rows.Next() {
		var month string
		var total int
		rows.Scan(&month, &total)
		charts.MonthlyComplaints = append(charts.MonthlyComplaints, MonthlyComplaint{
			Month: month,
			Total: total,
		})
	}
	
	// Complaints by status
	statusRows, err := db.DB.Query(`
		SELECT 
			CASE status
				WHEN 'pending_governor' THEN 'Menunggu Gubernur'
				WHEN 'investigation_assigned' THEN 'Investigasi'
				WHEN 'investigation_done' THEN 'Investigasi Selesai'
				WHEN 'governor_processing' THEN 'Diproses Gubernur'
				WHEN 'completed' THEN 'Selesai'
				WHEN 'rejected' THEN 'Ditolak'
				ELSE status
			END as name,
			COUNT(*) as value
		FROM complaints
		GROUP BY status
	`)
	if err == nil {
		defer statusRows.Close()
		for statusRows.Next() {
			var name string
			var value int
			statusRows.Scan(&name, &value)
			charts.ComplaintsByStatus = append(charts.ComplaintsByStatus, StatusCount{
				Name:  name,
				Value: value,
			})
		}
	}
	
	// Complaints by category
	catRows, err := db.DB.Query(`
		SELECT c.name, COUNT(*) as count
		FROM complaints co
		JOIN categories c ON co.category_id = c.id
		GROUP BY co.category_id, c.name
		ORDER BY count DESC
		LIMIT 6
	`)
	if err == nil {
		defer catRows.Close()
		for catRows.Next() {
			var name string
			var count int
			catRows.Scan(&name, &count)
			charts.ComplaintsByCategory = append(charts.ComplaintsByCategory, CategoryCount{
				Name:  name,
				Count: count,
			})
		}
	}
	
	return charts, nil
}

// GetAllUsers - ambil semua user
func (s *ComplaintService) GetAllUsers(role, search string, page, limit int) ([]UserResponse, int, error) {
	offset := (page - 1) * limit
	
	baseQuery := "FROM users WHERE 1=1"
	args := []interface{}{}
	
	if role != "" && role != "all" {
		baseQuery += " AND role = ?"
		args = append(args, role)
	}
	
	if search != "" {
		baseQuery += " AND (username LIKE ? OR fullname LIKE ? OR email LIKE ?)"
		searchTerm := "%" + search + "%"
		args = append(args, searchTerm, searchTerm, searchTerm)
	}
	
	var total int
	err := db.DB.QueryRow("SELECT COUNT(*) "+baseQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, err
	}
	
	dataQuery := "SELECT id, username, fullname, email, role, COALESCE(province_api_id, 0) as province_api_id, is_active, created_at " + baseQuery + " ORDER BY created_at DESC LIMIT ? OFFSET ?"
	dataArgs := append(args, limit, offset)
	
	rows, err := db.DB.Query(dataQuery, dataArgs...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	
	var users []UserResponse
	for rows.Next() {
		var u UserResponse
		var provinceID int
		rows.Scan(&u.ID, &u.Username, &u.Fullname, &u.Email, &u.Role, &provinceID, &u.IsActive, &u.CreatedAt)
		if provinceID > 0 {
			u.ProvinceApiID = &provinceID
		}
		users = append(users, u)
	}
	
	return users, total, nil
}

// GetRecentUsers - ambil 5 user terbaru
func (s *ComplaintService) GetRecentUsers() ([]RecentUser, error) {
	rows, err := db.DB.Query(`
		SELECT id, username, fullname, role, created_at
		FROM users 
		ORDER BY created_at DESC 
		LIMIT 5
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var users []RecentUser
	for rows.Next() {
		var u RecentUser
		rows.Scan(&u.ID, &u.Username, &u.Fullname, &u.Role, &u.CreatedAt)
		users = append(users, u)
	}
	
	return users, nil
}

// UpdateUser - update user
func (s *ComplaintService) UpdateUser(userID int, role string, provinceApiID *int) error {
	_, err := db.DB.Exec(`
		UPDATE users 
		SET role = ?, province_api_id = ?, updated_at = NOW()
		WHERE id = ?
	`, role, provinceApiID, userID)
	return err
}

// ToggleUserActive - toggle user active status
func (s *ComplaintService) ToggleUserActive(userID int, isActive bool) error {
	_, err := db.DB.Exec(`
		UPDATE users 
		SET is_active = ?, updated_at = NOW()
		WHERE id = ?
	`, isActive, userID)
	return err
}

// GetAllComplaintsForAdmin - ambil semua complaint untuk admin
func (s *ComplaintService) GetAllComplaintsForAdmin(status, search string, page, limit int) ([]Complaint, int, error) {
	offset := (page - 1) * limit
	
	query := `
		SELECT id, tracking_code, user_id, description, status, created_at
		FROM complaints
		WHERE 1=1
	`
	countQuery := "SELECT COUNT(*) FROM complaints WHERE 1=1"
	args := []interface{}{}
	
	if status != "" && status != "all" {
		query += " AND status = ?"
		countQuery += " AND status = ?"
		args = append(args, status)
	}
	
	if search != "" {
		query += " AND (tracking_code LIKE ? OR description LIKE ?)"
		countQuery += " AND (tracking_code LIKE ? OR description LIKE ?)"
		searchTerm := "%" + search + "%"
		args = append(args, searchTerm, searchTerm)
	}
	
	var total int
	err := db.DB.QueryRow(countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, err
	}
	
	query += " ORDER BY created_at DESC LIMIT ? OFFSET ?"
	args = append(args, limit, offset)
	
	rows, err := db.DB.Query(query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	
	var complaints []Complaint
	for rows.Next() {
		var c Complaint
		rows.Scan(&c.ID, &c.TrackingCode, &c.UserID, &c.Description, &c.Status, &c.CreatedAt)
		c.StatusText = getStatusText(c.Status)
		complaints = append(complaints, c)
	}
	
	return complaints, total, nil
}

// GetRecentComplaints - ambil 5 complaint terbaru
func (s *ComplaintService) GetRecentComplaints() ([]RecentComplaint, error) {
	rows, err := db.DB.Query(`
		SELECT c.id, c.tracking_code, c.description, c.status, c.created_at, 
			COALESCE(u.username, '') as user_name, COALESCE(u.fullname, '') as user_fullname
		FROM complaints c
		JOIN users u ON c.user_id = u.id
		ORDER BY c.created_at DESC 
		LIMIT 5
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var complaints []RecentComplaint
	for rows.Next() {
		var rc RecentComplaint
		rows.Scan(&rc.ID, &rc.TrackingCode, &rc.Description, &rc.Status, &rc.CreatedAt, &rc.UserName, &rc.UserFullname)
		rc.StatusText = getStatusText(rc.Status)
		complaints = append(complaints, rc)
	}
	
	return complaints, nil
}

// VerifyProcessReport - verifikasi laporan proses
func (s *ComplaintService) VerifyProcessReport(reportID int, status string, adminNote *string, adminID int) error {
	newStatus := "process_report_verified"
	if status == "rejected" {
		newStatus = "process_report_rejected"
	}
	
	_, err := db.DB.Exec(`
		UPDATE process_reports 
		SET status = ?, admin_notes = ?, verified_by = ?, verified_at = NOW()
		WHERE id = ?
	`, newStatus, adminNote, adminID, reportID)
	
	if err == nil && status == "verified" {
		var complaintID int
		db.DB.QueryRow("SELECT complaint_id FROM process_reports WHERE id = ?", reportID).Scan(&complaintID)
		db.DB.Exec("UPDATE complaints SET status = 'governor_processing' WHERE id = ?", complaintID)
	}
	
	return err
}

// VerifyCompletionReport - verifikasi laporan akhir
func (s *ComplaintService) VerifyCompletionReport(reportID int, status string, adminNote *string, adminID int) error {
	newStatus := "completion_report_verified"
	if status == "rejected" {
		newStatus = "completion_report_rejected"
	}
	
	_, err := db.DB.Exec(`
		UPDATE completion_reports 
		SET status = ?, admin_notes = ?, verified_by = ?, verified_at = NOW()
		WHERE id = ?
	`, newStatus, adminNote, adminID, reportID)
	
	if err == nil && status == "verified" {
		var complaintID int
		db.DB.QueryRow("SELECT complaint_id FROM completion_reports WHERE id = ?", reportID).Scan(&complaintID)
		db.DB.Exec("UPDATE complaints SET status = 'completed' WHERE id = ?", complaintID)
	}
	
	return err
}

func intPtr(i int) *int {
	return &i
}