package announcements

import (
	"context"
	"fmt"

	pb "client/proto"
)

var client pb.AnnouncementsClient

func SearchAnnouncements(offset, announcementID int, userID, SearchString, login, category, orderBy string) error {
	stream, err := client.SearchAnnouncements(context.Background())
	if err != nil {
		return err
	}

	req := &pb.SearchAnnouncementsRequest{
		Offset:         int32(offset),
		UserID:         userID,
		SearchString:   SearchString,
		Login:          login,
		Category:       category,
		Orderby:        orderBy,
		AnnouncementID: int32(announcementID),
	}

	if err := stream.Send(req); err != nil {
		return err
	}

	resp, err := stream.CloseAndRecv()
	if err != nil {
		return err
	}

	fmt.Println(resp.Error)
	fmt.Println(resp.AnnouncementsData)
	return nil
}
