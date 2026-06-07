package chat

type SendMessageRequest struct {
    ReceiverID int    `json:"receiver_id" binding:"required"`
    Message    string `json:"message" binding:"required"`
    Type       string `json:"type"`
    FileURL    *string `json:"file_url"`
}

type GetConversationsQuery struct {
    Page  int `form:"page,default=1"`
    Limit int `form:"limit,default=20"`
}

type GetMessagesQuery struct {
    Page  int `form:"page,default=1"`
    Limit int `form:"limit,default:50"`
}

type FollowRequestResponse struct {
    FollowerID       int     `json:"follower_id"`
    FollowerUsername string  `json:"follower_username"`
    FollowerName     string  `json:"follower_name"`
    FollowerAvatar   *string `json:"follower_avatar"`
}
