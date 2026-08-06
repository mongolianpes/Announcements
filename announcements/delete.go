package announcements

import (
	pb "announcements/proto"
	"context"
	"errors"
)

func (s *AnnouncementsServer) DeleteAnnouncement(ctx context.Context, req *pb.DeleteAnnouncementRequest) (*pb.DeleteAnnouncementResponse, error) {
	if _, err := db.Exec("DELETE FROM announcements WHERE announcement_id = $1", req.AnnouncementID); err != nil {
		return &pb.DeleteAnnouncementResponse{
			Error: "Объявление не удалено",
		}, errors.New("Объявление не удалено")
	}

	return &pb.DeleteAnnouncementResponse{}, nil
}
