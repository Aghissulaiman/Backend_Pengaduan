package complaint

import "time"

type Complaint struct {
    ID           int        `json:"id"`
    TrackingCode string     `json:"tracking_code"`
    UserID       int        `json:"user_id"`
    UserName     string     `json:"user_name"`
    ProvinceID   int        `json:"province_api_id"`
    ProvinceName string     `json:"province_name"`
    RegencyID    *int       `json:"regency_id"`
    RegencyName  *string    `json:"regency_name"`
    DistrictID   *int       `json:"district_id"`
    DistrictName *string    `json:"district_name"`
    VillageID    *int       `json:"village_id"`
    VillageName  *string    `json:"village_name"`
    LocationDetail string   `json:"location_detail"`
    CategoryID   int        `json:"category_id"`
    CategoryName string     `json:"category_name"`
    Description  string     `json:"description"`
    Photo        *string    `json:"photo"`
    Status       string     `json:"status"`
    StatusText   string     `json:"status_text"`
    RejectedReason *string  `json:"rejected_reason"`
    AssignedInvestigatorID *int `json:"assigned_investigator_id"`
    AssignedInvestigatorName *string `json:"assigned_investigator_name"`
    InvestigationResult *string `json:"investigation_result"`
    InvestigationEvidence *string `json:"investigation_evidence"`
    CreatedAt    time.Time  `json:"created_at"`
    UpdatedAt    time.Time  `json:"updated_at"`
}

type Category struct {
    ID          int     `json:"id"`
    Name        string  `json:"name"`
    Description *string `json:"description"`
    Icon        *string `json:"icon"`
    SortOrder   int     `json:"sort_order"`
}

type ProcessReport struct {
    ID           int        `json:"id"`
    ComplaintID  int        `json:"complaint_id"`
    GovernorID   int        `json:"governor_id"`
    GovernorName string     `json:"governor_name"`
    ProcessPhotos *string   `json:"process_photos"` // JSON array
    ProcessNotes *string    `json:"process_notes"`
    ProcessDate  *time.Time `json:"process_date"`
    Status       string     `json:"status"`
    AdminNotes   *string    `json:"admin_notes"`
    VerifiedBy   *int       `json:"verified_by"`
    VerifiedAt   *time.Time `json:"verified_at"`
    SubmittedAt  time.Time  `json:"submitted_at"`
}

type CompletionReport struct {
    ID            int        `json:"id"`
    ComplaintID   int        `json:"complaint_id"`
    GovernorID    int        `json:"governor_id"`
    GovernorName  string     `json:"governor_name"`
    FinalPhotos   *string    `json:"final_photos"`
    CompletionDate time.Time `json:"completion_date"`
    Cost          *float64   `json:"cost"`
    CostDetails   *string    `json:"cost_details"`
    WorkDetails   *string    `json:"work_details"`
    Status        string     `json:"status"`
    AdminNotes    *string    `json:"admin_notes"`
    VerifiedBy    *int       `json:"verified_by"`
    VerifiedAt    *time.Time `json:"verified_at"`
    IsPublished   bool       `json:"is_published"`
    PublishedAt   *time.Time `json:"published_at"`
    SubmittedAt   time.Time  `json:"submitted_at"`
}

type Publication struct {
    ID            int        `json:"id"`
    ComplaintID   int        `json:"complaint_id"`
    Title         string     `json:"title"`
    Summary       *string    `json:"summary"`
    ProcessPhotos *string    `json:"process_photos"`
    FinalPhotos   *string    `json:"final_photos"`
    CompletionDate *time.Time `json:"completion_date"`
    Cost          *float64   `json:"cost"`
    WorkDetails   *string    `json:"work_details"`
    PublishedBy   int        `json:"published_by"`
    PublishedAt   time.Time  `json:"published_at"`
    ViewCount     int        `json:"view_count"`
    LikeCount     int        `json:"like_count"`
}

type DashboardStats struct {
    TotalWarga            int `json:"total_warga"`
    TotalInvestigator     int `json:"total_investigator"`
    TotalGovernor         int `json:"total_governor"`
    TotalComplaints       int `json:"total_complaints"`
    NeedAttention         int `json:"need_attention"`
    NeedVerification      int `json:"need_verification"`
    Completed             int `json:"completed"`
    TodayComplaints       int `json:"today_complaints"`
    TodayPublications     int `json:"today_publications"`
}