package server

import (
	"log"
	"net"

	"github.com/cizzle-cloud/cloud-gateway/internal/router"
)

type Protocol string

const (
	HTTP Protocol = "http"
	GRPC Protocol = "grpc"
	TCP  Protocol = "tcp"
)

type Server struct {
	httpRouter router.HTTPRouter
	grpcRouter router.GRPCRouter
	tcpRouter  router.TCPRouter
}

func New() *Server {
	return &Server{}
}

func (s *Server) SetHTTPRouter(router router.HTTPRouter) {
	s.httpRouter = router
}

func (s *Server) SetGRPCRouter(router router.GRPCRouter) {
	s.grpcRouter = router
}

func (s *Server) SetTCPRouter(router router.TCPRouter) {
	s.tcpRouter = router
}

// Start starts the server and begins accepting connections
func (s *Server) Start() {
	s.handleConnection(&net.TCPConn{})
}

// handleConnection handles an incoming connection by detecting protocol and delegating
func (s *Server) handleConnection(conn net.Conn) {
	protocol, err := s.detectProtocol(conn)
	if err != nil {
		log.Printf("[SERVER] Error detecting protocol: %v", err)
		return
	}

	log.Println(protocol)

	// Delegate to appropriate handler
	switch protocol {
	case HTTP:
		s.serveHTTP(conn)
	case GRPC:
		s.serveGRPC(conn)
	case TCP:
		s.serveTCP(conn)
	default:
		log.Printf("[SERVER] Unsupported protocol: %s", protocol)
	}
}

// detectProtocol analyzes the first bytes of the connection to determine protocol
func (s *Server) detectProtocol(conn net.Conn) (Protocol, error) {
	log.Println(conn)
	return HTTP, nil
}

func (s *Server) serveHTTP(conn net.Conn) {
	log.Println(conn)
	// s.httpRouter.ServeHTTP(conn, &http.Request{})
}
func (s *Server) serveGRPC(conn net.Conn) {
	log.Println(conn)
	s.grpcRouter.ServeGRPC()
}
func (s *Server) serveTCP(conn net.Conn) {
	log.Println(conn)
	s.tcpRouter.ServeTCP()
}
