package province

import "time"

type Province struct {
    ID        int       `json:"id"`
    ApiID     int       `json:"api_id"`
    Name      string    `json:"name"`
    Code      *string   `json:"code"`
    IsActive  bool      `json:"is_active"`
    CreatedAt time.Time `json:"created_at"`
}

type Regency struct {
    ID           int     `json:"id"`
    ApiID        int     `json:"api_id"`
    ProvinceID   int     `json:"province_api_id"`
    ProvinceName string  `json:"province_name"`
    Name         string  `json:"name"`
    Type         string  `json:"type"`
}

type District struct {
    ID         int    `json:"id"`
    ApiID      int    `json:"api_id"`
    RegencyID  int    `json:"regency_id"`
    Name       string `json:"name"`
}

type Village struct {
    ID         int    `json:"id"`
    ApiID      int    `json:"api_id"`
    DistrictID int    `json:"district_id"`
    Name       string `json:"name"`
    PostalCode *string `json:"postal_code"`
}