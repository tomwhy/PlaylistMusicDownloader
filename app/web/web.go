package web

import (
	"io/ioutil"
	"net/http"
	"net/url"
	"os"

	"github.com/tomwhy/PlaylistMusicDownloader/apis/youtube"
	youtubeAuth "github.com/tomwhy/PlaylistMusicDownloader/apis/youtube/auth"
	youtubeModel "github.com/tomwhy/PlaylistMusicDownloader/apis/youtube/model"
	"golang.org/x/oauth2"
	"google.golang.org/api/option"
	youtubeapi "google.golang.org/api/youtube/v3"

	"github.com/tomwhy/PlaylistMusicDownloader/app"
	"github.com/tomwhy/PlaylistMusicDownloader/app/web/server"

	"github.com/labstack/echo"
)

type WebApp struct {
	server *server.WebServer

	googleAuthorizer youtubeAuth.Authorizer
}

func NewWebApp() app.App {
	app := &WebApp{
		server:           server.NewWebServer("html", "", 80),
		googleAuthorizer: youtubeAuth.NewAuthorizer(os.Getenv("CLIENT_ID"), os.Getenv("CLIENT_SECRET"), "http://localhost/authCallback", []string{youtubeapi.YoutubeReadonlyScope}),
	}

	app.server.GET("/", app.home)
	app.server.GET("/auth", app.authentication)
	app.server.GET("/authCallback", app.authenticateCallback)
	app.server.GET("/revoke", app.logout, app.authMiddleware)
	app.server.GET("/download/:id", app.home, app.authMiddleware)

	return app
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
		return echo.NewHTTPError(http.StatusInternalServerError, err)
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
		return echo.NewHTTPError(http.StatusUnauthorized, "Failed authentication")
	}

	token, err := app.googleAuthorizer.GetToken(c.QueryParam("code"))
	if err != nil {
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
	file, _ := ioutil.ReadFile("html/loggedOutHome.html")
	return c.HTMLBlob(http.StatusOK, file)
}

func (app *WebApp) loggedInHome(c echo.Context) error {
	api := app.youtubeAPI(c)
	page := c.QueryParam("page")

	playlists, nextPage, prevPage, err := api.GetAllPlaylists(page, 50)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err)
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

	_, err := http.PostForm(revokeURL, url.Values{})
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed logging out")
	}

	session.Delete("token")
	session.Save()
	return c.Redirect(http.StatusTemporaryRedirect, "/")
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
