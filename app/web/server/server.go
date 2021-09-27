package server

import (
	"fmt"
	"net/http"

	echossesion "github.com/go-session/echo-session"
	"github.com/go-session/session"
	"github.com/labstack/echo"
)

type WebServer struct {
	address string

	echoServer *echo.Echo
}

func NewWebServer(root, host string, port uint16) *WebServer {
	server := &WebServer{
		fmt.Sprintf("%v:%v", host, port),
		echo.New(),
	}

	server.echoServer.Renderer = NewTemplateRenderer(root)
	server.echoServer.Use(echossesion.New())
	server.echoServer.Static("/static", "html/static")
	return server
}

func (s *WebServer) Serve() error {
	var err error
	err = s.echoServer.Start(s.address)

	if err != http.ErrServerClosed {
		return err
	}

	return nil
}

func (s *WebServer) GET(route string, handler echo.HandlerFunc, middleware ...echo.MiddlewareFunc) {
	s.echoServer.GET(route, handler, middleware...)
}

func (s *WebServer) Session(c echo.Context) session.Store {
	return echossesion.FromContext(c)
}
