package service

import (
	"context"

	pb "announcements/proto"
)

var imagesServiceExternalConnections string

func (s *AnnouncementsServer) SearchAnnouncements(ctx context.Context, req *pb.SearchAnnouncementsRequest) (*pb.SearchAnnouncementsResponse, error) {
	if req.AnnouncementID != 0 {
		response, err := s.storage.GetAnnouncementInfo(ctx, req.AnnouncementID)
		if err != nil {
			return nil, err
		}

		return response, nil
	}

	response, err := s.storage.SearchAnnouncements(ctx, int(req.UserID), int(req.AuthorID), int(req.Offset), req.Category, req.SearchString, req.Orderby)
	if err != nil {
		return nil, err
	}

	return response, nil
}
