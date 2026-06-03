package complaint

import (
	"bytes"
	"io"
	"net/http"
	"strconv"

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