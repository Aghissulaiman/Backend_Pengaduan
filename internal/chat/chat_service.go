package chat

import (
    "database/sql"
    "errors"

    "pengaduan_be2/pkg/db"
)

type ChatService struct{}

func NewChatService() *ChatService {
    return &ChatService{}
}

// GetOrCreateConversation - ambil atau buat percakapan baru
func (s *ChatService) GetOrCreateConversation(userID, otherUserID int) (*Conversation, error) {
    if userID == otherUserID {
        return nil, errors.New("tidak bisa chat dengan diri sendiri")
    }

    var conv Conversation
    var lastMsg sql.NullString
    var lastMsgAt sql.NullTime

    err := db.DB.QueryRow(`
        SELECT c.id, c.participant1_id, c.participant2_id,
               c.last_message, c.last_message_at,
               c.participant1_unread_count, c.participant2_unread_count,
               c.created_at, c.updated_at
        FROM conversations c
        WHERE (c.participant1_id = ? AND c.participant2_id = ?)
           OR (c.participant1_id = ? AND c.participant2_id = ?)
    `, userID, otherUserID, otherUserID, userID).Scan(
        &conv.ID, &conv.Participant1ID, &conv.Participant2ID,
        &lastMsg, &lastMsgAt,
        &conv.Participant1UnreadCount, &conv.Participant2UnreadCount,
        &conv.CreatedAt, &conv.UpdatedAt,
    )

    if err == nil {
        if lastMsg.Valid {
            conv.LastMessage = &lastMsg.String
            conv.LastMessageAt = &lastMsgAt.Time
        }
        if conv.Participant1ID == userID {
            conv.UnreadCount = conv.Participant1UnreadCount
        } else {
            conv.UnreadCount = conv.Participant2UnreadCount
        }
        return &conv, nil
    }

    if err != sql.ErrNoRows {
        return nil, err
    }

    result, err := db.DB.Exec(`
        INSERT INTO conversations (participant1_id, participant2_id, created_at, updated_at)
        VALUES (?, ?, NOW(), NOW())`, userID, otherUserID)
    if err != nil {
        return nil, err
    }

    convID, _ := result.LastInsertId()
    conv.ID = int(convID)
    conv.Participant1ID = userID
    conv.Participant2ID = otherUserID
    conv.UnreadCount = 0
    conv.Participant1UnreadCount = 0
    conv.Participant2UnreadCount = 0

    return &conv, nil
}

// GetUserConversations - ambil semua percakapan user
func (s *ChatService) GetUserConversations(userID int, query *GetConversationsQuery) ([]Conversation, int, error) {
    offset := (query.Page - 1) * query.Limit

    rows, err := db.DB.Query(`
        SELECT c.id, c.participant1_id, c.participant2_id,
               c.last_message, c.last_message_at,
               c.participant1_unread_count, c.participant2_unread_count,
               c.created_at, c.updated_at
        FROM conversations c
        WHERE c.participant1_id = ? OR c.participant2_id = ?
        ORDER BY c.last_message_at DESC
        LIMIT ? OFFSET ?
    `, userID, userID, query.Limit, offset)

    if err != nil {
        return nil, 0, err
    }
    defer rows.Close()

    var conversations []Conversation
    for rows.Next() {
        var conv Conversation
        var lastMsg sql.NullString
        var lastMsgAt sql.NullTime

        err := rows.Scan(
            &conv.ID, &conv.Participant1ID, &conv.Participant2ID,
            &lastMsg, &lastMsgAt,
            &conv.Participant1UnreadCount, &conv.Participant2UnreadCount,
            &conv.CreatedAt, &conv.UpdatedAt,
        )
        if err != nil {
            continue
        }

        if lastMsg.Valid {
            conv.LastMessage = &lastMsg.String
            conv.LastMessageAt = &lastMsgAt.Time
        }

        if conv.Participant1ID == userID {
            conv.UnreadCount = conv.Participant1UnreadCount
        } else {
            conv.UnreadCount = conv.Participant2UnreadCount
        }

        conversations = append(conversations, conv)
    }

    var total int
    db.DB.QueryRow(`
        SELECT COUNT(*) FROM conversations 
        WHERE participant1_id = ? OR participant2_id = ?`, userID, userID).Scan(&total)

    return conversations, total, nil
}

