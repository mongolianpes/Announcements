package announcements

import (
	"context"
	"errors"

	"github.com/lib/pq"

	"announcements/embedding"
	pb "announcements/proto"
)

func (s *AnnouncementsServer) CreateAnnouncement(ctx context.Context, req *pb.CreateAnnouncementRequest) (*pb.CreateAnnouncementResponse, error) {
	data := &pb.CreateAnnouncementResponse{}

	const queryAddAnnouncementInfo = "INSERT INTO announcements (title, description, category, announcement_author_id) VALUES ($1, $2, $3, $4) RETURNING announcement_id"
	if err := db.QueryRow(queryAddAnnouncementInfo, req.Title, req.Description, req.Category, req.AuthorID).Scan(&data.AnnouncementID); err != nil {
		var errCause string
		if pqErr, ok := err.(*pq.Error); ok {
			switch pqErr.Code {
			case "22001":
				errCause = "Название или описание слишком длинное. Для нзвания используйте до 10 символов, для описания до 100"
			default:
				errCause = "Неизвестная ошибка, попробуйте позже"
			}
		} else {
			errCause = "Неизвестная ошибка, попробуйте позже"
		}
		return nil, errors.New(errCause)
	}

	if err := embedding.InsertEmbedding(db, int(data.AnnouncementID), req.Title+req.Description); err != nil {
		return nil, err
	}

	return data, nil
}

func (s *AnnouncementsServer) AddImages(ctx context.Context, req *pb.AddImagesRequest) (*pb.AddImagesResponse, error) {
	data := &pb.AddImagesResponse{}

	if len(req.ImagesPath) > 0 {
		if _, err := db.Exec("UPDATE announcements SET images_path = $1 WHERE announcement_id = $2", req.ImagesPath, req.AnnouncementID); err != nil {
			return nil, err
		}
	} else {
		return nil, errors.New("Необходимо отправить хотя бы одно изображение")
	}

	return data, nil
}
