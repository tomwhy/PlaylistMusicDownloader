package main

import (
	"github.com/sirupsen/logrus"
	"github.com/tomwhy/PlaylistMusicDownloader/app/web"
)

func main() {
	app := web.NewWebApp()

	err := app.Run()
	if err != nil {
		logrus.Fatal(err)
	}
}
