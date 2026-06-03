package feed

import "time"

// FeedPost - struct untuk post di feed
type FeedPost struct {
    ID           int        `json:"id"`
    TrackingCode string     `json:"tracking_code"`
    UserID       int        `json:"user_id"`
    UserName     string     `json:"user_name"`
    UserFullname string     `json:"user_fullname"`
    UserAvatar   *string    `json:"user_avatar"`
    ProvinceID   int        `json:"province_api_id"`
    LocationDetail string   `json:"location_detail"`
    CategoryID   int        `json:"category_id"`
    CategoryName string     `json:"category_name"`
    Description  string     `json:"description"`
    Photo        *string    `json:"photo"`
    Status       string     `json:"status"`
    LikesCount   int        `json:"likes_count"`
    CommentsCount int       `json:"comments_count"`
    IsLiked      bool       `json:"is_liked"`
    IsSaved      bool       `json:"is_saved"`
    CreatedAt    time.Time  `json:"created_at"`
    UpdatedAt    time.Time  `json:"updated_at"`
}

// FeedComment - struct untuk komentar
type FeedComment struct {
    ID        int       `json:"id"`
    PostID    int       `json:"post_id"`
    UserID    int       `json:"user_id"`
    UserName  string    `json:"user_name"`
    Text      string    `json:"text"`
    CreatedAt time.Time `json:"created_at"`
}