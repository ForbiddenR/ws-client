package hserver

import (
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
)

type Server struct {
	addr   string
	sender chan<- []byte
}

func NewServer(addr string, sender chan<- []byte) *Server {
	return &Server{
		addr:   addr,
		sender: sender,
	}
}

func (s *Server) Start() error {
	engine := gin.New()
	engine.Use(gin.Recovery())
	engine.POST("/send", s.send)
	return engine.Run(s.addr)
}

func (s *Server) send(c *gin.Context) {
	result, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}
	s.sender <- result
	c.String(http.StatusOK, "Ok")
}
