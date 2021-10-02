package web

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"

	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
	"github.com/tomwhy/PlaylistMusicDownloader/apis/youtube"
	youtubeAuth "github.com/tomwhy/PlaylistMusicDownloader/apis/youtube/auth"
	youtubeModel "github.com/tomwhy/PlaylistMusicDownloader/apis/youtube/model"
	audiodownloader "github.com/tomwhy/PlaylistMusicDownloader/internal/audioDownloader"
	"golang.org/x/oauth2"
	"google.golang.org/api/option"
	youtubeapi "google.golang.org/api/youtube/v3"

	"github.com/tomwhy/PlaylistMusicDownloader/app"
	"github.com/tomwhy/PlaylistMusicDownloader/app/web/model"
	"github.com/tomwhy/PlaylistMusicDownloader/app/web/server"

	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
)

type WebApp struct {
	server *server.WebServer

	googleAuthorizer youtubeAuth.Authorizer

	websocketUpgrader websocket.Upgrader
}

func NewWebApp() app.App {
	app := &WebApp{
		server: server.NewWebServer("./html", "", os.Getenv("PORT")),
		googleAuthorizer: youtubeAuth.NewAuthorizer(os.Getenv("CLIENT_ID"),
			os.Getenv("CLIENT_SECRET"),
			fmt.Sprintf("https://%v/authCallback", os.Getenv("HOST")),
			[]string{youtubeapi.YoutubeReadonlyScope}),
		websocketUpgrader: websocket.Upgrader{},
	}

	app.server.Use(app.httpsRedirect)
	app.server.Use(middleware.Recover())

	app.server.GET("/", app.home)
	app.server.GET("/auth", app.authentication)
	app.server.GET("/authCallback", app.authenticateCallback)
	app.server.GET("/revoke", app.logout, app.authMiddleware)
	app.server.GET("/download/:id", app.downloadPage, app.authMiddleware)

	app.server.GET("/api/download/:id", app.downloadPlaylistSongs, app.authMiddleware)
	app.server.GET("/api/download/song/:id", app.downloadSong, app.authMiddleware)
	return app
}

func (app *WebApp) httpsRedirect(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		originalProto := c.Request().Header.Get("x-forwarded-proto")
		if originalProto != "https" {
			return c.Redirect(http.StatusTemporaryRedirect, "https://"+c.Request().Host+c.Request().RequestURI)
		}

		return next(c)
	}
}

func (app *WebApp) authMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		if !app.isLoggedIn(c) {
			return c.Redirect(http.StatusTemporaryRedirect, "/")
		}

		return next(c)
	}
}

