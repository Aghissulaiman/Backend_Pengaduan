package complaint

import (
    "fmt"
    "net/http"
    "strconv"
    "pengaduan_be2/internal/dto"
    "github.com/gin-gonic/gin"
)

type ComplaintHandler struct {
    service *ComplaintService
}

func NewComplaintHandler() *ComplaintHandler {
    return &ComplaintHandler{service: NewComplaintService()}
}

// GetCategories godoc
// @Summary Get all complaint categories
// @Tags Complaint
// @Produce json
// @Success 200 {object} dto.Response
// @Router /api/complaints/categories [get]
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

// SubmitComplaint godoc
// @Summary Submit a new complaint
// @Tags Complaint
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body SubmitComplaintRequest true "Complaint data"
// @Success 201 {object} dto.Response
// @Failure 400 {object} dto.Response
// @Router /api/complaints/submit [post]
func (h *ComplaintHandler) SubmitComplaint(c *gin.Context) {
    var req SubmitComplaintRequest
    
    // Log request body
    body, _ := c.GetRawData()
    fmt.Printf("Raw request body: %s\n", string(body))
    
    if err := c.ShouldBindJSON(&req); err != nil {
        fmt.Printf("Binding error: %v\n", err)
        c.JSON(http.StatusBadRequest, dto.Response{
            Success: false,
            Message: err.Error(),
        })
        return
    }

    // Log request yang diterima
    fmt.Printf("Received request: %+v\n", req)

    userIDRaw, exists := c.Get("user_id")
    if !exists {
        fmt.Printf("User ID not found in context\n")
        c.JSON(http.StatusUnauthorized, dto.Response{
            Success: false,
            Message: "User tidak terautentikasi",
        })
        return
    }
    
    userID, ok := userIDRaw.(int)
    if !ok {
        fmt.Printf("User ID invalid type: %T\n", userIDRaw)
        c.JSON(http.StatusUnauthorized, dto.Response{
            Success: false,
            Message: "User ID tidak valid",
        })
        return
    }
    
    fmt.Printf("UserID from token: %d\n", userID)

    resp, err := h.service.SubmitComplaint(userID, &req)
    if err != nil {
        fmt.Printf("Service error: %v\n", err)
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

// GetMyComplaints godoc
// @Summary Get my complaints
// @Tags Complaint
// @Security BearerAuth
// @Produce json
// @Param page query int false "Page number"
// @Param limit query int false "Items per page"
// @Success 200 {object} dto.Response
// @Router /api/complaints/my [get]
func (h *ComplaintHandler) GetMyComplaints(c *gin.Context) {
    userIDRaw, exists := c.Get("user_id")
    if !exists {
        c.JSON(http.StatusUnauthorized, dto.Response{
            Success: false,
            Message: "User tidak terautentikasi",
        })
        return
    }
    
    userID, ok := userIDRaw.(int)
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
        fmt.Printf("GetUserComplaints error: %v\n", err)
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

// CheckStatus godoc
// @Summary Check complaint status by tracking code
// @Tags Complaint
// @Produce json
// @Param tracking_code path string true "Tracking code"
// @Success 200 {object} dto.Response
// @Failure 404 {object} dto.Response
// @Router /api/complaints/status/{tracking_code} [get]
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

// ============ GOVERNOR ENDPOINTS ============

// GetAllComplaints godoc
// @Summary Get all complaints (governor/admin/investigator)
// @Tags Complaint
// @Security BearerAuth
// @Produce json
// @Param status query string false "Filter by status"
// @Param province_api_id query int false "Filter by province (admin only)"
// @Param page query int false "Page number"
// @Param limit query int false "Items per page"
// @Success 200 {object} dto.Response
// @Router /api/complaints/all [get]
func (h *ComplaintHandler) GetAllComplaints(c *gin.Context) {
    role, _ := c.Get("role")
    userID, _ := c.Get("user_id")
    provinceID, _ := c.Get("province_api_id")

    var query GetComplaintsQuery
    query.Status = c.Query("status")
    if provinceApiID := c.Query("province_api_id"); provinceApiID != "" {
        query.ProvinceID, _ = strconv.Atoi(provinceApiID)
    }
    query.Page, _ = strconv.Atoi(c.DefaultQuery("page", "1"))
    query.Limit, _ = strconv.Atoi(c.DefaultQuery("limit", "10"))

    complaints, total, err := h.service.GetAllComplaints(
        role.(string), userID.(int), provinceID.(int), &query,
    )
    if err != nil {
        fmt.Printf("GetAllComplaints error: %v\n", err)
        c.JSON(http.StatusInternalServerError, dto.Response{
            Success: false,
            Message: "Gagal mengambil data",
        })
        return
    }

    c.JSON(http.StatusOK, gin.H{
        "success":    true,
        "data":       complaints,
        "pagination": dto.NewPagination(query.Page, query.Limit, total),
    })
}

// GetComplaintDetail godoc
// @Summary Get complaint detail
// @Tags Complaint
// @Security BearerAuth
// @Produce json
// @Param id path int true "Complaint ID"
// @Success 200 {object} dto.Response
// @Router /api/complaints/{id} [get]
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

// AssignInvestigator godoc
// @Summary Assign investigator to complaint (governor only)
// @Tags Complaint
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path int true "Complaint ID"
// @Param request body AssignInvestigatorRequest true "Investigator ID"
// @Success 200 {object} dto.Response
// @Router /api/governor/complaints/{id}/assign [post]
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
        Message: "Investigasi berhasil ditugaskan",
    })
}

// SubmitInvestigationResult godoc
// @Summary Submit investigation result (investigator only)
// @Tags Complaint
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path int true "Complaint ID"
// @Param request body InvestigationResultRequest true "Investigation result"
// @Success 200 {object} dto.Response
// @Router /api/investigator/complaints/{id}/result [post]
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
    if err := h.service.SubmitInvestigationResult(id, req.Result, *req.Evidence, req.IsValid, investigatorID.(int)); err != nil {
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

// SubmitProcessReport godoc
// @Summary Submit process report to admin (governor only)
// @Tags Complaint
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path int true "Complaint ID"
// @Param request body ProcessReportRequest true "Process report"
// @Success 200 {object} dto.Response
// @Router /api/governor/complaints/{id}/process-report [post]
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

// SubmitCompletionReport godoc
// @Summary Submit completion report to admin (governor only)
// @Tags Complaint
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path int true "Complaint ID"
// @Param request body CompletionReportRequest true "Completion report"
// @Success 200 {object} dto.Response
// @Router /api/governor/complaints/{id}/completion-report [post]
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

// GetDashboardStats godoc
// @Summary Get dashboard statistics (admin only)
// @Tags Admin
// @Security BearerAuth
// @Produce json
// @Success 200 {object} dto.Response
// @Router /api/admin/dashboard [get]
func (h *ComplaintHandler) GetDashboardStats(c *gin.Context) {
    stats, err := h.service.GetDashboardStats()
    if err != nil {
        fmt.Printf("GetDashboardStats error: %v\n", err)
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