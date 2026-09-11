package announcements

import (
	"context"
	"errors"

	"announcements/embedding"
	pb "announcements/proto"
)

func (s *AnnouncementsServer) DeleteAnnouncement(ctx context.Context, req *pb.DeleteAnnouncementRequest) (*pb.DeleteAnnouncementResponse, error) {
	if _, err := db.Exec("DELETE FROM announcements WHERE announcement_id = $1 AND announcement_author_id = $2", req.AnnouncementID, req.UserID); err != nil {
		return nil, errors.New("Объявление не удалено")
	}

	if err := embedding.UpdateUserEmbeddingAfterDeleteAnnouncement(db, req.UserID, req.AnnouncementID); err != nil {
		return nil, err
	}

	return &pb.DeleteAnnouncementResponse{}, nil
}
