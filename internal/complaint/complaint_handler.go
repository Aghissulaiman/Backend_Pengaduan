package complaint

import (
	"bytes"
	"io"
	"net/http"
	"strconv"
	"time"   
	"fmt"
	"database/sql"
	"github.com/gin-gonic/gin"
	"pengaduan_be2/internal/dto"
	"pengaduan_be2/pkg/db" // TAMBAHKAN IMPORT INI
)

type ComplaintHandler struct {
	service *ComplaintService
}

func NewComplaintHandler() *ComplaintHandler {
	return &ComplaintHandler{service: NewComplaintService()}
}

// GetCategories - GET /api/complaints/categories
func (h *ComplaintHandler) GetCategories(c *gin.Context) {
	categories, err := h.service.GetCategories()
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.Response{
			Success: false,
			Message: "Gagal mengambil data kategori",
		})
		return
	}

	c.JSON(http.StatusOK, dto.Response{
		Success: true,
		Data:    categories,
	})
}

// SubmitComplaint - POST /api/complaints/submit
func (h *ComplaintHandler) SubmitComplaint(c *gin.Context) {
	bodyBytes, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.Response{
			Success: false,
			Message: "Gagal membaca request body: " + err.Error(),
		})
		return
	}

	c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

	var req SubmitComplaintRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.Response{
			Success: false,
			Message: "Format JSON tidak valid: " + err.Error(),
		})
		return
	}

	userIDVal, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, dto.Response{
			Success: false,
			Message: "User tidak terautentikasi",
		})
		return
	}

	userID, ok := userIDVal.(int)
	if !ok {
		c.JSON(http.StatusUnauthorized, dto.Response{
			Success: false,
			Message: "User ID tidak valid",
		})
		return
	}

	resp, err := h.service.SubmitComplaint(userID, &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.Response{
			Success: false,
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, dto.Response{
		Success: true,
		Message: "Pengaduan berhasil dikirim",
		Data:    resp,
	})
}

// GetGovernorReports - GET /api/governor/reports
func (h *ComplaintHandler) GetGovernorReports(c *gin.Context) {
	provinceApiID, exists := c.Get("province_api_id")
	if !exists {
		c.JSON(http.StatusBadRequest, dto.Response{
			Success: false,
			Message: "Data provinsi tidak ditemukan",
		})
		return
	}

	var query ReportQuery
	query.Period = c.DefaultQuery("period", "month")
	query.StartDate = c.Query("start_date")
	query.EndDate = c.Query("end_date")
	query.Status = c.Query("status")
	query.Search = c.Query("search")
	query.Page, _ = strconv.Atoi(c.DefaultQuery("page", "1"))
	query.Limit, _ = strconv.Atoi(c.DefaultQuery("limit", "20"))

	complaints, stats, total, err := h.service.GetGovernorReports(provinceApiID.(int), &query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.Response{
			Success: false,
			Message: "Gagal mengambil data laporan: " + err.Error(),
		})
		return
	}

	// 🔥 Konversi complaints ke format yang lebih simpel untuk frontend
	var reportList []map[string]interface{}
	for _, comp := range complaints {
		reportList = append(reportList, map[string]interface{}{
			"id":              comp.ID,
			"tracking_code":   comp.TrackingCode,
			"description":     comp.Description,
			"location_detail": comp.LocationDetail,
			"status":          comp.Status,
			"status_text":     comp.StatusText,
			"created_at":      comp.CreatedAt,
			"user_name":       comp.UserName,
			"category_name":   comp.CategoryName,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"complaints": reportList,
			"stats":      stats,
			"total":      total,
			"page":       query.Page,
			"limit":      query.Limit,
		},
	})
}

// GetGovernorInvestigations - GET /api/governor/investigations
func (h *ComplaintHandler) GetGovernorInvestigations(c *gin.Context) {
	provinceApiID, exists := c.Get("province_api_id")
	if !exists {
		c.JSON(http.StatusBadRequest, dto.Response{
			Success: false,
			Message: "Data provinsi tidak ditemukan",
		})
		return
	}

	var query GetInvestigationsQuery
	query.Status = c.Query("status")
	query.Search = c.Query("search")
	query.Page, _ = strconv.Atoi(c.DefaultQuery("page", "1"))
	query.Limit, _ = strconv.Atoi(c.DefaultQuery("limit", "10"))

	investigations, total, err := h.service.GetGovernorInvestigations(provinceApiID.(int), &query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.Response{
			Success: false,
			Message: "Gagal mengambil data investigasi",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"investigations": investigations,
			"total":          total,
			"page":           query.Page,
			"limit":          query.Limit,
		},
	})
}

// GetMyComplaints - GET /api/complaints/my
func (h *ComplaintHandler) GetMyComplaints(c *gin.Context) {
	userIDVal, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, dto.Response{
			Success: false,
			Message: "User tidak terautentikasi",
		})
		return
	}

	userID, ok := userIDVal.(int)
	if !ok {
		c.JSON(http.StatusUnauthorized, dto.Response{
			Success: false,
			Message: "User ID tidak valid",
		})
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	complaints, total, err := h.service.GetUserComplaints(userID, page, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.Response{
			Success: false,
			Message: "Gagal mengambil data",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":    true,
		"data":       complaints,
		"pagination": dto.NewPagination(page, limit, total),
	})
}

// CheckStatus - GET /api/complaints/status/:tracking_code
func (h *ComplaintHandler) CheckStatus(c *gin.Context) {
	trackingCode := c.Param("tracking_code")
	complaint, err := h.service.GetComplaintByTrackingCode(trackingCode)
	if err != nil {
		c.JSON(http.StatusNotFound, dto.Response{
			Success: false,
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, dto.Response{
		Success: true,
		Data:    complaint,
	})
}

// GetAllComplaints - GET /api/complaints/all
func (h *ComplaintHandler) GetAllComplaints(c *gin.Context) {
	role, _ := c.Get("role")
	userID, _ := c.Get("user_id")
	provinceID, _ := c.Get("province_api_id")

	var query GetComplaintsQuery
	query.Status = c.Query("status")
	query.Search = c.Query("search")

	if provinceApiID := c.Query("province_api_id"); provinceApiID != "" {
		query.ProvinceID, _ = strconv.Atoi(provinceApiID)
	}
	query.Page, _ = strconv.Atoi(c.DefaultQuery("page", "1"))
	query.Limit, _ = strconv.Atoi(c.DefaultQuery("limit", "10"))

	provinceIDVal := 0
	if provinceID != nil {
		provinceIDVal = provinceID.(int)
	}

	complaints, total, err := h.service.GetAllComplaints(
		role.(string), userID.(int), provinceIDVal, &query,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.Response{
			Success: false,
			Message: "Gagal mengambil data",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"complaints": complaints,
			"total":      total,
			"page":       query.Page,
			"limit":      query.Limit,
		},
	})
}

// GetComplaintDetail - GET /api/complaints/:id
func (h *ComplaintHandler) GetComplaintDetail(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.Response{
			Success: false,
			Message: "ID tidak valid",
		})
		return
	}

	complaint, err := h.service.GetComplaintByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, dto.Response{
			Success: false,
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, dto.Response{
		Success: true,
		Data:    complaint,
	})
}