// GetMessages - ambil pesan dalam percakapan
func (s *ChatService) GetMessages(conversationID, userID int, query *GetMessagesQuery) ([]Message, int, error) {
    offset := (query.Page - 1) * query.Limit

    var isParticipant bool
    db.DB.QueryRow(`
        SELECT EXISTS(SELECT 1 FROM conversations 
        WHERE id = ? AND (participant1_id = ? OR participant2_id = ?))`,
        conversationID, userID, userID).Scan(&isParticipant)

    if !isParticipant {
        return nil, 0, errors.New("tidak memiliki akses ke percakapan ini")
    }

    rows, err := db.DB.Query(`
        SELECT m.id, m.conversation_id, m.sender_id, m.receiver_id,
               m.message, m.type, m.file_url, m.is_read, m.created_at
        FROM messages m
        WHERE m.conversation_id = ? AND m.is_deleted = FALSE
        ORDER BY m.created_at DESC
        LIMIT ? OFFSET ?
    `, conversationID, query.Limit, offset)

    if err != nil {
        return nil, 0, err
    }
    defer rows.Close()

    var messages []Message
    for rows.Next() {
        var msg Message
        var fileURL sql.NullString

        err := rows.Scan(
            &msg.ID, &msg.ConversationID, &msg.SenderID, &msg.ReceiverID,
            &msg.Message, &msg.Type, &fileURL, &msg.IsRead, &msg.CreatedAt,
        )
        if err != nil {
            continue
        }
        if fileURL.Valid {
            msg.FileURL = &fileURL.String
        }

        // Ambil nama sender dan receiver
        db.DB.QueryRow("SELECT COALESCE(fullname, username) FROM users WHERE id = ?", msg.SenderID).Scan(&msg.SenderName)
        db.DB.QueryRow("SELECT COALESCE(fullname, username) FROM users WHERE id = ?", msg.ReceiverID).Scan(&msg.ReceiverName)

        messages = append(messages, msg)
    }

    // Mark messages as read
    db.DB.Exec(`
        UPDATE messages SET is_read = TRUE 
        WHERE conversation_id = ? AND receiver_id = ? AND is_read = FALSE`,
        conversationID, userID)

    // Update unread count
    db.DB.Exec(`
        UPDATE conversations 
        SET participant1_unread_count = CASE 
            WHEN participant1_id = ? THEN 0 ELSE participant1_unread_count END,
            participant2_unread_count = CASE 
            WHEN participant2_id = ? THEN 0 ELSE participant2_unread_count END
        WHERE id = ?`,
        userID, userID, conversationID)

    var total int
    db.DB.QueryRow(`
        SELECT COUNT(*) FROM messages 
        WHERE conversation_id = ? AND is_deleted = FALSE`, conversationID).Scan(&total)

    return messages, total, nil
}

// SendMessage - kirim pesan
func (s *ChatService) SendMessage(senderID int, req *SendMessageRequest) (*Message, error) {
    conv, err := s.GetOrCreateConversation(senderID, req.ReceiverID)
    if err != nil {
        return nil, err
    }

    msgType := "text"
    if req.Type != "" {
        msgType = req.Type
    }

    result, err := db.DB.Exec(`
        INSERT INTO messages (conversation_id, sender_id, receiver_id, message, type, file_url, created_at)
        VALUES (?, ?, ?, ?, ?, ?, NOW())`,
        conv.ID, senderID, req.ReceiverID, req.Message, msgType, req.FileURL)
    if err != nil {
        return nil, err
    }

    msgID, _ := result.LastInsertId()

    db.DB.Exec(`
        UPDATE conversations 
        SET last_message = ?, last_message_at = NOW(),
            participant1_unread_count = CASE 
                WHEN participant1_id = ? THEN participant1_unread_count + 1 
                ELSE participant1_unread_count END,
            participant2_unread_count = CASE 
                WHEN participant2_id = ? THEN participant2_unread_count + 1 
                ELSE participant2_unread_count END,
            updated_at = NOW()
        WHERE id = ?`,
        req.Message, req.ReceiverID, senderID, conv.ID)

    var msg Message
    var fileURL sql.NullString

    err = db.DB.QueryRow(`
        SELECT m.id, m.conversation_id, m.sender_id, m.receiver_id,
               m.message, m.type, m.file_url, m.is_read, m.created_at
        FROM messages m
        WHERE m.id = ?`, msgID).Scan(
        &msg.ID, &msg.ConversationID, &msg.SenderID, &msg.ReceiverID,
        &msg.Message, &msg.Type, &fileURL, &msg.IsRead, &msg.CreatedAt,
    )
    if err != nil {
        return &msg, nil
    }
    if fileURL.Valid {
        msg.FileURL = &fileURL.String
    }

    db.DB.QueryRow("SELECT COALESCE(fullname, username) FROM users WHERE id = ?", msg.SenderID).Scan(&msg.SenderName)
    db.DB.QueryRow("SELECT COALESCE(fullname, username) FROM users WHERE id = ?", msg.ReceiverID).Scan(&msg.ReceiverName)

    return &msg, nil
}

