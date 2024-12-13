package hserver

import (
	"bytes"
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
	var d []byte
	result := bytes.NewBuffer(d)
	io.Copy(result, c.Request.Body)
	s.sender <- result.Bytes()
	c.String(http.StatusOK, "Ok")
}
