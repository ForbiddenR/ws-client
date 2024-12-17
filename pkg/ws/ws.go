package ws

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"log/slog"
	"net/http/httputil"
	"os"
	"time"

	"github.com/gorilla/websocket"
	"golang.org/x/sync/errgroup"
)

var log *slog.Logger

type Server struct {
	addresss string
	certPath string
	keyPath  string
	caPath   string
	mtls     bool
	sn       string
	sender   <-chan []byte
}

func NewServer(address, certPath, keyPath, caPath string, mtls bool, sn string, sender <-chan []byte) *Server {
	return &Server{
		addresss: address,
		certPath: certPath,
		keyPath:  keyPath,
		caPath:   caPath,
		mtls:     mtls,
		sn:       sn,
		sender:   sender,
	}
}

func (s *Server) Start(ctx context.Context) error {
	file, err := os.OpenFile("log.log", os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		panic(err)
	}
	defer file.Close()
	log = slog.New(slog.NewTextHandler(file, nil))

	tlsConfig := &tls.Config{
		InsecureSkipVerify: true,
		MinVersion:         tls.VersionTLS12,
	}

	if s.caPath != "" {
		caCert, err := os.ReadFile(s.caPath)
		if err != nil {
			panic(err)
		}
		caCertPool := x509.NewCertPool()
		caCertPool.AppendCertsFromPEM(caCert)
		tlsConfig.RootCAs = caCertPool
	}

	if s.mtls {
		cert, err := s.loadKeyPair()
		if err != nil {
			panic(err)
		}
		tlsConfig.Certificates = []tls.Certificate{cert}
	}

	dialer := websocket.Dialer{
		TLSClientConfig: tlsConfig,
		Proxy:           websocket.DefaultDialer.Proxy,
		Subprotocols:    []string{"ocpp2.0.1"},
	}
	conn, resp, err := dialer.Dial(s.addresss, nil)
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	conn.SetPongHandler(func(appData string) error {
		conn.SetReadDeadline(time.Now().Add(time.Second * 70))
		log.Info(fmt.Sprintf("<- pong {%s}", appData))
		return nil
	})

	responseDump, err := httputil.DumpResponse(resp, false)
	if err != nil {
		panic(err)
	}
	fmt.Println(string(responseDump))

	eg := errgroup.Group{}

	eg.Go(func() error {
		buffer := make([]byte, 10240)
		for {
			select {
			case <-ctx.Done():
				return ctx.Err()
			default:
			}
			mt, reader, err := conn.NextReader()
			if err != nil {
				return err
			}
			switch mt {
			case websocket.TextMessage:
				n, err := reader.Read(buffer)
				if err != nil {
					return err
				}
				log.Info(fmt.Sprintf(`<-%s. %s`, buffer[:n], s.sn))
			default:
				fmt.Println("bad message type", mt)
			}
		}
	})

	eg.Go(func() error {
		ticker := time.NewTicker(time.Second * 60)
		for {
			select {
			case <-ticker.C:
				pm := fmt.Sprintf("%s ping", s.sn)
				err := conn.WriteMessage(websocket.PingMessage, []byte(pm))
				if err != nil {
					return err
				}
				log.Info(fmt.Sprintf("-> ping {%s}", pm))
			case pm := <-s.sender:
				log.Info(fmt.Sprintf("->%s. %s", pm, s.sn))
				err := conn.WriteMessage(websocket.TextMessage, pm)
				if err != nil {
					return err
				}
			case <-ctx.Done():
				return ctx.Err()
			}
		}
	})
	return eg.Wait()
}

func (s *Server) loadKeyPair() (tls.Certificate, error) {
	return tls.LoadX509KeyPair(s.certPath, s.keyPath)
}