// UpdateComplaintStatus - PATCH /api/complaints/:id/status
func (h *ComplaintHandler) UpdateComplaintStatus(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.Response{
			Success: false,
			Message: "ID tidak valid",
		})
		return
	}

	var req UpdateStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.Response{
			Success: false,
			Message: err.Error(),
		})
		return
	}

	userID, _ := c.Get("user_id")
	role, _ := c.Get("role")

	if err := h.service.UpdateStatus(id, req.Status, req.RejectReason, userID.(int), role.(string)); err != nil {
		c.JSON(http.StatusBadRequest, dto.Response{
			Success: false,
			Message: err.Error(),
		})
		return
	}

	message := "Status berhasil diupdate"
	if req.Status == "rejected" {
		message = "Pengaduan ditolak"
	} else if req.Status == "investigation_assigned" {
		message = "Pengaduan diterima dan investigator ditugaskan"
	}

	c.JSON(http.StatusOK, dto.Response{
		Success: true,
		Message: message,
	})
}

// GetGovernorStats - GET /api/governor/dashboard/stats
func (h *ComplaintHandler) GetGovernorStats(c *gin.Context) {
	provinceApiID, exists := c.Get("province_api_id")
	if !exists {
		c.JSON(http.StatusBadRequest, dto.Response{
			Success: false,
			Message: "Data provinsi tidak ditemukan",
		})
		return
	}

	stats, err := h.service.GetGovernorStats(provinceApiID.(int))
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.Response{
			Success: false,
			Message: "Gagal mengambil statistik",
		})
		return
	}

	c.JSON(http.StatusOK, dto.Response{
		Success: true,
		Data:    stats,
	})
}

// GetInvestigators - GET /api/governor/investigators
func (h *ComplaintHandler) GetInvestigators(c *gin.Context) {
	provinceApiID, exists := c.Get("province_api_id")
	if !exists {
		c.JSON(http.StatusBadRequest, dto.Response{
			Success: false,
			Message: "Data provinsi tidak ditemukan",
		})
		return
	}

	investigators, err := h.service.GetInvestigatorsByProvince(provinceApiID.(int))
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.Response{
			Success: false,
			Message: "Gagal mengambil data investigator",
		})
		return
	}

	c.JSON(http.StatusOK, dto.Response{
		Success: true,
		Data:    investigators,
	})
}

// 👇👇👇 TAMBAHKAN HANDLER INI 👇👇👇
// GetGovernorComplaints - GET /api/governor/complaints
func (h *ComplaintHandler) GetGovernorComplaints(c *gin.Context) {
	// Ambil user_id dari context
	userIDVal, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, dto.Response{
			Success: false,
			Message: "User tidak terautentikasi",
		})
		return
	}

	userID := userIDVal.(int)

	// Ambil province_api_id dari database langsung (paling aman)
	var provinceApiID int
	err := db.DB.QueryRow(`
		SELECT COALESCE(province_api_id, 0) FROM users 
		WHERE id = ? AND role = 'governor'`,
		userID,
	).Scan(&provinceApiID)

	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.Response{
			Success: false,
			Message: "Gagal mengambil data provinsi: " + err.Error(),
		})
		return
	}

	if provinceApiID == 0 {
		c.JSON(http.StatusBadRequest, dto.Response{
			Success: false,
			Message: "Provinsi tidak ditemukan untuk user ini",
		})
		return
	}

	// Parse query params
	var query GovernorComplaintsQuery
	query.Status = c.Query("status")
	query.Search = c.Query("search")
	query.Page, _ = strconv.Atoi(c.DefaultQuery("page", "1"))
	query.Limit, _ = strconv.Atoi(c.DefaultQuery("limit", "10"))

	complaints, total, err := h.service.GetGovernorComplaints(provinceApiID, &query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.Response{
			Success: false,
			Message: "Gagal mengambil data pengaduan: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"complaints": complaints,
			"total":      total,
			"page":       query.Page,
			"limit":      query.Limit,
		},
	})
}
// 👆👆👆 TAMBAHKAN HANDLER INI 👆👆👆

// AssignInvestigator - POST /api/governor/complaints/:id/assign
func (h *ComplaintHandler) AssignInvestigator(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.Response{
			Success: false,
			Message: "ID tidak valid",
		})
		return
	}

	var req AssignInvestigatorRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.Response{
			Success: false,
			Message: err.Error(),
		})
		return
	}

	governorID, _ := c.Get("user_id")
	if err := h.service.AssignInvestigator(id, req.InvestigatorID, governorID.(int)); err != nil {
		c.JSON(http.StatusBadRequest, dto.Response{
			Success: false,
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, dto.Response{
		Success: true,
		Message: "Investigator berhasil ditugaskan",
	})
}

// GetInvestigatorComplaints - GET /api/investigator/complaints
func (h *ComplaintHandler) GetInvestigatorComplaints(c *gin.Context) {
	userIDVal, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, dto.Response{
			Success: false,
			Message: "User tidak terautentikasi",
		})
		return
	}

	investigatorID := userIDVal.(int)

	var query InvestigatorComplaintsQuery
	query.Status = c.Query("status")
	query.Search = c.Query("search")
	query.Page, _ = strconv.Atoi(c.DefaultQuery("page", "1"))
	query.Limit, _ = strconv.Atoi(c.DefaultQuery("limit", "10"))

	complaints, total, err := h.service.GetInvestigatorComplaints(investigatorID, &query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.Response{
			Success: false,
			Message: "Gagal mengambil data pengaduan: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"complaints": complaints,
			"total":      total,
			"page":       query.Page,
			"limit":      query.Limit,
		},
	})
}

