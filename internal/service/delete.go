package service

import (
	"context"

	pb "announcements/proto"
)

func (s *AnnouncementsServer) DeleteAnnouncement(ctx context.Context, req *pb.DeleteAnnouncementRequest) (*pb.DeleteAnnouncementResponse, error) {
	err := s.storage.UpdateTopPartnerEmbeddingBeforeDeleteAnnouncement(ctx, int(req.UserID), int(req.AnnouncementID))
	if err != nil {
		return nil, err
	}

	response, err := s.storage.DeleteAnnouncement(ctx, int(req.AnnouncementID), int(req.UserID))
	if err != nil {
		return nil, err
	}

	return response, nil
}
