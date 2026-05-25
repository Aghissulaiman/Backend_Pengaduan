package auth

import "time"

type User struct {
    ID            int        `json:"id"`
    Username      string     `json:"username"`
    Email         string     `json:"email"`
    Fullname      string     `json:"fullname"`
    Phone         *string    `json:"phone"`
    Avatar        *string    `json:"avatar"`
    ProvinceApiID *int       `json:"province_api_id"`
    RegencyApiID  *int       `json:"regency_api_id"`
    DistrictApiID *int       `json:"district_api_id"`
    VillageApiID  *int       `json:"village_api_id"`
    FullAddress   *string    `json:"full_address"`
    Role          string     `json:"role"`
    IsActive      bool       `json:"is_active"`
    IsVerified    bool       `json:"is_verified"`
    GoogleID      *string    `json:"-"`
    LastLogin     *time.Time `json:"last_login"`
    CreatedAt     time.Time  `json:"created_at"`
    UpdatedAt     time.Time  `json:"updated_at"`
}

type UserResponse struct {
    ID            int     `json:"id"`
    Username      string  `json:"username"`
    Email         string  `json:"email"`
    Fullname      string  `json:"fullname"`
    Phone         *string `json:"phone"`
    Avatar        *string `json:"avatar"`
    ProvinceApiID *int    `json:"province_api_id"`
    RegencyApiID  *int    `json:"regency_api_id"`
    DistrictApiID *int    `json:"district_api_id"`
    VillageApiID  *int    `json:"village_api_id"`
    FullAddress   *string `json:"full_address"`
    Role          string  `json:"role"`
}