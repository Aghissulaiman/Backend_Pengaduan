package complaint

import "time"

type Complaint struct {
	ID                       int        `json:"id"`
	TrackingCode             string     `json:"tracking_code"`
	UserID                   int        `json:"user_id"`
	UserName                 string     `json:"user_name"`
	UserFullname             string     `json:"user_fullname"`
	ProvinceID               int        `json:"province_api_id"`
	ProvinceName             string     `json:"province_name"`
	RegencyID                *int       `json:"regency_id"`
	RegencyName              *string    `json:"regency_name"`
	DistrictID               *int       `json:"district_id"`
	DistrictName             *string    `json:"district_name"`
	VillageID                *int       `json:"village_id"`
	VillageName              *string    `json:"village_name"`
	LocationDetail           string     `json:"location_detail"`
	CategoryID               int        `json:"category_id"`
	CategoryName             string     `json:"category_name"`
	Description              string     `json:"description"`
	Photo                    *string    `json:"photo"`
	Status                   string     `json:"status"`
	StatusText               string     `json:"status_text"`
	RejectedReason           *string    `json:"rejected_reason"`
	AssignedInvestigatorID   *int       `json:"assigned_investigator_id"`
	AssignedInvestigatorName *string    `json:"assigned_investigator_name"`
	InvestigationResult      *string    `json:"investigation_result"`
	InvestigationEvidence    *string    `json:"investigation_evidence"`
	InvestigationCompletedAt *time.Time `json:"investigation_completed_at"`
	GovernorProcessedAt      *time.Time `json:"governor_processed_at"`
	CreatedAt                time.Time  `json:"created_at"`
	UpdatedAt                time.Time  `json:"updated_at"`
}

type Category struct {
	ID          int     `json:"id"`
	Name        string  `json:"name"`
	Description *string `json:"description"`
	Icon        *string `json:"icon"`
	SortOrder   int     `json:"sort_order"`
}

type ProcessReport struct {
	ID            int        `json:"id"`
	ComplaintID   int        `json:"complaint_id"`
	GovernorID    int        `json:"governor_id"`
	GovernorName  string     `json:"governor_name"`
	ProcessPhotos *string    `json:"process_photos"`
	ProcessNotes  *string    `json:"process_notes"`
	ProcessDate   *time.Time `json:"process_date"`
	Status        string     `json:"status"`
	AdminNotes    *string    `json:"admin_notes"`
	VerifiedBy    *int       `json:"verified_by"`
	VerifiedAt    *time.Time `json:"verified_at"`
	SubmittedAt   time.Time  `json:"submitted_at"`
}

type CompletionReport struct {
	ID             int        `json:"id"`
	ComplaintID    int        `json:"complaint_id"`
	GovernorID     int        `json:"governor_id"`
	GovernorName   string     `json:"governor_name"`
	FinalPhotos    *string    `json:"final_photos"`
	CompletionDate time.Time  `json:"completion_date"`
	Cost           *float64   `json:"cost"`
	CostDetails    *string    `json:"cost_details"`
	WorkDetails    *string    `json:"work_details"`
	Status         string     `json:"status"`
	AdminNotes     *string    `json:"admin_notes"`
	VerifiedBy     *int       `json:"verified_by"`
	VerifiedAt     *time.Time `json:"verified_at"`
	IsPublished    bool       `json:"is_published"`
	PublishedAt    *time.Time `json:"published_at"`
	SubmittedAt    time.Time  `json:"submitted_at"`
}

type Publication struct {
	ID             int        `json:"id"`
	ComplaintID    int        `json:"complaint_id"`
	Title          string     `json:"title"`
	Summary        *string    `json:"summary"`
	ProcessPhotos  *string    `json:"process_photos"`
	FinalPhotos    *string    `json:"final_photos"`
	CompletionDate *time.Time `json:"completion_date"`
	Cost           *float64   `json:"cost"`
	WorkDetails    *string    `json:"work_details"`
	PublishedBy    int        `json:"published_by"`
	PublishedAt    time.Time  `json:"published_at"`
	ViewCount      int        `json:"view_count"`
	LikeCount      int        `json:"like_count"`
}