// GetInvestigatorStats - GET /api/investigator/dashboard/stats
func (h *ComplaintHandler) GetInvestigatorStats(c *gin.Context) {
	userIDVal, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, dto.Response{
			Success: false,
			Message: "User tidak terautentikasi",
		})
		return
	}

	investigatorID := userIDVal.(int)

	stats, err := h.service.GetInvestigatorStats(investigatorID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.Response{
			Success: false,
			Message: "Gagal mengambil statistik",
		})
		return
	}

	c.JSON(http.StatusOK, dto.Response{
		Success: true,
		Data:    stats,
	})
}

// SubmitInvestigationResultExtended - POST /api/investigator/complaints/:id/result (extended)
func (h *ComplaintHandler) SubmitInvestigationResultExtended(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.Response{
			Success: false,
			Message: "ID tidak valid",
		})
		return
	}

	var req struct {
		Result   string  `json:"result" binding:"required"`
		Evidence *string `json:"evidence"`
		IsValid  bool    `json:"is_valid"`
		Notes    string  `json:"notes"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.Response{
			Success: false,
			Message: err.Error(),
		})
		return
	}

	investigatorID, _ := c.Get("user_id")
	evidence := ""
	if req.Evidence != nil {
		evidence = *req.Evidence
	}

	if err := h.service.SubmitInvestigationResultWithNote(id, req.Result, evidence, req.IsValid, req.Notes, investigatorID.(int)); err != nil {
		c.JSON(http.StatusBadRequest, dto.Response{
			Success: false,
			Message: err.Error(),
		})
		return
	}

	message := "Hasil investigasi berhasil dikirim"
	if !req.IsValid {
		message = "Laporan dinyatakan tidak valid dan akan ditolak"
	}

	c.JSON(http.StatusOK, dto.Response{
		Success: true,
		Message: message,
	})
}

// SubmitInvestigationResult - POST /api/investigator/complaints/:id/result
func (h *ComplaintHandler) SubmitInvestigationResult(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.Response{
			Success: false,
			Message: "ID tidak valid",
		})
		return
	}

	var req InvestigationResultRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.Response{
			Success: false,
			Message: err.Error(),
		})
		return
	}

	investigatorID, _ := c.Get("user_id")
	evidence := ""
	if req.Evidence != nil {
		evidence = *req.Evidence
	}
	if err := h.service.SubmitInvestigationResult(id, req.Result, evidence, req.IsValid, investigatorID.(int)); err != nil {
		c.JSON(http.StatusBadRequest, dto.Response{
			Success: false,
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, dto.Response{
		Success: true,
		Message: "Hasil investigasi berhasil dikirim",
	})
}

// SubmitProcessReport - POST /api/governor/complaints/:id/process-report
func (h *ComplaintHandler) SubmitProcessReport(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.Response{
			Success: false,
			Message: "ID tidak valid",
		})
		return
	}

	var req ProcessReportRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.Response{
			Success: false,
			Message: err.Error(),
		})
		return
	}

	governorID, _ := c.Get("user_id")
	if err := h.service.SubmitProcessReport(id, governorID.(int), &req); err != nil {
		c.JSON(http.StatusBadRequest, dto.Response{
			Success: false,
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, dto.Response{
		Success: true,
		Message: "Laporan proses berhasil dikirim ke admin",
	})
}

// SubmitCompletionReport - POST /api/governor/complaints/:id/completion-report
func (h *ComplaintHandler) SubmitCompletionReport(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.Response{
			Success: false,
			Message: "ID tidak valid",
		})
		return
	}

	var req CompletionReportRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.Response{
			Success: false,
			Message: err.Error(),
		})
		return
	}

	governorID, _ := c.Get("user_id")
	if err := h.service.SubmitCompletionReport(id, governorID.(int), &req); err != nil {
		c.JSON(http.StatusBadRequest, dto.Response{
			Success: false,
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, dto.Response{
		Success: true,
		Message: "Laporan akhir berhasil dikirim ke admin",
	})
}

// GetDashboardStats - GET /api/admin/dashboard
func (h *ComplaintHandler) GetDashboardStats(c *gin.Context) {
	stats, err := h.service.GetDashboardStats()
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.Response{
			Success: false,
			Message: "Gagal mengambil data statistik",
		})
		return
	}

	c.JSON(http.StatusOK, dto.Response{
		Success: true,
		Data:    stats,
	})
}

// GetDashboardCharts - ambil data chart untuk dashboard admin
func (h *ComplaintHandler) GetDashboardCharts(c *gin.Context) {
	charts, err := h.service.GetDashboardCharts()
	if err != nil {
		c.JSON(500, gin.H{"success": false, "message": err.Error()})
		return
	}
	c.JSON(200, gin.H{"success": true, "data": charts})
}

// GetAllUsers - ambil semua user untuk admin
func (h *ComplaintHandler) GetAllUsers(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	role := c.Query("role")
	search := c.Query("search")
	
	users, total, err := h.service.GetAllUsers(role, search, page, limit)
	if err != nil {
		c.JSON(500, gin.H{"success": false, "message": err.Error()})
		return
	}
	
	c.JSON(200, gin.H{
		"success": true,
		"data": users,
		"total": total,
		"page": page,
		"limit": limit,
	})
}

// GetRecentUsers - ambil 5 user terbaru
func (h *ComplaintHandler) GetRecentUsers(c *gin.Context) {
	users, err := h.service.GetRecentUsers()
	if err != nil {
		c.JSON(500, gin.H{"success": false, "message": err.Error()})
		return
	}
	c.JSON(200, gin.H{"success": true, "data": users})
}

// UpdateUser - update user oleh admin
func (h *ComplaintHandler) UpdateUser(c *gin.Context) {
	userID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(400, gin.H{"success": false, "message": "ID user tidak valid"})
		return
	}
	
	var req struct {
		Role          string `json:"role"`
		ProvinceApiID *int   `json:"province_api_id"`
	}
	
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"success": false, "message": err.Error()})
		return
	}
	
	err = h.service.UpdateUser(userID, req.Role, req.ProvinceApiID)
	if err != nil {
		c.JSON(500, gin.H{"success": false, "message": err.Error()})
		return
	}
	
	c.JSON(200, gin.H{"success": true, "message": "User berhasil diupdate"})
}

// ToggleUserActive - aktifkan/nonaktifkan user
func (h *ComplaintHandler) ToggleUserActive(c *gin.Context) {
	userID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(400, gin.H{"success": false, "message": "ID user tidak valid"})
		return
	}
	
	var req struct {
		IsActive bool `json:"is_active"`
	}
	
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"success": false, "message": err.Error()})
		return
	}
	
	err = h.service.ToggleUserActive(userID, req.IsActive)
	if err != nil {
		c.JSON(500, gin.H{"success": false, "message": err.Error()})
		return
	}
	
	c.JSON(200, gin.H{"success": true, "message": "Status user berhasil diubah"})
}

// GetAllComplaintsForAdmin - ambil semua complaint untuk admin
func (h *ComplaintHandler) GetAllComplaintsForAdmin(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	status := c.Query("status")
	search := c.Query("search")
	
	complaints, total, err := h.service.GetAllComplaintsForAdmin(status, search, page, limit)
	if err != nil {
		c.JSON(500, gin.H{"success": false, "message": err.Error()})
		return
	}
	
	c.JSON(200, gin.H{
		"success": true,
		"data": complaints,
		"total": total,
		"page": page,
		"limit": limit,
	})
}

// GetRecentComplaints - ambil 5 complaint terbaru
func (h *ComplaintHandler) GetRecentComplaints(c *gin.Context) {
	complaints, err := h.service.GetRecentComplaints()
	if err != nil {
		c.JSON(500, gin.H{"success": false, "message": err.Error()})
		return
	}
	c.JSON(200, gin.H{"success": true, "data": complaints})
}

// GetComplaintDetailForAdmin - detail complaint untuk admin
func (h *ComplaintHandler) GetComplaintDetailForAdmin(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(400, gin.H{"success": false, "message": "ID tidak valid"})
		return
	}
	
	complaint, err := h.service.GetComplaintByID(id)
	if err != nil {
		c.JSON(404, gin.H{"success": false, "message": err.Error()})
		return
	}
	
	c.JSON(200, gin.H{"success": true, "data": complaint})
}

// VerifyProcessReport - verifikasi laporan proses
func (h *ComplaintHandler) VerifyProcessReport(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(400, gin.H{"success": false, "message": "ID tidak valid"})
		return
	}
	
	var req struct {
		Status    string  `json:"status" binding:"required,oneof=verified rejected"`
		AdminNote *string `json:"admin_note"`
	}
	
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"success": false, "message": err.Error()})
		return
	}
	
	adminID := c.GetInt("user_id")
	
	err = h.service.VerifyProcessReport(id, req.Status, req.AdminNote, adminID)
	if err != nil {
		c.JSON(500, gin.H{"success": false, "message": err.Error()})
		return
	}
	
	c.JSON(200, gin.H{"success": true, "message": "Laporan proses berhasil diverifikasi"})
}

// VerifyCompletionReport - verifikasi laporan akhir
func (h *ComplaintHandler) VerifyCompletionReport(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(400, gin.H{"success": false, "message": "ID tidak valid"})
		return
	}
	
	var req struct {
		Status    string  `json:"status" binding:"required,oneof=verified rejected"`
		AdminNote *string `json:"admin_note"`
	}
	
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"success": false, "message": err.Error()})
		return
	}
	
	adminID := c.GetInt("user_id")
	
	err = h.service.VerifyCompletionReport(id, req.Status, req.AdminNote, adminID)
	if err != nil {
		c.JSON(500, gin.H{"success": false, "message": err.Error()})
		return
	}
	
	c.JSON(200, gin.H{"success": true, "message": "Laporan akhir berhasil diverifikasi"})
}

// GetComplaintsCountByProvince - ambil jumlah pengaduan per provinsi
func (h *ComplaintHandler) GetComplaintsCountByProvince(c *gin.Context) {
	provinceApiIDStr := c.Query("province_api_id")
	if provinceApiIDStr == "" {
		c.JSON(400, gin.H{"success": false, "message": "province_api_id required"})
		return
	}
	
	provinceApiID, err := strconv.Atoi(provinceApiIDStr)
	if err != nil {
		c.JSON(400, gin.H{"success": false, "message": "invalid province_api_id"})
		return
	}
	
	var total int
	query := "SELECT COUNT(*) FROM complaints WHERE province_api_id = ?"
	err = db.DB.QueryRow(query, provinceApiID).Scan(&total)
	if err != nil {
		c.JSON(500, gin.H{"success": false, "message": err.Error()})
		return
	}
	
	c.JSON(200, gin.H{"success": true, "total": total})
}

// GetComplaintsByProvince - ambil complaints berdasarkan provinsi
func (h *ComplaintHandler) GetComplaintsByProvince(c *gin.Context) {
	provinceApiIDStr := c.Query("province_api_id")
	if provinceApiIDStr == "" {
		c.JSON(400, gin.H{"success": false, "message": "province_api_id required"})
		return
	}
	
	provinceApiID, err := strconv.Atoi(provinceApiIDStr)
	if err != nil {
		c.JSON(400, gin.H{"success": false, "message": "invalid province_api_id"})
		return
	}
	
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	status := c.Query("status")
	search := c.Query("search")
	
	offset := (page - 1) * limit
	
	// Query dengan filter province_api_id
	query := `
		SELECT 
			c.id, 
			c.tracking_code, 
			c.description, 
			c.location_detail, 
			c.status, 
			c.created_at,
			COALESCE(u.username, '') as user_name,
			COALESCE(u.fullname, '') as user_fullname,
			cat.name as category_name,
			p.name as province_name
		FROM complaints c
		JOIN users u ON c.user_id = u.id
		JOIN categories cat ON c.category_id = cat.id
		JOIN provinces p ON c.province_api_id = p.api_id
		WHERE c.province_api_id = ?
	`
	
	countQuery := "SELECT COUNT(*) FROM complaints WHERE province_api_id = ?"
	args := []interface{}{provinceApiID}
	countArgs := []interface{}{provinceApiID}
	
	// Filter status
	if status != "" && status != "all" {
		query += " AND c.status = ?"
		countQuery += " AND status = ?"
		args = append(args, status)
		countArgs = append(countArgs, status)
	}
	
	// Filter search
	if search != "" {
		query += " AND (c.tracking_code LIKE ? OR c.description LIKE ?)"
		countQuery += " AND (tracking_code LIKE ? OR description LIKE ?)"
		searchTerm := "%" + search + "%"
		args = append(args, searchTerm, searchTerm)
		countArgs = append(countArgs, searchTerm, searchTerm)
	}
	
	// Get total
	var total int
	err = db.DB.QueryRow(countQuery, countArgs...).Scan(&total)
	if err != nil {
		c.JSON(500, gin.H{"success": false, "message": err.Error()})
		return
	}
	
	// Add pagination
	query += " ORDER BY c.created_at DESC LIMIT ? OFFSET ?"
	args = append(args, limit, offset)
	
	rows, err := db.DB.Query(query, args...)
	if err != nil {
		c.JSON(500, gin.H{"success": false, "message": err.Error()})
		return
	}
	defer rows.Close()
	
	var complaints []map[string]interface{}
	for rows.Next() {
		var c Complaint
		err := rows.Scan(
			&c.ID, &c.TrackingCode, &c.Description, &c.LocationDetail,
			&c.Status, &c.CreatedAt, &c.UserName, &c.UserFullname,
			&c.CategoryName, &c.ProvinceName,
		)
		if err != nil {
			continue
		}
		
		c.StatusText = getStatusText(c.Status)
		
		complaints = append(complaints, map[string]interface{}{
			"id":              c.ID,
			"tracking_code":   c.TrackingCode,
			"description":     c.Description,
			"location_detail": c.LocationDetail,
			"status":          c.Status,
			"status_text":     c.StatusText,
			"created_at":      c.CreatedAt,
			"user_name":       c.UserName,
			"user_fullname":   c.UserFullname,
			"category_name":   c.CategoryName,
			"province_name":   c.ProvinceName,
		})
	}
	
	c.JSON(200, gin.H{
		"success": true,
		"data": complaints,
		"total": total,
		"page": page,
		"limit": limit,
	})
}

// GetProcessReports - ambil semua laporan proses untuk admin
func (h *ComplaintHandler) GetProcessReports(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	
	offset := (page - 1) * limit
	
	// Query untuk mengambil laporan proses
	query := `
		SELECT 
			pr.id,
			pr.complaint_id,
			pr.governor_id,
			pr.process_photos,
			pr.process_notes,
			pr.process_date,
			pr.status,
			pr.admin_notes,
			pr.submitted_at,
			c.tracking_code,
			c.description,
			c.location_detail,
			u.username as reporter_name,
			u.fullname as reporter_fullname,
			g.fullname as governor_name,
			p.name as province_name
		FROM process_reports pr
		JOIN complaints c ON pr.complaint_id = c.id
		JOIN users u ON c.user_id = u.id
		JOIN users g ON pr.governor_id = g.id
		JOIN provinces p ON c.province_api_id = p.api_id
		ORDER BY pr.submitted_at DESC
		LIMIT ? OFFSET ?
	`
	
	rows, err := db.DB.Query(query, limit, offset)
	if err != nil {
		c.JSON(500, gin.H{
			"success": false, 
			"message": "Database error: " + err.Error(),
		})
		return
	}
	defer rows.Close()
	
	var reports []map[string]interface{}
	for rows.Next() {
		var (
			id, complaintID, governorID int
			processPhotos, processNotes, processDate, status, adminNotes sql.NullString
			submittedAt time.Time
			trackingCode, description, locationDetail, reporterName, reporterFullname, governorName, provinceName string
		)
		
		err := rows.Scan(
			&id, &complaintID, &governorID,
			&processPhotos, &processNotes, &processDate,
			&status, &adminNotes, &submittedAt,
			&trackingCode, &description, &locationDetail,
			&reporterName, &reporterFullname, &governorName, &provinceName,
		)
		if err != nil {
			continue
		}
		
		report := map[string]interface{}{
			"id":           id,
			"complaint_id": complaintID,
			"governor_id":  governorID,
			"governor_name": governorName,
			"process_photos": nil,
			"process_notes":  nil,
			"process_date":   nil,
			"status":        "pending",
			"admin_notes":   nil,
			"submitted_at":  submittedAt,
			"complaint": map[string]interface{}{
				"tracking_code":   trackingCode,
				"description":     description,
				"location_detail": locationDetail,
				"user_name":       reporterName,
				"user_fullname":   reporterFullname,
				"province_name":   provinceName,
			},
		}
		
		if processPhotos.Valid {
			report["process_photos"] = processPhotos.String
		}
		if processNotes.Valid {
			report["process_notes"] = processNotes.String
		}
		if processDate.Valid {
			report["process_date"] = processDate.String
		}
		if status.Valid {
			report["status"] = status.String
		}
		if adminNotes.Valid {
			report["admin_notes"] = adminNotes.String
		}
		
		reports = append(reports, report)
	}
	
	// Hitung total
	var total int
	db.DB.QueryRow("SELECT COUNT(*) FROM process_reports").Scan(&total)
	
	c.JSON(200, gin.H{
		"success": true,
		"data":    reports,
		"total":   total,
		"page":    page,
		"limit":   limit,
	})
}

// GetCompletionReports - ambil semua laporan akhir untuk admin
func (h *ComplaintHandler) GetCompletionReports(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	
	offset := (page - 1) * limit
	
	// Query untuk mengambil laporan akhir
	query := `
		SELECT 
			cr.id,
			cr.complaint_id,
			cr.governor_id,
			cr.final_photos,
			cr.completion_date,
			cr.cost,
			cr.cost_details,
			cr.work_details,
			cr.status,
			cr.admin_notes,
			cr.submitted_at,
			c.tracking_code,
			c.description,
			c.location_detail,
			u.username as reporter_name,
			u.fullname as reporter_fullname,
			g.fullname as governor_name,
			p.name as province_name
		FROM completion_reports cr
		JOIN complaints c ON cr.complaint_id = c.id
		JOIN users u ON c.user_id = u.id
		JOIN users g ON cr.governor_id = g.id
		JOIN provinces p ON c.province_api_id = p.api_id
		ORDER BY cr.submitted_at DESC
		LIMIT ? OFFSET ?
	`
	
	rows, err := db.DB.Query(query, limit, offset)
	if err != nil {
		c.JSON(500, gin.H{
			"success": false, 
			"message": "Database error: " + err.Error(),
		})
		return
	}
	defer rows.Close()
	
	var reports []map[string]interface{}
	for rows.Next() {
		var (
			id, complaintID, governorID int
			finalPhotos, completionDate, costDetails, workDetails, status, adminNotes sql.NullString
			cost sql.NullFloat64
			submittedAt time.Time
			trackingCode, description, locationDetail, reporterName, reporterFullname, governorName, provinceName string
		)
		
		err := rows.Scan(
			&id, &complaintID, &governorID,
			&finalPhotos, &completionDate, &cost, &costDetails, &workDetails,
			&status, &adminNotes, &submittedAt,
			&trackingCode, &description, &locationDetail,
			&reporterName, &reporterFullname, &governorName, &provinceName,
		)
		if err != nil {
			continue
		}
		
		report := map[string]interface{}{
			"id":           id,
			"complaint_id": complaintID,
			"governor_id":  governorID,
			"governor_name": governorName,
			"final_photos":  nil,
			"completion_date": nil,
			"cost":          nil,
			"cost_details":  nil,
			"work_details":  nil,
			"status":        "pending",
			"admin_notes":   nil,
			"submitted_at":  submittedAt,
			"complaint": map[string]interface{}{
				"tracking_code":   trackingCode,
				"description":     description,
				"location_detail": locationDetail,
				"user_name":       reporterName,
				"user_fullname":   reporterFullname,
				"province_name":   provinceName,
			},
		}
		
		if finalPhotos.Valid {
			report["final_photos"] = finalPhotos.String
		}
		if completionDate.Valid {
			report["completion_date"] = completionDate.String
		}
		if cost.Valid {
			report["cost"] = cost.Float64
		}
		if costDetails.Valid {
			report["cost_details"] = costDetails.String
		}
		if workDetails.Valid {
			report["work_details"] = workDetails.String
		}
		if status.Valid {
			report["status"] = status.String
		}
		if adminNotes.Valid {
			report["admin_notes"] = adminNotes.String
		}
		
		reports = append(reports, report)
	}
	
	// Hitung total
	var total int
	db.DB.QueryRow("SELECT COUNT(*) FROM completion_reports").Scan(&total)
	
	c.JSON(200, gin.H{
		"success": true,
		"data":    reports,
		"total":   total,
		"page":    page,
		"limit":   limit,
	})
}
// GetProvincesWithUsers - ambil provinsi yang memiliki user dengan role tertentu
func (h *ComplaintHandler) GetProvincesWithUsers(c *gin.Context) {
	role := c.Query("role")
	if role == "" {
		c.JSON(400, gin.H{"success": false, "message": "role required"})
		return
	}
	
	query := `
		SELECT p.api_id, p.name, COUNT(u.id) as total_users
		FROM provinces p
		JOIN users u ON u.province_api_id = p.api_id
		WHERE u.role = ?
		GROUP BY p.api_id, p.name
		ORDER BY p.name ASC
	`
	
	rows, err := db.DB.Query(query, role)
	if err != nil {
		c.JSON(500, gin.H{"success": false, "message": err.Error()})
		return
	}
	defer rows.Close()
	
	var provinces []map[string]interface{}
	for rows.Next() {
		var apiID int
		var name string
		var totalUsers int
		rows.Scan(&apiID, &name, &totalUsers)
		provinces = append(provinces, map[string]interface{}{
			"api_id":      apiID,
			"name":        name,
			"total_users": totalUsers,
		})
	}
	
	c.JSON(200, gin.H{"success": true, "data": provinces})
}

// GetCategoriesForAdmin - ambil semua kategori untuk admin (dengan pagination)
func (h *ComplaintHandler) GetCategoriesForAdmin(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	search := c.Query("search")
	
	offset := (page - 1) * limit
	
	query := "FROM categories WHERE 1=1"
	args := []interface{}{}
	
	if search != "" {
		query += " AND name LIKE ?"
		args = append(args, "%"+search+"%")
	}
	
	var total int
	err := db.DB.QueryRow("SELECT COUNT(*) "+query, args...).Scan(&total)
	if err != nil {
		c.JSON(500, gin.H{"success": false, "message": err.Error()})
		return
	}
	
	rows, err := db.DB.Query(`
		SELECT id, name, COALESCE(description, ''), COALESCE(icon, ''), sort_order, is_active, created_at
		FROM categories
		WHERE 1=1 `+func() string {
		if search != "" {
			return "AND name LIKE ?"
		}
		return ""
	}()+` ORDER BY sort_order ASC, name ASC LIMIT ? OFFSET ?`,
		append(args, limit, offset)...,
	)
	if err != nil {
		c.JSON(500, gin.H{"success": false, "message": err.Error()})
		return
	}
	defer rows.Close()
	
	var categories []map[string]interface{}
	for rows.Next() {
		var id, sortOrder int
		var name, description, icon string
		var isActive bool
		var createdAt time.Time
		
		rows.Scan(&id, &name, &description, &icon, &sortOrder, &isActive, &createdAt)
		categories = append(categories, map[string]interface{}{
			"id":          id,
			"name":        name,
			"description": description,
			"icon":        icon,
			"sort_order":  sortOrder,
			"is_active":   isActive,
			"created_at":  createdAt,
		})
	}
	
	c.JSON(200, gin.H{
		"success": true,
		"data":    categories,
		"total":   total,
		"page":    page,
		"limit":   limit,
	})
}

// CreateCategory - tambah kategori baru
func (h *ComplaintHandler) CreateCategory(c *gin.Context) {
	var req struct {
		Name        string `json:"name" binding:"required"`
		Description string `json:"description"`
		Icon        string `json:"icon"`
		SortOrder   int    `json:"sort_order"`
		IsActive    bool   `json:"is_active"`
	}
	
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"success": false, "message": err.Error()})
		return
	}
	
	result, err := db.DB.Exec(`
		INSERT INTO categories (name, description, icon, sort_order, is_active)
		VALUES (?, ?, ?, ?, ?)
	`, req.Name, req.Description, req.Icon, req.SortOrder, req.IsActive)
	
	if err != nil {
		c.JSON(500, gin.H{"success": false, "message": err.Error()})
		return
	}
	
	id, _ := result.LastInsertId()
	
	c.JSON(200, gin.H{
		"success": true,
		"message": "Kategori berhasil ditambahkan",
		"data":    map[string]interface{}{"id": id},
	})
}

// UpdateCategory - update kategori
func (h *ComplaintHandler) UpdateCategory(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(400, gin.H{"success": false, "message": "ID tidak valid"})
		return
	}
	
	var req struct {
		Name        string `json:"name" binding:"required"`
		Description string `json:"description"`
		Icon        string `json:"icon"`
		SortOrder   int    `json:"sort_order"`
		IsActive    bool   `json:"is_active"`
	}
	
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"success": false, "message": err.Error()})
		return
	}
	
	_, err = db.DB.Exec(`
		UPDATE categories 
		SET name = ?, description = ?, icon = ?, sort_order = ?, is_active = ?, updated_at = NOW()
		WHERE id = ?
	`, req.Name, req.Description, req.Icon, req.SortOrder, req.IsActive, id)
	
	if err != nil {
		c.JSON(500, gin.H{"success": false, "message": err.Error()})
		return
	}
	
	c.JSON(200, gin.H{"success": true, "message": "Kategori berhasil diupdate"})
}

// DeleteCategory - hapus kategori
func (h *ComplaintHandler) DeleteCategory(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(400, gin.H{"success": false, "message": "ID tidak valid"})
		return
	}
	
	_, err = db.DB.Exec("DELETE FROM categories WHERE id = ?", id)
	if err != nil {
		c.JSON(500, gin.H{"success": false, "message": err.Error()})
		return
	}
	
	c.JSON(200, gin.H{"success": true, "message": "Kategori berhasil dihapus"})
}

// ToggleCategoryActive - aktif/nonaktif kategori
func (h *ComplaintHandler) ToggleCategoryActive(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(400, gin.H{"success": false, "message": "ID tidak valid"})
		return
	}
	
	var req struct {
		IsActive bool `json:"is_active"`
	}
	
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"success": false, "message": err.Error()})
		return
	}
	
	_, err = db.DB.Exec(`
		UPDATE categories SET is_active = ?, updated_at = NOW() WHERE id = ?
	`, req.IsActive, id)
	
	if err != nil {
		c.JSON(500, gin.H{"success": false, "message": err.Error()})
		return
	}
	
	c.JSON(200, gin.H{"success": true, "message": "Status kategori berhasil diubah"})
}


// GetReportStats - ambil statistik laporan
func (h *ComplaintHandler) GetReportStats(c *gin.Context) {
	period := c.DefaultQuery("period", "month")
	startDate := c.Query("start_date")
	endDate := c.Query("end_date")
	
	var stats ReportStats
	
	// Base query for date filter
	dateFilter := ""
	switch period {
	case "week":
		dateFilter = "AND created_at >= DATE_SUB(NOW(), INTERVAL 1 WEEK)"
	case "month":
		dateFilter = "AND created_at >= DATE_SUB(NOW(), INTERVAL 1 MONTH)"
	case "year":
		dateFilter = "AND created_at >= DATE_SUB(NOW(), INTERVAL 1 YEAR)"
	case "custom":
		if startDate != "" && endDate != "" {
			dateFilter = fmt.Sprintf("AND DATE(created_at) BETWEEN '%s' AND '%s'", startDate, endDate)
		}
	}
	
	// Total complaints
	db.DB.QueryRow("SELECT COUNT(*) FROM complaints WHERE 1=1 "+dateFilter).Scan(&stats.TotalComplaints)
	
	// Status counts
	db.DB.QueryRow("SELECT COUNT(*) FROM complaints WHERE status = 'pending_governor' "+dateFilter).Scan(&stats.PendingGovernor)
	db.DB.QueryRow("SELECT COUNT(*) FROM complaints WHERE status = 'investigation_assigned' "+dateFilter).Scan(&stats.InvestigationAssigned)
	db.DB.QueryRow("SELECT COUNT(*) FROM complaints WHERE status = 'investigation_done' "+dateFilter).Scan(&stats.InvestigationDone)
	db.DB.QueryRow("SELECT COUNT(*) FROM complaints WHERE status = 'governor_processing' "+dateFilter).Scan(&stats.GovernorProcessing)
	db.DB.QueryRow("SELECT COUNT(*) FROM complaints WHERE status = 'completed' "+dateFilter).Scan(&stats.Completed)
	db.DB.QueryRow("SELECT COUNT(*) FROM complaints WHERE status = 'rejected' "+dateFilter).Scan(&stats.Rejected)
	
	// This month & week (tanpa filter date)
	db.DB.QueryRow("SELECT COUNT(*) FROM complaints WHERE MONTH(created_at) = MONTH(CURDATE()) AND YEAR(created_at) = YEAR(CURDATE())").Scan(&stats.ThisMonth)
	db.DB.QueryRow("SELECT COUNT(*) FROM complaints WHERE YEARWEEK(created_at) = YEARWEEK(CURDATE())").Scan(&stats.ThisWeek)
	
	// Avg completion days
	db.DB.QueryRow(`
		SELECT COALESCE(AVG(DATEDIFF(updated_at, created_at)), 0) 
		FROM complaints WHERE status = 'completed'`+dateFilter).Scan(&stats.AvgCompletionDays)
	
	// By category
	rows, err := db.DB.Query(`
		SELECT c.name, COUNT(co.id) as count
		FROM complaints co
		JOIN categories c ON co.category_id = c.id
		WHERE 1=1 `+dateFilter+`
		GROUP BY co.category_id, c.name
		ORDER BY count DESC
	`)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var item CategoryCount
			rows.Scan(&item.Name, &item.Count)
			stats.ByCategory = append(stats.ByCategory, item)
		}
	}
	
	// By status for pie chart
	statusRows, err := db.DB.Query(`
		SELECT 
			CASE status
				WHEN 'pending_governor' THEN 'Menunggu'
				WHEN 'investigation_assigned' THEN 'Investigasi'
				WHEN 'investigation_done' THEN 'Investigasi Selesai'
				WHEN 'governor_processing' THEN 'Diproses'
				WHEN 'completed' THEN 'Selesai'
				WHEN 'rejected' THEN 'Ditolak'
				ELSE status
			END as name,
			COUNT(*) as value
		FROM complaints
		WHERE 1=1 `+dateFilter+`
		GROUP BY status
	`)
	if err == nil {
		defer statusRows.Close()
		for statusRows.Next() {
			var item StatusCount
			statusRows.Scan(&item.Name, &item.Value)
			stats.ByStatus = append(stats.ByStatus, item)
		}
	}
	
	// By month
	monthRows, err := db.DB.Query(`
		SELECT 
			DATE_FORMAT(created_at, '%b') as month,
			COUNT(*) as total
		FROM complaints
		WHERE 1=1 `+dateFilter+`
		GROUP BY YEAR(created_at), MONTH(created_at)
		ORDER BY MIN(created_at) ASC
		LIMIT 12
	`)
	if err == nil {
		defer monthRows.Close()
		for monthRows.Next() {
			var month string
			var total int
			monthRows.Scan(&month, &total)
			stats.ByMonth = append(stats.ByMonth, MonthCount{Month: month, Count: total})
		}
	}
	
	// By province
	provRows, err := db.DB.Query(`
		SELECT p.name, COUNT(c.id) as total
		FROM complaints c
		JOIN provinces p ON c.province_api_id = p.api_id
		WHERE 1=1 `+dateFilter+`
		GROUP BY c.province_api_id, p.name
		ORDER BY total DESC
		LIMIT 10
	`)
	if err == nil {
		defer provRows.Close()
		for provRows.Next() {
			var name string
			var total int
			provRows.Scan(&name, &total)
			stats.ByProvince = append(stats.ByProvince, ProvinceCount{Name: name, Total: total})
		}
	}
	
	c.JSON(200, gin.H{"success": true, "data": stats})
}

// ExportReport - export laporan ke Excel/PDF
func (h *ComplaintHandler) ExportReport(c *gin.Context) {
	// format := c.DefaultQuery("format", "excel")
	// period := c.DefaultQuery("period", "month")
	
	// TODO: Implement export logic here
	// Untuk sementara, return response dulu
	
	c.JSON(200, gin.H{
		"success": true, 
		"message": "Fitur export sedang dalam pengembangan",
		"data": map[string]interface{}{
			"format": c.DefaultQuery("format", "excel"),
			"period": c.DefaultQuery("period", "month"),
		},
	})
}

// GetActivityLogs - ambil log aktivitas
func (h *ComplaintHandler) GetActivityLogs(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	search := c.Query("search")
	action := c.Query("action")
	userID := c.Query("user_id")
	
	offset := (page - 1) * limit
	
	query := `
		SELECT 
			a.id,
			a.user_id,
			u.username,
			u.fullname,
			u.role,
			a.action,
			a.complaint_id,
			c.tracking_code,
			a.old_status,
			a.new_status,
			a.ip_address,
			a.user_agent,
			a.created_at
		FROM activity_logs a
		LEFT JOIN users u ON a.user_id = u.id
		LEFT JOIN complaints c ON a.complaint_id = c.id
		WHERE 1=1
	`
	countQuery := "SELECT COUNT(*) FROM activity_logs WHERE 1=1"
	args := []interface{}{}
	
	if search != "" {
		query += " AND (u.username LIKE ? OR u.fullname LIKE ? OR c.tracking_code LIKE ?)"
		countQuery += " AND (user_id IN (SELECT id FROM users WHERE username LIKE ? OR fullname LIKE ?) OR complaint_id IN (SELECT id FROM complaints WHERE tracking_code LIKE ?))"
		searchTerm := "%" + search + "%"
		args = append(args, searchTerm, searchTerm, searchTerm)
	}
	
	if action != "" && action != "all" {
		query += " AND a.action = ?"
		countQuery += " AND action = ?"
		args = append(args, action)
	}
	
	if userID != "" && userID != "all" {
		query += " AND a.user_id = ?"
		countQuery += " AND user_id = ?"
		args = append(args, userID)
	}
	
	var total int
	err := db.DB.QueryRow(countQuery, args...).Scan(&total)
	if err != nil {
		c.JSON(500, gin.H{"success": false, "message": err.Error()})
		return
	}
	
	query += " ORDER BY a.created_at DESC LIMIT ? OFFSET ?"
	args = append(args, limit, offset)
	
	rows, err := db.DB.Query(query, args...)
	if err != nil {
		c.JSON(500, gin.H{"success": false, "message": err.Error()})
		return
	}
	defer rows.Close()
	
	var activities []map[string]interface{}
	for rows.Next() {
		var id, userID int
		var username, fullname, role, action, ipAddress, userAgent, trackingCode sql.NullString
		var oldStatus, newStatus sql.NullString
		var complaintID sql.NullInt64
		var createdAt time.Time
		
		rows.Scan(&id, &userID, &username, &fullname, &role, &action,
			&complaintID, &trackingCode, &oldStatus, &newStatus, &ipAddress, &userAgent, &createdAt)
		
		activity := map[string]interface{}{
			"id":            id,
			"user_id":       userID,
			"user_name":     username.String,
			"user_fullname": fullname.String,
			"user_role":     role.String,
			"action":        action.String,
			"complaint_id":  nil,
			"tracking_code": nil,
			"old_status":    nil,
			"new_status":    nil,
			"ip_address":    ipAddress.String,
			"user_agent":    userAgent.String,
			"created_at":    createdAt,
		}
		
		if complaintID.Valid {
			activity["complaint_id"] = int(complaintID.Int64)
		}
		if trackingCode.Valid {
			activity["tracking_code"] = trackingCode.String
		}
		if oldStatus.Valid {
			activity["old_status"] = oldStatus.String
		}
		if newStatus.Valid {
			activity["new_status"] = newStatus.String
		}
		
		activities = append(activities, activity)
	}
	
	c.JSON(200, gin.H{
		"success": true,
		"data":    activities,
		"total":   total,
		"page":    page,
		"limit":   limit,
	})
}

// GetActivityStats - ambil statistik log aktivitas
func (h *ComplaintHandler) GetActivityStats(c *gin.Context) {
	var stats struct {
		Total     int `json:"total"`
		Today     int `json:"today"`
		ThisWeek  int `json:"this_week"`
		ThisMonth int `json:"this_month"`
	}
	
	db.DB.QueryRow("SELECT COUNT(*) FROM activity_logs").Scan(&stats.Total)
	db.DB.QueryRow("SELECT COUNT(*) FROM activity_logs WHERE DATE(created_at) = CURDATE()").Scan(&stats.Today)
	db.DB.QueryRow("SELECT COUNT(*) FROM activity_logs WHERE YEARWEEK(created_at) = YEARWEEK(CURDATE())").Scan(&stats.ThisWeek)
	db.DB.QueryRow("SELECT COUNT(*) FROM activity_logs WHERE MONTH(created_at) = MONTH(CURDATE()) AND YEAR(created_at) = YEAR(CURDATE())").Scan(&stats.ThisMonth)
	
	c.JSON(200, gin.H{"success": true, "data": stats})
}