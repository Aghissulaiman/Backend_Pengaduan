package province

type ProvinceResponse struct {
    ID   int    `json:"id"`
    Name string `json:"name"`
    Code *string `json:"code"`
}

type RegencyResponse struct {
    ID           int    `json:"id"`
    ProvinceID   int    `json:"province_api_id"`
    ProvinceName string `json:"province_name"`
    Name         string `json:"name"`
    Type         string `json:"type"`
}

type DistrictResponse struct {
    ID        int    `json:"id"`
    RegencyID int    `json:"regency_id"`
    Name      string `json:"name"`
}

type VillageResponse struct {
    ID         int     `json:"id"`
    DistrictID int     `json:"district_id"`
    Name       string  `json:"name"`
    PostalCode *string `json:"postal_code"`
}

type SyncProvincesRequest struct {
    Provinces []ProvinceData `json:"provinces"`
}

type ProvinceData struct {
    ID    int    `json:"id"`
    Value string `json:"value"`
}

type SyncRegenciesRequest struct {
    Regencies []RegencyData `json:"regencies"`
}

type RegencyData struct {
    ID         int    `json:"id"`
    ProvinceID int    `json:"province_api_id"`
    Value      string `json:"value"`
    Type       string `json:"type"`
}