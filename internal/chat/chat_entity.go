package chat

import "time"

type Conversation struct {
    ID                    int        `json:"id"`
    Participant1ID        int        `json:"participant1_id"`
    Participant2ID        int        `json:"participant2_id"`
    Participant1Name      string     `json:"participant1_name"`
    Participant2Name      string     `json:"participant2_name"`
    Participant1Avatar    *string    `json:"participant1_avatar"`
    Participant2Avatar    *string    `json:"participant2_avatar"`
    LastMessage           *string    `json:"last_message"`
    LastMessageAt         *time.Time `json:"last_message_at"`
    Participant1UnreadCount int      `json:"participant1_unread_count"`
    Participant2UnreadCount int      `json:"participant2_unread_count"`
    UnreadCount           int        `json:"unread_count"`
    CreatedAt             time.Time  `json:"created_at"`
    UpdatedAt             time.Time  `json:"updated_at"`
}

type Message struct {
    ID             int        `json:"id"`
    ConversationID int        `json:"conversation_id"`
    SenderID       int        `json:"sender_id"`
    SenderName     string     `json:"sender_name"`
    SenderAvatar   *string    `json:"sender_avatar"`
    ReceiverID     int        `json:"receiver_id"`
    ReceiverName   string     `json:"receiver_name"`
    Message        string     `json:"message"`
    Type           string     `json:"type"`
    FileURL        *string    `json:"file_url"`
    IsRead         bool       `json:"is_read"`
    CreatedAt      time.Time  `json:"created_at"`
}

type ChatUser struct {
    ID       int     `json:"id"`
    Username string  `json:"username"`
    Fullname string  `json:"fullname"`
    Email    string  `json:"email"`
    Role     string  `json:"role"`
    Avatar   *string `json:"avatar"`
    IsFollow bool    `json:"is_follow"`
}

type Contact struct {
    ID            int     `json:"id"`
    Username      string  `json:"username"`
    Fullname      string  `json:"fullname"`
    Email         string  `json:"email"`
    Role          string  `json:"role"`
    Avatar        *string `json:"avatar"`
    IsFollowing   bool    `json:"is_following"`
    IsFollowedBy  bool    `json:"is_followed_by"`
    LastMessage   *string `json:"last_message"`
    UnreadCount   int     `json:"unread_count"`
}

