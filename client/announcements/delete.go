package announcements

import (
	"context"
	"fmt"

	pb "client/proto"
)

func DeleteAnnouncement(id int) error {
	stream, err := client.DeleteAnnouncement(context.Background())
	if err != nil {
		return err
	}

	req := &pb.DeleteAnnouncementRequest{
		AnnouncementID: int32(id),
	}

	if err := stream.Send(req); err != nil {
		return err
	}

	resp, err := stream.CloseAndRecv()
	if err != nil {
		return err
	}

	fmt.Println(resp)
	return nil
}
