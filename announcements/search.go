package announcements

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/lib/pq"
	"github.com/pgvector/pgvector-go"

	"announcements/embedding"
	pb "announcements/proto"
)

const (
	limitAnnouncementsToShowInt = 5
	whereSnippetSQLSearchString = ` (title ILIKE '%%' || $%d || '%%'
		OR description ILIKE '%%' || $%d || '%%'
		OR SIMILARITY(title, $%d) > 0.3
		OR SIMILARITY(description, $%d) > 0.3)`
	whereSnippetSQLUserID      = " (announcement_author_id = $%d)"
	whereSnippetSQLCategory    = " (category = $%d)"
	snippetSQLOrderByCreateAt  = " ORDER BY create_at DESC"
	snippetSQLOrderByEmbedding = " ORDER BY embedding <=> (SELECT embedding FROM users WHERE user_id = $%d)"
	snippetSQLOffsetAndLimit   = " OFFSET $%d LIMIT $%d"
)

const (
	PathToDefaultImage = "d.webp"
)

var imagesServiceExternalConnections string

func getAnnouncementInfo(announcementID, userID int32) ([]*pb.AnnouncementData, error) {
	announcementData := []*pb.AnnouncementData{}

	var authorID int
	var authorName string
	var userEmbedding pgvector.Vector
	var title string
	var description string
	var images []string
	var category string
	var announcementEmbedding pgvector.Vector
	fmt.Println(announcementID)
	sqlRow := db.QueryRow("SELECT title, description, announcement_author_id, images_path, category, embedding FROM announcements WHERE announcement_id = $1", announcementID)
	if err := sqlRow.Scan(&title, &description, &authorID, pq.Array(&images), &category, &announcementEmbedding); err != nil {
		return announcementData, errors.New("Нет объявления с таким id")
	}

	if err := db.QueryRow("SELECT name, embedding FROM users WHERE user_id = $1", authorID).Scan(&authorName, &userEmbedding); err != nil {
		authorName = "Неизвестно"
	}

	userEmbeddingFloat32 := userEmbedding.Slice()
	announcementEmbeddingFloat32 := announcementEmbedding.Slice()
	go embedding.UpdateUserEmbedding(db, &userEmbeddingFloat32, &announcementEmbeddingFloat32, userID)

	announcementData = append(announcementData, &pb.AnnouncementData{
		AuthorName:         authorName,
		AuthorID:           int32(authorID),
		Title:              title,
		Description:        description,
		Category:           category,
		LinkToAnnouncement: "/announcements?id=" + strconv.Itoa(int(announcementID)),
		AnnouncementID:     announcementID,
		Images:             images,
	})

	for _, a := range announcementData {
		fmt.Println("one an", &a)
	}

	return announcementData, nil
}

func (s *AnnouncementsServer) SearchAnnouncements(ctx context.Context, req *pb.SearchAnnouncementsRequest) (*pb.SearchAnnouncementsResponse, error) {
	data := &pb.SearchAnnouncementsResponse{}

	if req.AnnouncementID != 0 {
		userIDInt, err := strconv.Atoi(req.UserID)
		if err != nil {
			return nil, err
		}

		data.AnnouncementsData, err = getAnnouncementInfo(req.AnnouncementID, int32(userIDInt))
		if err != nil {
			return nil, err
		}

		return data, nil
	}

	query := "SELECT announcement_id, title, description, announcement_author_id, images_path[1], category FROM announcements"
	var snippets []string
	countArgs := 1
	args := []interface{}{}

	if req.SearchString != "" {
		snippets, args, countArgs = combineSQLSnippets(snippets, args, countArgs, whereSnippetSQLSearchString, req.SearchString)
	}

	if req.AuthorID != "" {
		snippets, args, countArgs = combineSQLSnippets(snippets, args, countArgs, whereSnippetSQLUserID, req.AuthorID)
	}

	if req.Category != "" {
		snippets, args, countArgs = combineSQLSnippets(snippets, args, countArgs, whereSnippetSQLCategory, req.Category)
	}

	if len(snippets) >= 1 {
		query += " WHERE " + strings.Join(snippets, " AND ")
	}

	if req.Orderby == "new" {
		query += snippetSQLOrderByCreateAt
	} else {
		query += fmt.Sprintf(snippetSQLOrderByEmbedding, countArgs)
		args = append(args, req.UserID)
		countArgs++
	}

	query += fmt.Sprintf(snippetSQLOffsetAndLimit, countArgs, countArgs+1)
	args = append(args, req.Offset*limitAnnouncementsToShowInt, limitAnnouncementsToShowInt)
	countArgs += 2

	announcements, err := db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer announcements.Close()

	var announcementID int
	var authorID int
	var authorName string
	var title string
	var description string
	var category string
	var firstImagesPath sql.NullString
	for announcements.Next() {
		if err := announcements.Scan(&announcementID, &title, &description, &authorID, &firstImagesPath, &category); err != nil {
			return nil, err
		}

		if err := db.QueryRow("SELECT name FROM users WHERE user_id = $1", authorID).Scan(&authorName); err != nil {
			return nil, err
		}

		if imagesServiceExternalConnections == "" {
			imagesServiceExternalConnections = os.Getenv("IMAGES_SERVICE_EXTERNAL_CONNECTIONS")
		}

		var firstImagesPathSlice []string
		if firstImagesPath.String != "" {
			firstImagesPathSlice = []string{imagesServiceExternalConnections + firstImagesPath.String}
		} else {
			firstImagesPathSlice = []string{imagesServiceExternalConnections + PathToDefaultImage}
		}

		if len(description) > 40 {
			description = description[:37] + "..."
		}

		data.AnnouncementsData = append(data.AnnouncementsData, &pb.AnnouncementData{
			AuthorName:         authorName,
			Title:              title,
			Description:        description,
			Category:           category,
			LinkToAnnouncement: "/announcements?id=" + strconv.Itoa(announcementID),
			Images:             firstImagesPathSlice,
			AnnouncementID:     int32(announcementID),
		})
	}

	return data, nil
}

func combineSQLSnippets(snippets []string, args []interface{}, countArgs int, template string, valueToInsert interface{}) ([]string, []interface{}, int) {
	numPlaceholders := strings.Count(template, "%d")
	formatArgs := make([]interface{}, numPlaceholders)
	for i := range formatArgs {
		formatArgs[i] = countArgs
	}

	snippets = append(snippets, fmt.Sprintf(template, formatArgs...))
	args = append(args, valueToInsert)
	countArgs++
	return snippets, args, countArgs
}
