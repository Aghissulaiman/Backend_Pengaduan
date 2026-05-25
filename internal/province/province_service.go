package province

import (
    "database/sql"
    "errors"
    "pengaduan_be2/pkg/db"
)

type ProvinceService struct{}

func NewProvinceService() *ProvinceService {
    return &ProvinceService{}
}

// GetAllProvinces ambil semua provinsi
func (s *ProvinceService) GetAllProvinces() ([]Province, error) {
    rows, err := db.DB.Query("SELECT id, api_id, name, code FROM provinces WHERE is_active = TRUE ORDER BY name")
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    var provinces []Province
    for rows.Next() {
        var p Province
        var code sql.NullString
        rows.Scan(&p.ID, &p.ApiID, &p.Name, &code)
        if code.Valid {
            p.Code = &code.String
        }
        provinces = append(provinces, p)
    }
    return provinces, nil
}

// GetProvinceByID ambil provinsi berdasarkan ID
func (s *ProvinceService) GetProvinceByID(id int) (*Province, error) {
    var p Province
    var code sql.NullString
    err := db.DB.QueryRow("SELECT id, api_id, name, code FROM provinces WHERE id = ? AND is_active = TRUE", id).
        Scan(&p.ID, &p.ApiID, &p.Name, &code)
    if err != nil {
        return nil, errors.New("provinsi tidak ditemukan")
    }
    if code.Valid {
        p.Code = &code.String
    }
    return &p, nil
}

// GetRegenciesByProvince ambil kabupaten/kota berdasarkan provinsi
func (s *ProvinceService) GetRegenciesByProvince(provinceID int) ([]Regency, error) {
    rows, err := db.DB.Query(`
        SELECT r.id, r.api_id, r.province_id, p.name, r.name, r.type
        FROM regencies r
        JOIN provinces p ON r.province_id = p.id
        WHERE r.province_id = ? AND r.is_active = TRUE
        ORDER BY r.name`, provinceID)
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    var regencies []Regency
    for rows.Next() {
        var reg Regency
        rows.Scan(&reg.ID, &reg.ApiID, &reg.ProvinceID, &reg.ProvinceName, &reg.Name, &reg.Type)
        regencies = append(regencies, reg)
    }
    return regencies, nil
}

// GetDistrictsByRegency ambil kecamatan berdasarkan kabupaten
func (s *ProvinceService) GetDistrictsByRegency(regencyID int) ([]District, error) {
    rows, err := db.DB.Query(`
        SELECT id, api_id, regency_id, name
        FROM districts
        WHERE regency_id = ? AND is_active = TRUE
        ORDER BY name`, regencyID)
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    var districts []District
    for rows.Next() {
        var d District
        rows.Scan(&d.ID, &d.ApiID, &d.RegencyID, &d.Name)
        districts = append(districts, d)
    }
    return districts, nil
}

// GetVillagesByDistrict ambil desa/kelurahan berdasarkan kecamatan
func (s *ProvinceService) GetVillagesByDistrict(districtID int) ([]Village, error) {
    rows, err := db.DB.Query(`
        SELECT id, api_id, district_id, name, postal_code
        FROM villages
        WHERE district_id = ? AND is_active = TRUE
        ORDER BY name`, districtID)
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    var villages []Village
    for rows.Next() {
        var v Village
        var postalCode sql.NullString
        rows.Scan(&v.ID, &v.ApiID, &v.DistrictID, &v.Name, &postalCode)
        if postalCode.Valid {
            v.PostalCode = &postalCode.String
        }
        villages = append(villages, v)
    }
    return villages, nil
}

// SyncProvinces sinkronisasi data provinsi dari API eksternal
func (s *ProvinceService) SyncProvinces(provinces []ProvinceData) error {
    for _, p := range provinces {
        _, err := db.DB.Exec(`
            INSERT INTO provinces (api_id, name) VALUES (?, ?)
            ON DUPLICATE KEY UPDATE name = VALUES(name)`,
            p.ID, p.Value)
        if err != nil {
            return err
        }
    }
    return nil
}

// SyncRegencies sinkronisasi data kabupaten dari API eksternal
func (s *ProvinceService) SyncRegencies(regencies []RegencyData) error {
    for _, r := range regencies {
        // Cari province_id berdasarkan api_id
        var provinceID int
        err := db.DB.QueryRow("SELECT id FROM provinces WHERE api_id = ?", r.ProvinceID).Scan(&provinceID)
        if err != nil {
            continue
        }

        _, err = db.DB.Exec(`
            INSERT INTO regencies (api_id, province_id, province_api_id, name, type)
            VALUES (?, ?, ?, ?, ?)
            ON DUPLICATE KEY UPDATE name = VALUES(name), type = VALUES(type)`,
            r.ID, provinceID, r.ProvinceID, r.Value, r.Type)
        if err != nil {
            return err
        }
    }
    return nil
}