package complaint

import "time" 

type SubmitComplaintRequest struct {
	ProvinceID     int     `json:"province_api_id" binding:"required"`
	RegencyID      *int    `json:"regency_id"`
	DistrictID     *int    `json:"district_id"`
	VillageID      *int    `json:"village_id"`
	LocationDetail string  `json:"location_detail" binding:"required"`
	CategoryID     int     `json:"category_id" binding:"required"`
	Description    string  `json:"description" binding:"required,min=10"`
	Photo          *string `json:"photo"`
}

type UpdateStatusRequest struct {
	Status       string  `json:"status" binding:"required,oneof=investigation_assigned rejected"`
	RejectReason *string `json:"reject_reason"`
}

type AssignInvestigatorRequest struct {
	InvestigatorID int `json:"investigator_id" binding:"required"`
}

type InvestigationResultRequest struct {
	Result   string  `json:"result" binding:"required"`
	Evidence *string `json:"evidence"`
	IsValid  bool    `json:"is_valid"`
}

type ProcessReportRequest struct {
	ProcessPhotos *string `json:"process_photos"`
	ProcessNotes  *string `json:"process_notes"`
	ProcessDate   *string `json:"process_date"`
}

type ProcessReportVerifyRequest struct {
	Status    string  `json:"status" binding:"required,oneof=verified rejected"`
	AdminNote *string `json:"admin_note"`
}

type CompletionReportRequest struct {
	FinalPhotos    *string  `json:"final_photos"`
	CompletionDate string   `json:"completion_date" binding:"required"`
	Cost           *float64 `json:"cost"`
	CostDetails    *string  `json:"cost_details"`
	WorkDetails    *string  `json:"work_details"`
}

type CompletionReportVerifyRequest struct {
	Status    string  `json:"status" binding:"required,oneof=verified rejected"`
	AdminNote *string `json:"admin_note"`
}

type PublicationRequest struct {
	Title   string  `json:"title" binding:"required"`
	Summary *string `json:"summary"`
}

type GetComplaintsQuery struct {
	Status     string `form:"status"`
	ProvinceID int    `form:"province_api_id"`
	Search     string `form:"search"`
	Page       int    `form:"page,default=1"`
	Limit      int    `form:"limit,default=10"`
}

type SubmitComplaintResponse struct {
	TrackingCode string `json:"tracking_code"`
	ID           int    `json:"id"`
}

type GovernorComplaintsQuery struct {
	Status string `form:"status"`
	Search string `form:"search"`
	Page   int    `form:"page,default=1"`
	Limit  int    `form:"limit,default=10"`
}

type InvestigatorComplaintsQuery struct {
	Status string `form:"status"`
	Search string `form:"search"`
	Page   int    `form:"page,default=1"`
	Limit  int    `form:"limit,default=10"`
}

// GetInvestigationsQuery - query untuk investigasi
type GetInvestigationsQuery struct {
	Status string `form:"status"`
	Search string `form:"search"`
	Page   int    `form:"page,default=1"`
	Limit  int    `form:"limit,default=10"`
}

// InvestigationResponse - response untuk investigasi
type InvestigationResponse struct {
	ID               int        `json:"id"`
	ComplaintID      int        `json:"complaint_id"`
	TrackingCode     string     `json:"tracking_code"`
	Description      string     `json:"description"`
	LocationDetail   string     `json:"location_detail"`
	Status           string     `json:"status"`
	StatusText       string     `json:"status_text"`
	CreatedAt        time.Time  `json:"created_at"`
	UserName         string     `json:"user_name"`
	UserFullname     string     `json:"user_fullname"`
	InvestigatorName *string    `json:"investigator_name"`
	InvestigatorID   *int       `json:"investigator_id"`
	AssignedAt       *time.Time `json:"assigned_at"`
	Notes            *string    `json:"notes"`
}

// ReportQuery - query untuk laporan
type ReportQuery struct {
	Period     string `form:"period"`
	StartDate  string `form:"start_date"`
	EndDate    string `form:"end_date"`
	Status     string `form:"status"`
	Search     string `form:"search"`
	Page       int    `form:"page,default=1"`
	Limit      int    `form:"limit,default=20"`
}

// ReportStatsResponse - response statistik laporan
type ReportStatsResponse struct {
	Total            int            `json:"total"`
	Pending          int            `json:"pending"`
	Investigation    int            `json:"investigation"`
	Completed        int            `json:"completed"`
	Rejected         int            `json:"rejected"`
	ThisMonth        int            `json:"thisMonth"`
	ThisWeek         int            `json:"thisWeek"`
	AvgCompletionDays int           `json:"avgCompletionDays"`
	ByCategory       []CategoryCount `json:"byCategory"`
	ByStatus         []StatusCount   `json:"byStatus"`
	ByMonth          []MonthCount    `json:"byMonth"`
}

type CategoryCount struct {
	Name  string `json:"name"`
	Count int    `json:"count"`
}

type StatusCount struct {
	Name  string `json:"name"`
	Value int    `json:"value"`  // ← ganti Count menjadi Value
}

type MonthCount struct {
	Month string `json:"month"`
	Count int    `json:"count"`
}