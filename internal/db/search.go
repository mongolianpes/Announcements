package db

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

	pb "announcements/proto"
)

const (
	limitAnnouncementsToShowInt = 5
	whereSnippetSQLSearchString = ` (title ILIKE '%%' || $%d || '%%'
		OR description ILIKE '%%' || $%d || '%%'
		OR SIMILARITY(title, $%d) > 0.3
		OR SIMILARITY(description, $%d) > 0.3)`
	whereSnippetSQLUserID                     = " (announcement_author_id = $%d)"
	whereSnippetSQLCategory                   = " (category = $%d)"
	snippetSQLOrderByCreateAt                 = " ORDER BY create_at DESC"
	snippetSQLOrderByEmbedding                = " ORDER BY embedding <=> (SELECT embedding FROM users WHERE user_id = $%d)"
	snippetSQLOrderByMyAnnouncementsEmbedding = " ORDER BY embedding <=> (SELECT AVG(embedding)::vector(768) FROM announcements WHERE announcement_author_id = $%d)"
	snippetSQLOffsetAndLimit                  = " OFFSET $%d LIMIT $%d"
)

const pathToDefaultImage = "d.webp"

var imagesServiceExternalConnections string

func (s *PostgresStorage) GetAnnouncementInfo(ctx context.Context, announcementID int64) (*pb.SearchAnnouncementsResponse, error) {
	var authorID int
	var authorName string
	var userEmbedding pgvector.Vector
	var title string
	var description string
	var images []string
	var category string
	sqlRow := s.db.QueryRow("SELECT title, description, announcement_author_id, images_path, category FROM announcements WHERE announcement_id = $1", announcementID)
	if err := sqlRow.Scan(&title, &description, &authorID, pq.Array(&images), &category); err != nil {
		return nil, errors.New("Нет объявления с таким id")
	}

	if err := s.db.QueryRow("SELECT name, embedding FROM users WHERE user_id = $1", authorID).Scan(&authorName, &userEmbedding); err != nil {
		authorName = "Неизвестно"
	}

	announcementData := []*pb.AnnouncementData{}
	announcementData = append(announcementData, &pb.AnnouncementData{
		AuthorID:           int64(authorID),
		Title:              title,
		Description:        description,
		Category:           category,
		LinkToAnnouncement: "/announcements?id=" + strconv.Itoa(int(announcementID)),
		AnnouncementID:     announcementID,
		Images:             images,
	})

	data := &pb.SearchAnnouncementsResponse{
		AnnouncementsData: announcementData,
	}

	return data, nil
}

func (s *PostgresStorage) SearchAnnouncements(ctx context.Context, userID, searchAuthorID, offset int, searchCategory, searchString, orderBy string) (*pb.SearchAnnouncementsResponse, error) {
	query := "SELECT announcement_id, title, description, announcement_author_id, images_path[1], category FROM announcements"
	var snippets []string
	countArgs := 1
	args := []interface{}{}

	if searchString != "" {
		snippets, args, countArgs = combineSQLSnippets(snippets, args, countArgs, whereSnippetSQLSearchString, searchString)
	}

	if searchAuthorID != 0 {
		snippets, args, countArgs = combineSQLSnippets(snippets, args, countArgs, whereSnippetSQLUserID, searchAuthorID)
	}

	if searchCategory != "" {
		snippets, args, countArgs = combineSQLSnippets(snippets, args, countArgs, whereSnippetSQLCategory, searchCategory)
	}

	if len(snippets) >= 1 {
		query += " WHERE " + strings.Join(snippets, " AND ")
	}

	if orderBy == "new" {
		query += snippetSQLOrderByCreateAt
	} else if orderBy == "my" {
		query += fmt.Sprintf(snippetSQLOrderByMyAnnouncementsEmbedding, countArgs)
		args = append(args, userID)
		countArgs++
	} else {
		query += fmt.Sprintf(snippetSQLOrderByEmbedding, countArgs)
		args = append(args, userID)
		countArgs++
	}

	query += fmt.Sprintf(snippetSQLOffsetAndLimit, countArgs, countArgs+1)
	args = append(args, offset*limitAnnouncementsToShowInt, limitAnnouncementsToShowInt)
	countArgs += 2

	announcements, err := s.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer announcements.Close()

	data := &pb.SearchAnnouncementsResponse{}
	var announcementID int
	var authorID int
	var title string
	var description string
	var category string
	var firstImagesPath sql.NullString
	for announcements.Next() {
		if err := announcements.Scan(&announcementID, &title, &description, &authorID, &firstImagesPath, &category); err != nil {
			return nil, err
		}

		if imagesServiceExternalConnections == "" {
			imagesServiceExternalConnections = os.Getenv("IMAGES_SERVICE_EXTERNAL_CONNECTIONS")
		}

		var firstImagesPathSlice []string
		if firstImagesPath.String != "" {
			firstImagesPathSlice = []string{imagesServiceExternalConnections + firstImagesPath.String}
		} else {
			firstImagesPathSlice = []string{imagesServiceExternalConnections + pathToDefaultImage}
		}

		if len(description) > 40 {
			description = description[:37] + "..."
		}

		data.AnnouncementsData = append(data.AnnouncementsData, &pb.AnnouncementData{
			AuthorID:           int64(authorID),
			Title:              title,
			Description:        description,
			Category:           category,
			LinkToAnnouncement: "/announcements?id=" + strconv.Itoa(announcementID),
			Images:             firstImagesPathSlice,
			AnnouncementID:     int64(announcementID),
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
