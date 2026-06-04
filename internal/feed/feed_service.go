package feed

import (
    "database/sql"
    "fmt"
    "log"
    "strconv"
    "strings"

    "pengaduan_be2/pkg/db"
)

type FeedService struct{}

func NewFeedService() *FeedService {
    return &FeedService{}
}

// GetFeedPosts - ambil posts dengan filter
func (s *FeedService) GetFeedPosts(userID int, req *FeedRequest, provinceID, categoryID, status string) ([]FeedPost, int, error) {
    offset := (req.Page - 1) * req.Limit

    // Build WHERE conditions
    var conditions []string
    var args []interface{}

    // Base condition: post yang punya foto dari manapun sumbernya
    conditions = append(conditions, `(
        (c.photo IS NOT NULL AND c.photo != '') OR
        (c.investigation_evidence IS NOT NULL AND c.investigation_evidence != '' AND c.investigation_evidence != '[]') OR
        EXISTS (SELECT 1 FROM process_reports pr WHERE pr.complaint_id = c.id AND pr.process_photos IS NOT NULL AND pr.process_photos != '' AND pr.process_photos != '[]') OR
        EXISTS (SELECT 1 FROM completion_reports cr WHERE cr.complaint_id = c.id AND cr.final_photos IS NOT NULL AND cr.final_photos != '' AND cr.final_photos != '[]')
    )`)

    // Filter province
    if provinceID != "" && provinceID != "null" && provinceID != "0" {
        conditions = append(conditions, "c.province_api_id = ?")
        provinceInt, _ := strconv.Atoi(provinceID)
        args = append(args, provinceInt)
    }

    // Filter category
    if categoryID != "" && categoryID != "null" && categoryID != "0" {
        conditions = append(conditions, "c.category_id = ?")
        categoryInt, _ := strconv.Atoi(categoryID)
        args = append(args, categoryInt)
    }

    // Filter status
    if status != "" && status != "null" && status != "all" {
        conditions = append(conditions, "c.status = ?")
        args = append(args, status)
    }

    // Build WHERE clause
    whereClause := ""
    if len(conditions) > 0 {
        whereClause = "WHERE " + strings.Join(conditions, " AND ")
    }

    // Order by berdasarkan type
    orderBy := "c.created_at DESC"
    if req.Type == "popular" {
        orderBy = "(SELECT COUNT(*) FROM feed_likes WHERE post_id = c.id) DESC, c.created_at DESC"
    }

    // Main query
    query := fmt.Sprintf(`
        SELECT 
            c.id, 
            c.tracking_code, 
            c.user_id, 
            COALESCE(u.username, '') as user_name,
            COALESCE(u.fullname, '') as user_fullname,
            u.avatar as user_avatar,
            c.province_api_id,
            COALESCE(c.location_detail, '') as location_detail,
            COALESCE(c.description, '') as description,
            c.photo,
            c.status, 
            c.created_at, 
            c.updated_at,
            COALESCE((SELECT COUNT(*) FROM feed_likes WHERE post_id = c.id), 0) as likes_count,
            COALESCE((SELECT COUNT(*) FROM feed_comments WHERE post_id = c.id), 0) as comments_count,
            CASE WHEN EXISTS(SELECT 1 FROM feed_likes WHERE post_id = c.id AND user_id = ?) THEN 1 ELSE 0 END as is_liked,
            CASE WHEN EXISTS(SELECT 1 FROM feed_saves WHERE post_id = c.id AND user_id = ?) THEN 1 ELSE 0 END as is_saved
        FROM complaints c
        INNER JOIN users u ON c.user_id = u.id
        %s
        ORDER BY %s
        LIMIT ? OFFSET ?
    `, whereClause, orderBy)

    // Add userID for subqueries (2 kali) and pagination
    queryArgs := append([]interface{}{userID, userID}, args...)
    queryArgs = append(queryArgs, req.Limit, offset)

    rows, err := db.DB.Query(query, queryArgs...)
    if err != nil {
        log.Println("Error executing feed query:", err)
        return nil, 0, err
    }
    defer rows.Close()

    var posts []FeedPost
    for rows.Next() {
        var p FeedPost
        var userAvatar sql.NullString
        var photo sql.NullString
        var provinceApiID sql.NullInt64

        err := rows.Scan(
            &p.ID,
            &p.TrackingCode,
            &p.UserID,
            &p.UserName,
            &p.UserFullname,
            &userAvatar,
            &provinceApiID,
            &p.LocationDetail,
            &p.Description,
            &photo,
            &p.Status,
            &p.CreatedAt,
            &p.UpdatedAt,
            &p.LikesCount,
            &p.CommentsCount,
            &p.IsLiked,
            &p.IsSaved,
        )
        if err != nil {
            log.Println("Error scanning row:", err)
            continue
        }

        if userAvatar.Valid && userAvatar.String != "" {
            p.UserAvatar = &userAvatar.String
        }
        if provinceApiID.Valid {
            p.ProvinceID = int(provinceApiID.Int64)
        }
        if photo.Valid && photo.String != "" {
            p.Photo = &photo.String
        }

        posts = append(posts, p)
    }

    if err = rows.Err(); err != nil {
        log.Println("Error after rows iteration:", err)
        return nil, 0, err
    }

    // Count total dengan filter yang sama
    countQuery := fmt.Sprintf(`
        SELECT COUNT(*) 
        FROM complaints c
        INNER JOIN users u ON c.user_id = u.id
        %s
    `, whereClause)

    var total int
    err = db.DB.QueryRow(countQuery, args...).Scan(&total)
    if err != nil {
        log.Println("Error counting total:", err)
        return nil, 0, err
    }

    return posts, total, nil
}

