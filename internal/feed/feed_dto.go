package feed

// FeedRequest - request untuk get feed
type FeedRequest struct {
    Page  int    `form:"page,default=1"`
    Limit int    `form:"limit,default=20"`
    Type  string `form:"type,default=for-you"`
}

// LikeRequest - request untuk like/unlike
type LikeRequest struct {
    PostID int `json:"post_id" binding:"required"`
}

// SaveRequest - request untuk save/unsave
type SaveRequest struct {
    PostID int `json:"post_id" binding:"required"`
}

// CommentRequest - request untuk tambah komentar
type CommentRequest struct {
    PostID int    `json:"post_id" binding:"required"`
    Text   string `json:"text" binding:"required"`
}

// FeedResponse - response untuk get feed
type FeedResponse struct {
    Posts      []FeedPost `json:"posts"`
    Total      int        `json:"total"`
    Page       int        `json:"page"`
    Limit      int        `json:"limit"`
    TotalPages int        `json:"total_pages"`
}