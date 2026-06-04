package auth

import "time"

type RegisterRequest struct {
    Username   string `json:"username" binding:"required,min=3"`
    Password   string `json:"password" binding:"required,min=6"`
    Email      string `json:"email" binding:"required,email"`
    Fullname   string `json:"fullname" binding:"required"`
    Phone      string `json:"phone"`
    ProvinceID int    `json:"province_api_id"`
}

type LoginRequest struct {
    Username string `json:"username" binding:"required"`
    Password string `json:"password" binding:"required"`
}

type GoogleLoginRequest struct {
    GoogleID   string `json:"google_id" binding:"required"`
    Email      string `json:"email" binding:"required"`
    Fullname   string `json:"fullname" binding:"required"`
    Avatar     string `json:"avatar"`
    ProvinceID int    `json:"province_api_id"`
}

type LoginResponse struct {
    Token string       `json:"token"`
    User  UserResponse `json:"user"`
}

type UpdateProfileRequest struct {
    Fullname      string `json:"fullname"`
    Phone         string `json:"phone"`
    ProvinceApiID int    `json:"province_api_id"`
    RegencyApiID  int    `json:"regency_api_id"`
    DistrictApiID int    `json:"district_api_id"`
    VillageApiID  int    `json:"village_api_id"`
    AddressDetail string `json:"address_detail"`
}

type SubmitComplaintRequest struct {
    CategoryID  int    `json:"category_id" binding:"required"`
    Location    string `json:"location" binding:"required"`
    Description string `json:"description" binding:"required"`
}

type SubmitComplaintResponse struct {
    ID           int    `json:"id"`
    TrackingCode string `json:"tracking_code"`
}

type Complaint struct {
    ID             int       `json:"id"`
    TrackingCode   string    `json:"tracking_code"`
    Description    string    `json:"description"`
    LocationDetail string    `json:"location_detail"`
    Status         string    `json:"status"`
    CreatedAt      time.Time `json:"created_at"`
}

// UserProfileResponse - response untuk profile user lain
type UserProfileResponse struct {
    ID            int     `json:"id"`
    Username      string  `json:"username"`
    Fullname      string  `json:"fullname"`
    Email         string  `json:"email"`
    Avatar        *string `json:"avatar"`
    Bio           *string `json:"bio"`
    Role          string  `json:"role"`
    ProvinceName  *string `json:"province_name"`
    IsFollowing   bool    `json:"is_following"`
    FollowersCount int    `json:"followers_count"`
    FollowingCount int    `json:"following_count"`
    PostsCount    int     `json:"posts_count"`
    JoinedDate    string  `json:"joined_date"`
}

// UserPostResponse - response untuk posts user
type UserPostResponse struct {
    ID           int     `json:"id"`
    TrackingCode string  `json:"tracking_code"`
    Description  string  `json:"description"`
    LocationDetail string `json:"location_detail"`
    Status       string  `json:"status"`
    CreatedAt    string  `json:"created_at"`
    Photo        *string `json:"photo"`
    LikesCount   int     `json:"likes_count"`
    CommentsCount int    `json:"comments_count"`
    IsLiked      bool    `json:"is_liked"`
    IsSaved      bool    `json:"is_saved"`
}