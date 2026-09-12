package main

import (
	"log"
	"net"

	"google.golang.org/grpc"

	"announcements/internal/db"
	"announcements/internal/service"
	pb "announcements/proto"
)

func main() {
	storage, err := db.NewPostgresStorage()
	if err != nil {
		panic(err)
	}

	lis, err := net.Listen("tcp", ":8086")
	if err != nil {
		log.Fatalf("не удалось слушать порт: %v", err)
	}

	grpcServer := grpc.NewServer()
	pb.RegisterAnnouncementsServer(grpcServer, service.NewAnnouncementServer(storage))

	log.Println("gRPC сервер запущен на :8086")
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("ошибка сервера: %v", err)
	}
}