// GetChatUsers - ambil daftar user untuk chat
func (s *ChatService) GetChatUsers(userID int, search string) ([]ChatUser, error) {
    query := `
        SELECT id, username, fullname, email, role, avatar
        FROM users
        WHERE id != ? AND is_active = TRUE
    `
    args := []interface{}{userID}

    if search != "" {
        query += " AND (username LIKE ? OR fullname LIKE ? OR email LIKE ?)"
        searchTerm := "%" + search + "%"
        args = append(args, searchTerm, searchTerm, searchTerm)
    }

    query += " ORDER BY fullname ASC LIMIT 50"

    rows, err := db.DB.Query(query, args...)
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    var users []ChatUser
    for rows.Next() {
        var u ChatUser
        var avatar sql.NullString

        err := rows.Scan(&u.ID, &u.Username, &u.Fullname, &u.Email, &u.Role, &avatar)
        if err != nil {
            continue
        }
        if avatar.Valid {
            u.Avatar = &avatar.String
        }
        users = append(users, u)
    }

    return users, nil
}

/// GetContacts - ambil daftar kontak
func (s *ChatService) GetContacts(userID int, search string) ([]Contact, error) {
    query := `
        SELECT 
            u.id, 
            u.username, 
            u.fullname, 
            u.email, 
            u.role, 
            u.avatar,
            COALESCE(f1.status = 'accepted', false) as is_following,
            COALESCE(f2.status = 'accepted', false) as is_followed_by,
            NULL as last_message,
            0 as unread_count
        FROM users u
        LEFT JOIN follows f1 ON f1.follower_id = ? AND f1.following_id = u.id
        LEFT JOIN follows f2 ON f2.follower_id = u.id AND f2.following_id = ?
        WHERE u.id != ? AND u.is_active = true
    `
    args := []interface{}{userID, userID, userID}
    
    if search != "" {
        query += " AND (u.username LIKE ? OR u.fullname LIKE ? OR u.email LIKE ?)"
        searchTerm := "%" + search + "%"
        args = append(args, searchTerm, searchTerm, searchTerm)
    }
    
    query += " ORDER BY u.fullname ASC LIMIT 100"
    
    rows, err := db.DB.Query(query, args...)
    if err != nil {
        return nil, err
    }
    defer rows.Close()
    
    var contacts []Contact
    for rows.Next() {
        var c Contact
        var avatar sql.NullString
        
        err := rows.Scan(
            &c.ID, &c.Username, &c.Fullname, &c.Email, &c.Role, &avatar,
            &c.IsFollowing, &c.IsFollowedBy, &c.LastMessage, &c.UnreadCount,
        )
        if err != nil {
            continue
        }
        if avatar.Valid {
            c.Avatar = &avatar.String
        }
        contacts = append(contacts, c)
    }
    
    return contacts, nil
}

// FollowUser - follow user
func (s *ChatService) FollowUser(followerID, followingID int) error {
    if followerID == followingID {
        return errors.New("tidak bisa follow diri sendiri")
    }

    var exists bool
    db.DB.QueryRow("SELECT EXISTS(SELECT 1 FROM follows WHERE follower_id = ? AND following_id = ?)", followerID, followingID).Scan(&exists)
    if exists {
        return errors.New("sudah follow")
    }

    _, err := db.DB.Exec("INSERT INTO follows (follower_id, following_id, status) VALUES (?, ?, 'pending')", followerID, followingID)
    return err
}

// AcceptFollow - accept follow request
func (s *ChatService) AcceptFollow(followerID, followingID int) error {
    result, err := db.DB.Exec("UPDATE follows SET status = 'accepted' WHERE follower_id = ? AND following_id = ? AND status = 'pending'", followerID, followingID)
    if err != nil {
        return err
    }
    rows, _ := result.RowsAffected()
    if rows == 0 {
        return errors.New("permintaan follow tidak ditemukan")
    }
    return nil
}

// RejectFollow - reject follow request
func (s *ChatService) RejectFollow(followerID, followingID int) error {
    result, err := db.DB.Exec("DELETE FROM follows WHERE follower_id = ? AND following_id = ? AND status = 'pending'", followerID, followingID)
    if err != nil {
        return err
    }
    rows, _ := result.RowsAffected()
    if rows == 0 {
        return errors.New("permintaan follow tidak ditemukan")
    }
    return nil
}

// GetFollowRequests - get pending follow requests
func (s *ChatService) GetFollowRequests(userID int) ([]FollowRequestResponse, error) {
    rows, err := db.DB.Query(`
        SELECT u.id, u.username, u.fullname, u.avatar
        FROM follows f
        JOIN users u ON f.follower_id = u.id
        WHERE f.following_id = ? AND f.status = 'pending'
        ORDER BY f.created_at DESC
    `, userID)
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    var requests []FollowRequestResponse
    for rows.Next() {
        var req FollowRequestResponse
        var avatar sql.NullString
        err := rows.Scan(&req.FollowerID, &req.FollowerUsername, &req.FollowerName, &avatar)
        if err != nil {
            continue
        }
        if avatar.Valid {
            req.FollowerAvatar = &avatar.String
        }
        requests = append(requests, req)
    }
    return requests, nil
}