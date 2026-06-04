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

type FollowRequest struct {
    FollowingID int `json:"following_id" binding:"required"`
}