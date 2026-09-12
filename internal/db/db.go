package db

import (
	"context"

	pb "announcements/proto"
)

const (
	createAnnouncement      = "INSERT INTO announcements (title, description, category, announcement_author_id) VALUES ($1, $2, $3, $4) RETURNING announcement_id"
	addImagesInAnnoucements = "UPDATE announcements SET images_path = $1 WHERE announcement_id = $2"
	deleteAnnouncement      = "DELETE FROM announcements WHERE announcement_id = $1 AND announcement_author_id = $2"
)

func (s *PostgresStorage) CreateAnnouncement(ctx context.Context, title, description, category string, authorID int) (*pb.CreateAnnouncementResponse, error) {
	data := &pb.CreateAnnouncementResponse{}
	if err := s.db.QueryRow(createAnnouncement, title, description, category, authorID).Scan(&data.AnnouncementID); err != nil {
		return nil, err
	}

	return data, nil
}

func (s *PostgresStorage) AddImages(ctx context.Context, announcementID int, imagesPath []string) (*pb.AddImagesResponse, error) {
	data := &pb.AddImagesResponse{}
	if _, err := s.db.Exec(addImagesInAnnoucements, imagesPath, announcementID); err != nil {
		return nil, err
	}
	return data, nil
}

func (s *PostgresStorage) DeleteAnnouncement(ctx context.Context, announcementID, userID int) (*pb.DeleteAnnouncementResponse, error) {
	data := &pb.DeleteAnnouncementResponse{}
	if _, err := s.db.Exec(deleteAnnouncement, announcementID, userID); err != nil {
		return nil, err
	}
	return data, nil
}
