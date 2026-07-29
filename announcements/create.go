package announcements

import (
	"io"

	"github.com/lib/pq"

	"announcements/embedding"
	pb "announcements/proto"
)

func (s *AnnouncementsServer) CreateAnnouncement(stream pb.Announcements_CreateAnnouncementServer) error {
	data := &pb.CreateAnnouncementResponse{}

	var announcementID int
	for {
		req, err := stream.Recv()
		if err == io.EOF {
			return stream.Send(&pb.CreateAnnouncementResponse{})
		}
		if err != nil {
			return err
		}

		if len(req.ImagesPath) >= 1 {
			if _, err := db.Exec("UPDATE announcements SET images_path = $1 WHERE announcement_id = $2", req.ImagesPath, announcementID); err != nil {
				data.Error = "Картинки не сохранены"
			} else {
				data.Error = ""
			}
			return stream.Send(data)
		} else {
			if err := db.QueryRow("INSERT INTO announcements (title, description, category, announcement_author_id, images_path) VALUES ($1, $2, $3, $4, $5) RETURNING announcement_id", req.Title, req.Description, req.Category, req.AuthorID, req.ImagesPath).Scan(&announcementID); err != nil {
				if pqErr, ok := err.(*pq.Error); ok {
					switch pqErr.Code {
					case "22001":
						data.Error = "Название или описание слишком длинное. Для нзвания используйте до 10 символов, для описания до 100"
					default:
						data.Error = "Неизвестная ошибка, попробуйте позже"
					}
				} else {
					data.Error = "Неизвестная ошибка, попробуйте позже"
				}
			} else {
				data.Error = ""
			}

			go embedding.InsertEmbedding(db, announcementID, req.Title+req.Description, "UPDATE announcements SET embedding = $1::float8[] WHERE announcement_id = $1")

			stream.Send(data)
		}
	}
}
