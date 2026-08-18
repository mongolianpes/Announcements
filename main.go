package main

import (
	"log"
	"log/slog"
	"net"

	"google.golang.org/grpc"

	"announcements/announcements"
	"announcements/embedding"
	pb "announcements/proto"
)

func main() {
	go func() {
		if err := embedding.RunSetterDefaultEmbedding(); err != nil {
			slog.Warn("Ошибка при обновлении стандартного (популярного эмбеддинга)", "error", err)
		}
	}()

	announcements.ConnectToDB()

	lis, err := net.Listen("tcp", ":8086")
	if err != nil {
		log.Fatalf("не удалось слушать порт: %v", err)
	}

	grpcServer := grpc.NewServer()
	pb.RegisterAnnouncementsServer(grpcServer, &announcements.AnnouncementsServer{})

	log.Println("gRPC сервер запущен на :8086")
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("ошибка сервера: %v", err)
	}
}