type DashboardStats struct {
	TotalWarga        int `json:"total_warga"`
	TotalInvestigator int `json:"total_investigator"`
	TotalGovernor     int `json:"total_governor"`
	TotalComplaints   int `json:"total_complaints"`
	NeedAttention     int `json:"need_attention"`
	NeedVerification  int `json:"need_verification"`
	Completed         int `json:"completed"`
	TodayComplaints   int `json:"today_complaints"`
	TodayPublications int `json:"today_publications"`
}

type GovernorStatsResponse struct {
	Total                 int `json:"total"`
	PendingGovernor       int `json:"pending_governor"`
	InvestigationAssigned int `json:"investigation_assigned"`
	InvestigationDone     int `json:"investigation_done"`
	GovernorProcessing    int `json:"governor_processing"`
	Completed             int `json:"completed"`
	Rejected              int `json:"rejected"`
}

type Investigator struct {
	ID            int    `json:"id"`
	Username      string `json:"username"`
	Fullname      string `json:"fullname"`
	Email         string `json:"email"`
	ProvinceApiID int    `json:"province_api_id"`
}
// InvestigatorStatsResponse - statistik untuk dashboard investigator
type InvestigatorStatsResponse struct {
	Total                 int `json:"total"`
	InvestigationAssigned int `json:"investigation_assigned"`
	InvestigationDone     int `json:"investigation_done"`
	Completed             int `json:"completed"`
}

// ChartData untuk dashboard admin (grafik)
type ChartData struct {
	MonthlyComplaints    []MonthlyComplaint `json:"monthly_complaints"`
	ComplaintsByStatus   []StatusCount      `json:"complaints_by_status"`
	ComplaintsByCategory []CategoryCount    `json:"complaints_by_category"`
}

type MonthlyComplaint struct {
	Month string `json:"month"`
	Total int    `json:"total"`
}



// UserResponse untuk admin kelola user
type UserResponse struct {
	ID           int       `json:"id"`
	Username     string    `json:"username"`
	Fullname     string    `json:"fullname"`
	Email        string    `json:"email"`
	Role         string    `json:"role"`
	ProvinceApiID *int     `json:"province_api_id"`
	IsActive     bool      `json:"is_active"`
	CreatedAt    time.Time `json:"created_at"`
}

// RecentComplaint untuk dashboard (5 terbaru)
type RecentComplaint struct {
	ID           int       `json:"id"`
	TrackingCode string    `json:"tracking_code"`
	Description  string    `json:"description"`
	Status       string    `json:"status"`
	StatusText   string    `json:"status_text"`
	CreatedAt    time.Time `json:"created_at"`
	UserName     string    `json:"user_name"`
	UserFullname string    `json:"user_fullname"`
}

// RecentUser untuk dashboard (5 terbaru)
type RecentUser struct {
	ID        int       `json:"id"`
	Username  string    `json:"username"`
	Fullname  string    `json:"fullname"`
	Role      string    `json:"role"`
	CreatedAt time.Time `json:"created_at"`
}

// Tambahkan di akhir file complaint_entity.go

type ReportStats struct {
	TotalComplaints       int              `json:"total_complaints"`
	PendingGovernor       int              `json:"pending_governor"`
	InvestigationAssigned int              `json:"investigation_assigned"`
	InvestigationDone     int              `json:"investigation_done"`
	GovernorProcessing    int              `json:"governor_processing"`
	Completed             int              `json:"completed"`
	Rejected              int              `json:"rejected"`
	ThisMonth             int              `json:"this_month"`
	ThisWeek              int              `json:"this_week"`
	AvgCompletionDays     float64          `json:"avg_completion_days"`
	ByCategory            []CategoryCount  `json:"by_category"`
	ByStatus              []StatusCount    `json:"by_status"`
	ByMonth               []MonthCount     `json:"by_month"`
	ByProvince            []ProvinceCount  `json:"by_province"`
}



type ProvinceCount struct {
	Name  string `json:"name"`
	Total int    `json:"total"`
}

