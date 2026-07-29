package announcements

import (
	pb "announcements/proto"
	"context"
	"errors"
	"fmt"
)

func (s *AnnouncementsServer) DeleteAnnouncement(ctx context.Context, req *pb.DeleteAnnouncementRequest) (*pb.DeleteAnnouncementResponse, error) {
	fmt.Println("announcement id to del: ", req.AnnouncementID)
	if _, err := db.Exec("DELETE FROM announcements WHERE announcement_id = $1", req.AnnouncementID); err != nil {
		return &pb.DeleteAnnouncementResponse{
			Error: "Объявление не удалено",
		}, errors.New("Объявление не удалено")
	}

	return &pb.DeleteAnnouncementResponse{}, nil
}
