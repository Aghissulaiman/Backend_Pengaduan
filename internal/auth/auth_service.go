package auth

import (
    "database/sql"
    "errors"
    "fmt"
    "os"
    "time"
    "pengaduan_be2/pkg/db"
    "pengaduan_be2/pkg/utils"

    "github.com/golang-jwt/jwt/v5"
    "golang.org/x/crypto/bcrypt"
)

type AuthService struct{}

func NewAuthService() *AuthService {
    return &AuthService{}
}

func (s *AuthService) Register(req *RegisterRequest) error {
    var exists int
    err := db.DB.QueryRow("SELECT COUNT(*) FROM users WHERE username = ? OR email = ?",
        req.Username, req.Email).Scan(&exists)
    if err != nil {
        return errors.New("gagal cek user")
    }
    if exists > 0 {
        return errors.New("username atau email sudah terdaftar")
    }

    hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
    if err != nil {
        return errors.New("gagal hash password")
    }

    _, err = db.DB.Exec(`
        INSERT INTO users (username, password, email, fullname, phone, role, is_verified)
        VALUES (?, ?, ?, ?, ?, 'user', FALSE)`,
        req.Username, string(hashedPassword), req.Email, req.Fullname,
        utils.NullIfEmpty(req.Phone),
    )

    return err
}

func (s *AuthService) Login(req *LoginRequest) (*LoginResponse, error) {
    var user User
    var hashedPassword string
    var fullAddress sql.NullString
    var provinceApiID, regencyApiID, districtApiID, villageApiID sql.NullInt64

    query := `SELECT id, username, email, fullname, phone, avatar, 
              province_api_id, regency_api_id, district_api_id, village_api_id, 
              full_address, role, password 
              FROM users WHERE username = ? AND is_active = TRUE`
    err := db.DB.QueryRow(query, req.Username).Scan(
        &user.ID, &user.Username, &user.Email, &user.Fullname,
        &user.Phone, &user.Avatar,
        &provinceApiID, &regencyApiID, &districtApiID, &villageApiID,
        &fullAddress, &user.Role, &hashedPassword,
    )

    if err != nil {
        return nil, errors.New("username atau password salah")
    }

    if provinceApiID.Valid {
        val := int(provinceApiID.Int64)
        user.ProvinceApiID = &val
    }
    if regencyApiID.Valid {
        val := int(regencyApiID.Int64)
        user.RegencyApiID = &val
    }
    if districtApiID.Valid {
        val := int(districtApiID.Int64)
        user.DistrictApiID = &val
    }
    if villageApiID.Valid {
        val := int(villageApiID.Int64)
        user.VillageApiID = &val
    }
    if fullAddress.Valid {
        user.FullAddress = &fullAddress.String
    }

    if err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(req.Password)); err != nil {
        return nil, errors.New("username atau password salah")
    }

    db.DB.Exec("UPDATE users SET last_login = NOW() WHERE id = ?", user.ID)

    secret := []byte(os.Getenv("JWT_SECRET"))
    token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
        "user_id":     user.ID,
        "username":    user.Username,
        "role":        user.Role,
        "province_api_id": user.ProvinceApiID,
        "exp":         time.Now().Add(24 * time.Hour).Unix(),
    })

    tokenString, err := token.SignedString(secret)
    if err != nil {
        return nil, errors.New("gagal generate token")
    }

    resp := &LoginResponse{Token: tokenString}
    resp.User.ID = user.ID
    resp.User.Username = user.Username
    resp.User.Email = user.Email
    resp.User.Fullname = user.Fullname
    resp.User.Phone = user.Phone
    resp.User.Avatar = user.Avatar
    resp.User.ProvinceApiID = user.ProvinceApiID
    resp.User.RegencyApiID = user.RegencyApiID
    resp.User.DistrictApiID = user.DistrictApiID
    resp.User.VillageApiID = user.VillageApiID
    resp.User.FullAddress = user.FullAddress
    resp.User.Role = user.Role

    return resp, nil
}

func (s *AuthService) GoogleLogin(req *GoogleLoginRequest) (*LoginResponse, error) {
    var user User
    var provinceApiID, regencyApiID, districtApiID, villageApiID sql.NullInt64
    var fullAddress sql.NullString

    err := db.DB.QueryRow(`
        SELECT id, username, email, fullname, phone, avatar, 
               province_api_id, regency_api_id, district_api_id, village_api_id, 
               full_address, role 
        FROM users WHERE google_id = ? OR email = ?`,
        req.GoogleID, req.Email,
    ).Scan(&user.ID, &user.Username, &user.Email, &user.Fullname,
        &user.Phone, &user.Avatar,
        &provinceApiID, &regencyApiID, &districtApiID, &villageApiID,
        &fullAddress, &user.Role)

    if err == sql.ErrNoRows {
        username := req.Email[:min(20, len(req.Email))]

        _, err := db.DB.Exec(`
            INSERT INTO users (username, email, fullname, avatar, google_id, role, is_verified)
            VALUES (?, ?, ?, ?, ?, 'user', TRUE)`,
            username, req.Email, req.Fullname, req.Avatar, req.GoogleID,
        )
        if err != nil {
            return nil, errors.New("gagal membuat akun")
        }

        db.DB.QueryRow(`
            SELECT id, username, email, fullname, phone, avatar, 
                   province_api_id, regency_api_id, district_api_id, village_api_id, 
                   full_address, role 
            FROM users WHERE google_id = ?`,
            req.GoogleID,
        ).Scan(&user.ID, &user.Username, &user.Email, &user.Fullname,
            &user.Phone, &user.Avatar,
            &provinceApiID, &regencyApiID, &districtApiID, &villageApiID,
            &fullAddress, &user.Role)
    } else if err != nil {
        return nil, errors.New("gagal login dengan Google")
    }

    if provinceApiID.Valid {
        val := int(provinceApiID.Int64)
        user.ProvinceApiID = &val
    }
    if regencyApiID.Valid {
        val := int(regencyApiID.Int64)
        user.RegencyApiID = &val
    }
    if districtApiID.Valid {
        val := int(districtApiID.Int64)
        user.DistrictApiID = &val
    }
    if villageApiID.Valid {
        val := int(villageApiID.Int64)
        user.VillageApiID = &val
    }
    if fullAddress.Valid {
        user.FullAddress = &fullAddress.String
    }

    db.DB.Exec("UPDATE users SET last_login = NOW() WHERE id = ?", user.ID)

    secret := []byte(os.Getenv("JWT_SECRET"))
    token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
        "user_id":     user.ID,
        "username":    user.Username,
        "role":        user.Role,
        "province_api_id": user.ProvinceApiID,
        "exp":         time.Now().Add(24 * time.Hour).Unix(),
    })

    tokenString, err := token.SignedString(secret)
    if err != nil {
        return nil, errors.New("gagal generate token")
    }

    resp := &LoginResponse{Token: tokenString}
    resp.User.ID = user.ID
    resp.User.Username = user.Username
    resp.User.Email = user.Email
    resp.User.Fullname = user.Fullname
    resp.User.Phone = user.Phone
    resp.User.Avatar = user.Avatar
    resp.User.ProvinceApiID = user.ProvinceApiID
    resp.User.RegencyApiID = user.RegencyApiID
    resp.User.DistrictApiID = user.DistrictApiID
    resp.User.VillageApiID = user.VillageApiID
    resp.User.FullAddress = user.FullAddress
    resp.User.Role = user.Role

    return resp, nil
}

