package db

import (
	"context"
	"database/sql"
	"fmt"
	"os"

	pb "announcements/proto"
)

type PostgresStorage struct {
	db *sql.DB
}

type AnnouncementsStorage interface {
	CreateAnnouncement(ctx context.Context, title, description, category string, authorID int) (*pb.CreateAnnouncementResponse, error)
	AddImages(ctx context.Context, announcementID int, imagesPath []string) (*pb.AddImagesResponse, error)
	DeleteAnnouncement(ctx context.Context, announcementID, userID int) (*pb.DeleteAnnouncementResponse, error)
	GetAnnouncementInfo(ctx context.Context, announcementID int64) (*pb.SearchAnnouncementsResponse, error)
	SearchAnnouncements(ctx context.Context, userID, searchAuthorID, offset int, searchCategory, searchString, orderBy string) (*pb.SearchAnnouncementsResponse, error)
	UpdateTopPartnerEmbeddingBeforeDeleteAnnouncement(ctx context.Context, authorID, announcementID int) error
	InsertEmbedding(ctx context.Context, rowID int, embedding []float64) error
	SaveEmbeddingText(ctx context.Context, rowID int, text string) error
	GetSavedEmbeddings(ctx context.Context, offset, limit int) (map[int]string, error)
	DeleteSavedEmbedding(ctx context.Context, id int) error
}

func NewPostgresStorage() (*PostgresStorage, error) {
	db, err := connectToDB()
	if err != nil {
		return nil, err
	}
	return &PostgresStorage{
		db: db,
	}, nil
}

func connectToDB() (*sql.DB, error) {
	host := os.Getenv("DB_HOST")
	port := os.Getenv("DB_PORT")
	user := os.Getenv("DB_USER")
	password := os.Getenv("DB_PASSWORD")
	dbname := os.Getenv("DB_NAME")

	var err error
	psqlInfo := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable", host, port, user, password, dbname)
	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		return nil, err
	}

	err = db.Ping()
	if err != nil {
		return nil, err
	}

	return db, nil
}
