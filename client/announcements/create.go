package announcements

import (
	"context"
	"errors"

	pb "client/proto"
)

func CreateAnnouncement(title, description, category, authorID string, imagesPath []string) error {
	stream, err := client.CreateAnnouncement(context.Background())
	if err != nil {
		return err
	}

	req := &pb.CreateAnnouncementRequest{
		Title:       title,
		Description: description,
		Category:    category,
		AuthorID:    authorID,
	}

	if err := stream.Send(req); err != nil {
		return err
	}

	resp, err := stream.Recv()
	if err != nil {
		return err
	}
	if resp.Error != "" {
		return errors.New(resp.Error)
	}

	if len(imagesPath) >= 1 {
		if err := stream.Send(&pb.CreateAnnouncementRequest{
			ImagesPath: imagesPath,
		}); err != nil {
			return err
		}
	}

	if err := stream.CloseSend(); err != nil {
		return err
	}

	final, err := stream.Recv()
	if err != nil {
		return err
	}
	if final.Error != "" {
		return errors.New(final.Error)
	}

	return nil
}
