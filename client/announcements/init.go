package announcements

import (
	pb "client/proto"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func InitService() error {
	conn, err := grpc.NewClient("localhost:8086", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return err
	}

	client = pb.NewAnnouncementsClient(conn)
	return nil
}