func (app *WebApp) authentication(c echo.Context) error {
	session := app.server.Session(c)
	authentication_url, state, err := app.googleAuthorizer.GetAuthURL()
	if err != nil {
		logrus.Error("Failed getting Authentication url", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed authenticating")
	}

	session.Set("state", state)
	session.Save()

	return c.Redirect(http.StatusTemporaryRedirect, authentication_url)
}

func (app *WebApp) authenticateCallback(c echo.Context) error {
	session := app.server.Session(c)

	if state, ok := session.Get("state"); !ok || c.QueryParam("state") != state {
		session.Delete("state")
		session.Save()
		logrus.Info("Got invalid state")
		return echo.NewHTTPError(http.StatusUnauthorized, "Failed authentication")
	}

	token, err := app.googleAuthorizer.GetToken(c.QueryParam("code"))
	if err != nil {
		logrus.Error("Failed getting token.", err)
		return echo.NewHTTPError(http.StatusUnauthorized, "Failed authentication")
	}

	session.Set("token", *token)
	session.Save()

	return c.Redirect(http.StatusFound, "/")
}

func (app *WebApp) home(c echo.Context) error {
	if app.isLoggedIn(c) {
		return app.loggedInHome(c)
	} else {
		return app.loggedOutHome(c)
	}
}

func (app *WebApp) loggedOutHome(c echo.Context) error {
	return RenderHTML(c, "./html/loggedOutHome.html")
}

func (app *WebApp) loggedInHome(c echo.Context) error {
	api := app.youtubeAPI(c)
	page := c.QueryParam("page")

	playlists, nextPage, prevPage, err := api.GetAllPlaylists(page, 50)
	if err != nil {
		logrus.Error("Failed getting playlists.", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed getting playlists")
	}

	params := struct {
		Playlists []youtubeModel.YoutubePlaylist
		Next      string
		Prev      string
	}{playlists, nextPage, prevPage}

	return c.Render(http.StatusOK, "loggedInHome.html", params)
}

func (app *WebApp) logout(c echo.Context) error {
	session := app.server.Session(c)
	token, _ := session.Get("token")

	urlParams := url.Values{}
	urlParams.Add("token", token.(oauth2.Token).AccessToken)
	revokeURL := "https://oauth2.googleapis.com/revoke?" + urlParams.Encode()

	response, err := http.PostForm(revokeURL, url.Values{})
	if err != nil {
		logrus.Error("Failed revoking token.", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed logging out")
	}
	defer response.Body.Close()

	session.Delete("token")
	session.Save()
	return c.Redirect(http.StatusTemporaryRedirect, "/")
}

func (app *WebApp) downloadPage(c echo.Context) error {
	return c.Render(http.StatusOK, "download.html", c.Param("id"))
}

func (app *WebApp) downloadPlaylistSongs(c echo.Context) error {
	ws, err := app.websocketUpgrader.Upgrade(c.Response(), c.Request(), nil)
	if err != nil {
		logrus.Error("Failed upgrading to websocket")
		return err
	}
	defer ws.Close()

	songsContext, cancelSongs := context.WithCancel(context.Background())
	songs, errChan := app.youtubeAPI(c).GetPlaylistSongs(songsContext, c.Param("id"))
	for song := range songs {
		logrus.Info("Getting urls for: ", song.Title)

		if err != nil {
			logrus.Error("Failed getting songs for playlist: ", c.Param("id"), ".", err)
			ws.WriteJSON(model.WebsocketMessage{MessageType: "error", Data: err.Error()})
			break
		}

		song.DownloadUrl = "/api/download/song/" + song.Id
		ws.WriteJSON(model.WebsocketMessage{MessageType: "song", Data: song})
	}

	cancelSongs()
	err = <-errChan
	if err != nil {
		ws.WriteJSON(model.WebsocketMessage{MessageType: "error", Data: err.Error()})
	}

	return err
}

func (app *WebApp) downloadSong(c echo.Context) error {
	videoId := c.Param("id")

	stream, filename, err := audiodownloader.DownloadAudio(videoId)
	if err != nil {
		logrus.Error("Failed to download song: ", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Could not download song")
	}

	c.Response().Header().Add("Content-Disposition", fmt.Sprintf("attachment; filename="+filename))
	return c.Stream(http.StatusOK, echo.MIMEOctetStream, stream)
}

func (app *WebApp) isLoggedIn(c echo.Context) bool {
	session := app.server.Session(c)
	_, ok := session.Get("token")
	return ok
}

func (app *WebApp) Run() error {
	return app.server.Serve()
}

func (app *WebApp) youtubeAPI(c echo.Context) *youtube.YoutubeAPI {
	session := app.server.Session(c)
	token, _ := session.Get("token")
	client := app.googleAuthorizer.CreateClient(token.(oauth2.Token))
	return youtube.NewYoutubeAPI(option.WithHTTPClient(client))
}

func RenderHTML(c echo.Context, filepath string) error {
	file, err := ioutil.ReadFile(filepath)
	if err != nil {
		logrus.Error("Could not read file: ", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Could not find file")
	}
	return c.HTMLBlob(http.StatusOK, file)
}
