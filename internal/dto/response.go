    package dto

    type Response struct {
        Success bool        `json:"success"`
        Message string      `json:"message"`
        Data    interface{} `json:"data,omitempty"`
    }

    type Pagination struct {
        Page       int `json:"page"`
        Limit      int `json:"limit"`
        Total      int `json:"total"`
        TotalPages int `json:"total_pages"`
    }

    func NewPagination(page, limit, total int) Pagination {
        totalPages := (total + limit - 1) / limit
        if totalPages < 1 {
            totalPages = 1
        }
        return Pagination{
            Page:       page,
            Limit:      limit,
            Total:      total,
            TotalPages: totalPages,
        }
    }