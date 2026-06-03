package province

import (
    "net/http"
    "strconv"
    "pengaduan_be2/internal/dto"
    "github.com/gin-gonic/gin"
)

type ProvinceHandler struct {
    service *ProvinceService
}

func NewProvinceHandler() *ProvinceHandler {
    return &ProvinceHandler{service: NewProvinceService()}
}

// GetAllProvinces godoc
// @Summary Get all provinces
// @Tags Province
// @Produce json
// @Success 200 {object} dto.Response
// @Router /api/provinces [get]
func (h *ProvinceHandler) GetAllProvinces(c *gin.Context) {
    provinces, err := h.service.GetAllProvinces()
    if err != nil {
        c.JSON(http.StatusInternalServerError, dto.Response{
            Success: false,
            Message: "Gagal mengambil data provinsi",
        })
        return
    }

    c.JSON(http.StatusOK, dto.Response{
        Success: true,
        Data:    provinces,
    })
}

// GetRegencies godoc
// @Summary Get regencies by province ID
// @Tags Province
// @Produce json
// @Param province_api_id path int true "Province ID"
// @Success 200 {object} dto.Response
// @Router /api/provinces/{province_api_id}/regencies [get]
func (h *ProvinceHandler) GetRegencies(c *gin.Context) {
    provinceID, err := strconv.Atoi(c.Param("province_api_id"))
    if err != nil {
        c.JSON(http.StatusBadRequest, dto.Response{
            Success: false,
            Message: "ID provinsi tidak valid",
        })
        return
    }

    regencies, err := h.service.GetRegenciesByProvince(provinceID)
    if err != nil {
        c.JSON(http.StatusInternalServerError, dto.Response{
            Success: false,
            Message: "Gagal mengambil data kabupaten/kota",
        })
        return
    }

    c.JSON(http.StatusOK, dto.Response{
        Success: true,
        Data:    regencies,
    })
}

// GetDistricts godoc
// @Summary Get districts by regency ID
// @Tags Province
// @Produce json
// @Param regency_id path int true "Regency ID"
// @Success 200 {object} dto.Response
// @Router /api/regencies/{regency_id}/districts [get]
func (h *ProvinceHandler) GetDistricts(c *gin.Context) {
    regencyID, err := strconv.Atoi(c.Param("regency_id"))
    if err != nil {
        c.JSON(http.StatusBadRequest, dto.Response{
            Success: false,
            Message: "ID kabupaten/kota tidak valid",
        })
        return
    }

    districts, err := h.service.GetDistrictsByRegency(regencyID)
    if err != nil {
        c.JSON(http.StatusInternalServerError, dto.Response{
            Success: false,
            Message: "Gagal mengambil data kecamatan",
        })
        return
    }

    c.JSON(http.StatusOK, dto.Response{
        Success: true,
        Data:    districts,
    })
}

// GetVillages godoc
// @Summary Get villages by district ID
// @Tags Province
// @Produce json
// @Param district_id path int true "District ID"
// @Success 200 {object} dto.Response
// @Router /api/districts/{district_id}/villages [get]
func (h *ProvinceHandler) GetVillages(c *gin.Context) {
    districtID, err := strconv.Atoi(c.Param("district_id"))
    if err != nil {
        c.JSON(http.StatusBadRequest, dto.Response{
            Success: false,
            Message: "ID kecamatan tidak valid",
        })
        return
    }

    villages, err := h.service.GetVillagesByDistrict(districtID)
    if err != nil {
        c.JSON(http.StatusInternalServerError, dto.Response{
            Success: false,
            Message: "Gagal mengambil data desa/kelurahan",
        })
        return
    }

    c.JSON(http.StatusOK, dto.Response{
        Success: true,
        Data:    villages,
    })
}

// SyncProvinces godoc
// @Summary Sync provinces from external API (admin only)
// @Tags Admin
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body SyncProvincesRequest true "Provinces data"
// @Success 200 {object} dto.Response
// @Router /api/admin/provinces/sync [post]
func (h *ProvinceHandler) SyncProvinces(c *gin.Context) {
    var req SyncProvincesRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, dto.Response{
            Success: false,
            Message: err.Error(),
        })
        return
    }

    if err := h.service.SyncProvinces(req.Provinces); err != nil {
        c.JSON(http.StatusInternalServerError, dto.Response{
            Success: false,
            Message: "Gagal sinkronisasi data provinsi",
        })
        return
    }

    c.JSON(http.StatusOK, dto.Response{
        Success: true,
        Message: "Sinkronisasi provinsi berhasil",
    })
}

// SyncRegencies godoc
// @Summary Sync regencies from external API (admin only)
// @Tags Admin
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body SyncRegenciesRequest true "Regencies data"
// @Success 200 {object} dto.Response
// @Router /api/admin/regencies/sync [post]
func (h *ProvinceHandler) SyncRegencies(c *gin.Context) {
    var req SyncRegenciesRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, dto.Response{
            Success: false,
            Message: err.Error(),
        })
        return
    }

    if err := h.service.SyncRegencies(req.Regencies); err != nil {
        c.JSON(http.StatusInternalServerError, dto.Response{
            Success: false,
            Message: "Gagal sinkronisasi data kabupaten/kota",
        })
        return
    }

    c.JSON(http.StatusOK, dto.Response{
        Success: true,
        Message: "Sinkronisasi kabupaten/kota berhasil",
    })
}