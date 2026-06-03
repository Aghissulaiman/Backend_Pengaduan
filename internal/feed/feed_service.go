package feed

import (
    "database/sql"

    "pengaduan_be2/pkg/db"
)

type FeedService struct{}

func NewFeedService() *FeedService {
    return &FeedService{}
}

// GetFeedPosts - ambil posts untuk feed (completed + yang sudah ada bukti)
func (s *FeedService) GetFeedPosts(userID int, req *FeedRequest) ([]FeedPost, int, error) {
    offset := (req.Page - 1) * req.Limit

    query := `
        SELECT 
            c.id, c.tracking_code, c.user_id, 
            COALESCE(u.username, '') as user_name,
            COALESCE(u.fullname, '') as user_fullname,
            c.location_detail, c.description, 
            CASE 
                WHEN cr.final_photos IS NOT NULL AND cr.final_photos != '' THEN cr.final_photos
                WHEN c.photo IS NOT NULL AND c.photo != '' THEN c.photo
                ELSE NULL
            END as display_photo,
            c.status, c.created_at, c.updated_at,
            COALESCE((SELECT COUNT(*) FROM feed_likes WHERE post_id = c.id), 0) as likes_count,
            COALESCE((SELECT COUNT(*) FROM feed_comments WHERE post_id = c.id), 0) as comments_count,
            CASE WHEN EXISTS(SELECT 1 FROM feed_likes WHERE post_id = c.id AND user_id = ?) THEN 1 ELSE 0 END as is_liked,
            CASE WHEN EXISTS(SELECT 1 FROM feed_saves WHERE post_id = c.id AND user_id = ?) THEN 1 ELSE 0 END as is_saved,
            COALESCE(cr.final_photos, '') as completion_photo,
            COALESCE(pr.process_photos, '') as process_photo,
            COALESCE(cr.work_details, '') as work_details,
            COALESCE(pr.process_notes, '') as process_notes
        FROM complaints c
        JOIN users u ON c.user_id = u.id
        LEFT JOIN completion_reports cr ON c.id = cr.complaint_id AND cr.status = 'verified'
        LEFT JOIN process_reports pr ON c.id = pr.complaint_id AND pr.status = 'verified'
        WHERE c.status IN ('completed', 'process_report_verified', 'completion_report_verified')
           OR (c.status = 'investigation_done' AND c.investigation_result IS NOT NULL)
           OR (c.status = 'governor_processing' AND c.photo IS NOT NULL)
        ORDER BY c.updated_at DESC
        LIMIT ? OFFSET ?
    `

    rows, err := db.DB.Query(query, userID, userID, req.Limit, offset)
    if err != nil {
        return nil, 0, err
    }
    defer rows.Close()

    var posts []FeedPost
    for rows.Next() {
        var p FeedPost
        var photo, completionPhoto, processPhoto, workDetails, processNotes sql.NullString

        err := rows.Scan(
            &p.ID, &p.TrackingCode, &p.UserID,
            &p.UserName, &p.UserFullname,
            &p.LocationDetail, &p.Description,
            &photo, &p.Status,
            &p.CreatedAt, &p.UpdatedAt,
            &p.LikesCount, &p.CommentsCount,
            &p.IsLiked, &p.IsSaved,
            &completionPhoto, &processPhoto,
            &workDetails, &processNotes,
        )
        if err != nil {
            continue
        }
        
        // Prioritaskan foto dari completion_report, lalu process_report, lalu complaint asli
        if completionPhoto.Valid && completionPhoto.String != "" {
            p.Photo = &completionPhoto.String
        } else if processPhoto.Valid && processPhoto.String != "" {
            p.Photo = &processPhoto.String
        } else if photo.Valid && photo.String != "" {
            p.Photo = &photo.String
        }
        
        // Tambahkan deskripsi tambahan jika ada work_details atau process_notes
        if workDetails.Valid && workDetails.String != "" {
            p.Description = p.Description + "\n\n📋 Hasil Pekerjaan: " + workDetails.String
        }
        if processNotes.Valid && processNotes.String != "" {
            p.Description = p.Description + "\n\n📝 Catatan: " + processNotes.String
        }
        
        posts = append(posts, p)
    }

    // Count total
    var total int
    db.DB.QueryRow(`
        SELECT COUNT(*) FROM complaints c
        WHERE c.status IN ('completed', 'process_report_verified', 'completion_report_verified')
           OR (c.status = 'investigation_done' AND c.investigation_result IS NOT NULL)
           OR (c.status = 'governor_processing' AND c.photo IS NOT NULL)
    `).Scan(&total)

    return posts, total, nil
}

// LikePost - like atau unlike post
func (s *FeedService) LikePost(userID, postID int) error {
    var exists bool
    db.DB.QueryRow("SELECT EXISTS(SELECT 1 FROM feed_likes WHERE user_id = ? AND post_id = ?)", userID, postID).Scan(&exists)

    if exists {
        _, err := db.DB.Exec("DELETE FROM feed_likes WHERE user_id = ? AND post_id = ?", userID, postID)
        return err
    }
    _, err := db.DB.Exec("INSERT INTO feed_likes (user_id, post_id) VALUES (?, ?)", userID, postID)
    return err
}

// SavePost - save atau unsave post
func (s *FeedService) SavePost(userID, postID int) error {
    var exists bool
    db.DB.QueryRow("SELECT EXISTS(SELECT 1 FROM feed_saves WHERE user_id = ? AND post_id = ?)", userID, postID).Scan(&exists)

    if exists {
        _, err := db.DB.Exec("DELETE FROM feed_saves WHERE user_id = ? AND post_id = ?", userID, postID)
        return err
    }
    _, err := db.DB.Exec("INSERT INTO feed_saves (user_id, post_id) VALUES (?, ?)", userID, postID)
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
        SELECT fc.id, fc.post_id, fc.user_id, u.username, fc.text, fc.created_at
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