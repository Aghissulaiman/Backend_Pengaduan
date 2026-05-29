package complaint

import (
    "database/sql"
    "errors"
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

// SubmitComplaint warga submit pengaduan baru
func (s *ComplaintService) SubmitComplaint(userID int, req *SubmitComplaintRequest) (*SubmitComplaintResponse, error) {
    // Generate tracking code
    trackingCode := utils.GenerateTrackingCode()

    // Debug log
    fmt.Printf("=== SUBMIT COMPLAINT ===\n")
    fmt.Printf("UserID: %d\n", userID)
    fmt.Printf("ProvinceApiID: %d\n", req.ProvinceID)
    fmt.Printf("RegencyID: %v\n", req.RegencyID)
    fmt.Printf("DistrictID: %v\n", req.DistrictID)
    fmt.Printf("VillageID: %v\n", req.VillageID)
    fmt.Printf("LocationDetail: %s\n", req.LocationDetail)
    fmt.Printf("CategoryID: %d\n", req.CategoryID)
    fmt.Printf("Description: %s\n", req.Description)
    fmt.Printf("Photo: %v\n", req.Photo)

    result, err := db.DB.Exec(`
        INSERT INTO complaints (
            tracking_code, user_id, province_api_id, regency_id, district_id, village_id,
            location_detail, category_id, description, photo, status
        ) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, 'pending_governor')`,
        trackingCode, userID, req.ProvinceID, req.RegencyID, req.DistrictID,
        req.VillageID, req.LocationDetail, req.CategoryID, req.Description, req.Photo,
    )

    if err != nil {
        fmt.Printf("SQL Error: %v\n", err)
        return nil, errors.New("gagal menyimpan pengaduan: " + err.Error())
    }

    id, _ := result.LastInsertId()

    fmt.Printf("Complaint created with ID: %d, TrackingCode: %s\n", id, trackingCode)

    // Insert notifikasi ke gubernur provinsi terkait
    go s.notifyGovernor(req.ProvinceID, int(id), trackingCode)

    return &SubmitComplaintResponse{
        TrackingCode: trackingCode,
        ID:           int(id),
    }, nil
}

func (s *ComplaintService) notifyGovernor(provinceID, complaintID int, trackingCode string) {
    // Cari gubernur provinsi tersebut
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
        SELECT c.id, c.tracking_code, c.user_id, u.username, c.province_api_id, p.name,
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
        c, err := s.scanComplaint(rows)
        if err != nil {
            continue
        }
        complaints = append(complaints, c)
    }

    var total int
    db.DB.QueryRow("SELECT COUNT(*) FROM complaints WHERE user_id = ?", userID).Scan(&total)

    return complaints, total, nil
}

// GetAllComplaints untuk admin/governor/investigator
func (s *ComplaintService) GetAllComplaints(role string, userID, provinceID int, query *GetComplaintsQuery) ([]Complaint, int, error) {
    offset := (query.Page - 1) * query.Limit

    sqlQuery := `
        SELECT c.id, c.tracking_code, c.user_id, u.username, c.province_api_id, p.name,
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
        WHERE 1=1
    `
    countQuery := "SELECT COUNT(*) FROM complaints WHERE 1=1"
    args := []interface{}{}

    // Filter berdasarkan role
    if role == "governor" {
        sqlQuery += " AND c.province_api_id = ?"
        countQuery += " AND province_api_id = ?"
        args = append(args, provinceID)
    } else if role == "investigator" {
        sqlQuery += " AND c.province_api_id = ? AND c.status IN ('investigation_assigned', 'pending_governor')"
        countQuery += " AND province_api_id = ? AND status IN ('investigation_assigned', 'pending_governor')"
        args = append(args, provinceID)
    }

    if query.Status != "" {
        sqlQuery += " AND c.status = ?"
        countQuery += " AND status = ?"
        args = append(args, query.Status)
    }

    if query.ProvinceID > 0 && role == "admin" {
        sqlQuery += " AND c.province_api_id = ?"
        countQuery += " AND province_api_id = ?"
        args = append(args, query.ProvinceID)
    }

    sqlQuery += " ORDER BY c.created_at DESC LIMIT ? OFFSET ?"
    args = append(args, query.Limit, offset)

    rows, err := db.DB.Query(sqlQuery, args...)
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

    var total int
    var countArgs []interface{}
    if role == "governor" {
        countArgs = append(countArgs, provinceID)
    } else if role == "investigator" {
        countArgs = append(countArgs, provinceID)
    }
    if query.Status != "" {
        countArgs = append(countArgs, query.Status)
    }
    if query.ProvinceID > 0 && role == "admin" {
        countArgs = append(countArgs, query.ProvinceID)
    }
    db.DB.QueryRow(countQuery, countArgs...).Scan(&total)

    return complaints, total, nil
}

// GetComplaintByID ambil detail pengaduan
func (s *ComplaintService) GetComplaintByID(id int) (*Complaint, error) {
    row := db.DB.QueryRow(`
        SELECT c.id, c.tracking_code, c.user_id, u.username, c.province_api_id, p.name,
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
        WHERE c.id = ?`, id)

    complaint, err := s.scanComplaintRow(row)
    if err != nil {
        return nil, errors.New("pengaduan tidak ditemukan")
    }

    return complaint, nil
}

