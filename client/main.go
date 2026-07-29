package main

import (
	"client/announcements"
	"fmt"
)

func main() {
	if err := announcements.InitService(); err != nil {
		fmt.Println(err.Error())
	}

	if err := announcements.CreateAnnouncement("1", "2", "Участки", "24", []string{}); err != nil {
		fmt.Println(err.Error())
	}

	if err := announcements.DeleteAnnouncement(2); err != nil {
		fmt.Println(err.Error())
	}
}