func (s *AuthService) GetUserByID(userID int) (*UserResponse, error) {
    var user UserResponse
    var phone, avatar, fullAddress sql.NullString
    var provinceApiID, regencyApiID, districtApiID, villageApiID sql.NullInt64

    query := `
        SELECT 
            u.id, u.username, u.email, u.fullname, u.phone, u.avatar, 
            u.province_api_id, u.regency_api_id, u.district_api_id, u.village_api_id, 
            u.full_address, u.role
        FROM users u
        WHERE u.id = ? AND u.is_active = TRUE
    `

    err := db.DB.QueryRow(query, userID).Scan(
        &user.ID, &user.Username, &user.Email, &user.Fullname,
        &phone, &avatar,
        &provinceApiID, &regencyApiID, &districtApiID, &villageApiID,
        &fullAddress, &user.Role,
    )

    if err != nil {
        return nil, errors.New("user tidak ditemukan")
    }

    if phone.Valid {
        user.Phone = &phone.String
    }
    if avatar.Valid {
        user.Avatar = &avatar.String
    }
    if fullAddress.Valid {
        user.FullAddress = &fullAddress.String
    }

    if provinceApiID.Valid {
        val := int(provinceApiID.Int64)
        user.ProvinceApiID = &val
    }
    if regencyApiID.Valid {
        val := int(regencyApiID.Int64)
        user.RegencyApiID = &val
    }
    if districtApiID.Valid {
        val := int(districtApiID.Int64)
        user.DistrictApiID = &val
    }
    if villageApiID.Valid {
        val := int(villageApiID.Int64)
        user.VillageApiID = &val
    }

    return &user, nil
}

func (s *AuthService) UpdateProfile(userID int, req *UpdateProfileRequest) error {
    _, err := db.DB.Exec(`
        UPDATE users 
        SET fullname = COALESCE(NULLIF(?, ''), fullname),
            phone = COALESCE(NULLIF(?, ''), phone),
            province_api_id = COALESCE(?, province_api_id),
            regency_api_id = COALESCE(?, regency_api_id),
            district_api_id = COALESCE(?, district_api_id),
            village_api_id = COALESCE(?, village_api_id),
            full_address = COALESCE(NULLIF(?, ''), full_address),
            updated_at = NOW()
        WHERE id = ?`,
        req.Fullname,
        req.Phone,
        req.ProvinceApiID,
        req.RegencyApiID,
        req.DistrictApiID,
        req.VillageApiID,
        req.AddressDetail,
        userID,
    )

    if err != nil {
        return fmt.Errorf("gagal update profile: %v", err)
    }

    return nil
}

func (s *AuthService) GetUserComplaints(userID int) ([]Complaint, error) {
    rows, err := db.DB.Query(`
        SELECT id, tracking_code, description, location_detail, status, created_at
        FROM complaints
        WHERE user_id = ?
        ORDER BY created_at DESC`, userID)
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    var complaints []Complaint
    for rows.Next() {
        var c Complaint
        rows.Scan(&c.ID, &c.TrackingCode, &c.Description, &c.LocationDetail, &c.Status, &c.CreatedAt)
        complaints = append(complaints, c)
    }
    return complaints, nil
}

func (s *AuthService) SubmitComplaint(userID int, req *SubmitComplaintRequest) (*SubmitComplaintResponse, error) {
    trackingCode := fmt.Sprintf("TRK%d", time.Now().UnixNano())

    result, err := db.DB.Exec(`
        INSERT INTO complaints (tracking_code, user_id, category_id, location, description, status)
        VALUES (?, ?, ?, ?, ?, 'pending')`,
        trackingCode, userID, req.CategoryID, req.Location, req.Description)
    if err != nil {
        return nil, errors.New("gagal menyimpan pengaduan")
    }

    id, _ := result.LastInsertId()

    return &SubmitComplaintResponse{
        ID:           int(id),
        TrackingCode: trackingCode,
    }, nil
}

func min(a, b int) int {
    if a < b {
        return a
    }
    return b
}