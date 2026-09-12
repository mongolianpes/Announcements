package service

import (
	"announcements/internal/db"
	pb "announcements/proto"
)

type AnnouncementsServer struct {
	pb.UnimplementedAnnouncementsServer
	storage db.AnnouncementsStorage
}

func NewAnnouncementServer(storage db.AnnouncementsStorage) *AnnouncementsServer {
	return &AnnouncementsServer{
		storage: storage,
	}
}