// GetComplaintByTrackingCode cek status publik
func (s *ComplaintService) GetComplaintByTrackingCode(trackingCode string) (*Complaint, error) {
    row := db.DB.QueryRow(`
        SELECT c.id, c.tracking_code, c.user_id, u.username, c.province_api_id, p.name,
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

// Helper scan complaint
func (s *ComplaintService) scanComplaint(rows *sql.Rows) (Complaint, error) {
    var c Complaint
    var regencyID, districtID, villageID sql.NullInt64
    var regencyName, districtName, villageName, photo, rejectedReason sql.NullString
    var assignedInvID sql.NullInt64
    var assignedInvName, investigationResult, investigationEvidence sql.NullString

    err := rows.Scan(
        &c.ID, &c.TrackingCode, &c.UserID, &c.UserName,
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
        &c.ID, &c.TrackingCode, &c.UserID, &c.UserName,
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

// UpdateStatus - untuk berbagai role
func (s *ComplaintService) UpdateStatus(complaintID int, status string, note *string, userID int, role string) error {
    if !utils.IsValidStatus(status) {
        return errors.New("status tidak valid")
    }

    // Cek permission berdasarkan role
    var complaint Complaint
    err := db.DB.QueryRow("SELECT status, province_api_id FROM complaints WHERE id = ?", complaintID).
        Scan(&complaint.Status, &complaint.ProvinceID)
    if err != nil {
        return errors.New("pengaduan tidak ditemukan")
    }

    // Validasi role
    if role == "governor" && complaint.Status != "investigation_done" && status != "governor_processing" {
        return errors.New("gubernur hanya bisa memproses setelah investigasi selesai")
    }
    if role == "investigator" && complaint.Status != "investigation_assigned" && status != "investigation_done" {
        return errors.New("investigator hanya bisa mengubah status dari investigation_assigned ke investigation_done")
    }

    // Update status
    _, err = db.DB.Exec(`
        UPDATE complaints 
        SET status = ?, updated_at = NOW(),
            rejected_reason = CASE WHEN ? = 'rejected' AND ? IS NOT NULL THEN ? ELSE rejected_reason END,
            investigation_completed_at = CASE WHEN ? = 'investigation_done' THEN NOW() ELSE investigation_completed_at END,
            governor_processed_at = CASE WHEN ? = 'governor_processing' THEN NOW() ELSE governor_processed_at END
        WHERE id = ?`,
        status, status, note, note, status, status, complaintID,
    )

    if err != nil {
        return err
    }

    // Insert log
    db.DB.Exec(`
        INSERT INTO activity_logs (user_id, action, complaint_id, old_status, new_status)
        VALUES (?, 'UPDATE_STATUS', ?, ?, ?)`,
        userID, complaintID, complaint.Status, status,
    )

    return nil
}

// AssignInvestigator - governor tugaskan investigator
func (s *ComplaintService) AssignInvestigator(complaintID, investigatorID int, governorID int) error {
    // Cek apakah investigator valid dan satu provinsi
    var investigatorProvinceID int
    err := db.DB.QueryRow("SELECT province_api_id FROM users WHERE id = ? AND role = 'investigator' AND is_active = TRUE", investigatorID).
        Scan(&investigatorProvinceID)
    if err != nil {
        return errors.New("investigator tidak valid")
    }

    var complaintProvinceID int
    db.DB.QueryRow("SELECT province_api_id FROM complaints WHERE id = ?", complaintID).Scan(&complaintProvinceID)

    if investigatorProvinceID != complaintProvinceID {
        return errors.New("investigator harus dari provinsi yang sama")
    }

    _, err = db.DB.Exec(`
        UPDATE complaints 
        SET assigned_investigator_id = ?, status = 'investigation_assigned', assigned_investigator_at = NOW()
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
            status = ?, rejected_reason = COALESCE(?, rejected_reason),
            investigation_completed_at = NOW()
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
        UPDATE complaints SET status = 'process_report_submitted' WHERE id = ?`,
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
        UPDATE complaints SET status = 'completion_report_submitted' WHERE id = ?`,
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

// Helper functions
func getStatusText(status string) string {
    statusMap := map[string]string{
        "pending_governor":           "Menunggu Gubernur",
        "investigation_assigned":     "Investigasi Ditugaskan",
        "investigation_done":         "Investigasi Selesai",
        "governor_processing":        "Diproses Gubernur",
        "process_report_submitted":   "Laporan Proses Dikirim",
        "process_report_verified":    "Laporan Proses Diverifikasi",
        "completion_report_submitted": "Laporan Akhir Dikirim",
        "completed":                  "Selesai",
    }
    if val, ok := statusMap[status]; ok {
        return val
    }
    return status
}

func intPtr(i int) *int {
    return &i
}

type SubmitComplaintResponse struct {
    TrackingCode string `json:"tracking_code"`
    ID           int    `json:"id"`
}