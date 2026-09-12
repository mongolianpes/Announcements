package service

import (
	"context"
	"errors"

	"announcements/internal/embedding"
	pb "announcements/proto"
)

const (
	announcementTitleLengthLimit       = 10
	announcementDescriptionLengthLimit = 100
)

func (s *AnnouncementsServer) CreateAnnouncement(ctx context.Context, req *pb.CreateAnnouncementRequest) (*pb.CreateAnnouncementResponse, error) {
	if len(req.Title) > announcementTitleLengthLimit {
		return nil, errors.New("Название слишком длинное, используйте до 10 символов")
	}
	if len(req.Description) > announcementDescriptionLengthLimit {
		return nil, errors.New("Описание слишком длинное, используйте до 100 символов")
	}

	response, err := s.storage.CreateAnnouncement(ctx, req.Title, req.Description, req.Category, int(req.AuthorID))
	if err != nil {
		return nil, err
	}

	if err := embedding.InsertEmbedding(ctx, s.storage, int(response.AnnouncementID), req.Title+req.Description); err != nil {
		return nil, err
	}

	return response, nil
}

func (s *AnnouncementsServer) AddImages(ctx context.Context, req *pb.AddImagesRequest) (*pb.AddImagesResponse, error) {
	data, err := s.storage.AddImages(ctx, int(req.AnnouncementID), req.ImagesPath)
	if err != nil {
		return nil, err
	}

	return data, nil
}
