package complaint

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
    Status    string  `json:"status" binding:"required"`
    Note      *string `json:"note"`
}

type AssignInvestigatorRequest struct {
    InvestigatorID int `json:"investigator_id" binding:"required"`
}

type InvestigationResultRequest struct {
    Result  string  `json:"result" binding:"required"`
    Evidence *string `json:"evidence"`
    IsValid bool    `json:"is_valid"`
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
    Page       int    `form:"page,default=1"`
    Limit      int    `form:"limit,default=10"`
}