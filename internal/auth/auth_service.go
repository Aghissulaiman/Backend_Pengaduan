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

func (s *AuthService) GetUserProfileByUsername(currentUserID int, username string) (*UserProfileResponse, error) {
    var profile UserProfileResponse
    var avatar, bio, provinceName sql.NullString
    var joinedAt time.Time

    query := `
        SELECT u.id, u.username, u.fullname, u.email, u.avatar, u.bio, u.role, p.name as province_name,
               u.created_at,
               (SELECT COUNT(*) FROM complaints WHERE user_id = u.id) as posts_count,
               (SELECT COUNT(*) FROM follows WHERE following_id = u.id AND status = 'accepted') as followers_count,
               (SELECT COUNT(*) FROM follows WHERE follower_id = u.id AND status = 'accepted') as following_count,
               EXISTS(SELECT 1 FROM follows WHERE follower_id = ? AND following_id = u.id AND status = 'accepted') as is_following
        FROM users u
        LEFT JOIN provinces p ON u.province_api_id = p.api_id
        WHERE u.username = ? AND u.is_active = TRUE
    `

    err := db.DB.QueryRow(query, currentUserID, username).Scan(
        &profile.ID, &profile.Username, &profile.Fullname, &profile.Email,
        &avatar, &bio, &profile.Role, &provinceName, &joinedAt,
        &profile.PostsCount, &profile.FollowersCount, &profile.FollowingCount, &profile.IsFollowing,
    )
    if err != nil {
        // 🔥 TAMBAHKAN LOG UNTUK DEBUG
        fmt.Printf("Error GetUserProfileByUsername: %v, currentUserID: %d, username: %s\n", err, currentUserID, username)
        return nil, errors.New("user tidak ditemukan")
    }

    if avatar.Valid {
        profile.Avatar = &avatar.String
    }
    if bio.Valid {
        profile.Bio = &bio.String
    }
    if provinceName.Valid {
        profile.ProvinceName = &provinceName.String
    }
    profile.JoinedDate = joinedAt.Format("2006-01-02")

    return &profile, nil
}

// FollowUser - follow user
func (s *AuthService) FollowUser(followerID, followingID int) error {
    if followerID == followingID {
        return errors.New("tidak bisa follow diri sendiri")
    }

    // Cek apakah sudah follow
    var exists bool
    db.DB.QueryRow("SELECT EXISTS(SELECT 1 FROM follows WHERE follower_id = ? AND following_id = ?)", 
        followerID, followingID).Scan(&exists)
    if exists {
        return errors.New("sudah follow")
    }

    // Insert follow dengan status 'accepted' (langsung accepted, tanpa pending)
    _, err := db.DB.Exec(`
        INSERT INTO follows (follower_id, following_id, status, created_at)
        VALUES (?, ?, 'accepted', NOW())`,
        followerID, followingID)
    
    return err
}

// UnfollowUser - unfollow user
func (s *AuthService) UnfollowUser(followerID, followingID int) error {
    result, err := db.DB.Exec(`
        DELETE FROM follows 
        WHERE follower_id = ? AND following_id = ?`,
        followerID, followingID)
    if err != nil {
        return err
    }
    
    rows, _ := result.RowsAffected()
    if rows == 0 {
        return errors.New("tidak sedang follow")
    }
    return nil
}

// GetUserPostsByUsername - ambil posts user lain
func (s *AuthService) GetUserPostsByUsername(currentUserID int, username string) ([]UserPostResponse, error) {
    query := `
        SELECT c.id, c.tracking_code, c.description, c.location_detail, c.status, c.created_at, c.photo,
               COALESCE((SELECT COUNT(*) FROM feed_likes WHERE post_id = c.id), 0) as likes_count,
               COALESCE((SELECT COUNT(*) FROM feed_comments WHERE post_id = c.id), 0) as comments_count,
               CASE WHEN EXISTS(SELECT 1 FROM feed_likes WHERE post_id = c.id AND user_id = ?) THEN 1 ELSE 0 END as is_liked,
               CASE WHEN EXISTS(SELECT 1 FROM feed_saves WHERE post_id = c.id AND user_id = ?) THEN 1 ELSE 0 END as is_saved
        FROM complaints c
        JOIN users u ON c.user_id = u.id
        WHERE u.username = ? AND c.status IN ('completed', 'process_report_verified', 'investigation_done')
        ORDER BY c.created_at DESC
    `

    rows, err := db.DB.Query(query, currentUserID, currentUserID, username)
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    var posts []UserPostResponse
    for rows.Next() {
        var p UserPostResponse
        var photo sql.NullString
        var createdAt time.Time

        err := rows.Scan(
            &p.ID, &p.TrackingCode, &p.Description, &p.LocationDetail,
            &p.Status, &createdAt, &photo,
            &p.LikesCount, &p.CommentsCount, &p.IsLiked, &p.IsSaved,
        )
        if err != nil {
            continue
        }
        p.CreatedAt = createdAt.Format("2006-01-02T15:04:05Z07:00")
        if photo.Valid {
            p.Photo = &photo.String
        }
        posts = append(posts, p)
    }

    return posts, nil
}

func min(a, b int) int {
    if a < b {
        return a
    }
    return b
}