// LikePost - like atau unlike post
func (s *FeedService) LikePost(userID, postID int) error {
    var exists bool
    err := db.DB.QueryRow("SELECT EXISTS(SELECT 1 FROM feed_likes WHERE user_id = ? AND post_id = ?)", userID, postID).Scan(&exists)
    if err != nil {
        return err
    }

    if exists {
        _, err = db.DB.Exec("DELETE FROM feed_likes WHERE user_id = ? AND post_id = ?", userID, postID)
    } else {
        _, err = db.DB.Exec("INSERT INTO feed_likes (user_id, post_id) VALUES (?, ?)", userID, postID)
    }
    return err
}

// SavePost - save atau unsave post
func (s *FeedService) SavePost(userID, postID int) error {
    var exists bool
    err := db.DB.QueryRow("SELECT EXISTS(SELECT 1 FROM feed_saves WHERE user_id = ? AND post_id = ?)", userID, postID).Scan(&exists)
    if err != nil {
        return err
    }

    if exists {
        _, err = db.DB.Exec("DELETE FROM feed_saves WHERE user_id = ? AND post_id = ?", userID, postID)
    } else {
        _, err = db.DB.Exec("INSERT INTO feed_saves (user_id, post_id) VALUES (?, ?)", userID, postID)
    }
    return err
}

// AddComment - tambah komentar
func (s *FeedService) AddComment(userID, postID int, text string) error {
    _, err := db.DB.Exec("INSERT INTO feed_comments (user_id, post_id, text) VALUES (?, ?, ?)", userID, postID, text)
    return err
}

// GetComments - ambil komentar
func (s *FeedService) GetComments(postID int) ([]FeedComment, error) {
    rows, err := db.DB.Query(`
        SELECT fc.id, fc.post_id, fc.user_id, COALESCE(u.username, '') as user_name, fc.text, fc.created_at
        FROM feed_comments fc
        JOIN users u ON fc.user_id = u.id
        WHERE fc.post_id = ?
        ORDER BY fc.created_at DESC
        LIMIT 50
    `, postID)
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    var comments []FeedComment
    for rows.Next() {
        var c FeedComment
        err := rows.Scan(&c.ID, &c.PostID, &c.UserID, &c.UserName, &c.Text, &c.CreatedAt)
        if err != nil {
            continue
        }
        comments = append(comments, c)
    }
    return comments, nil
}