package db

import (
	"log"
	"net"
	"os"

	"gopkg.in/tomb.v2"
)

type Server struct {
	engine MongoEngine
	ln     net.Listener
	t      tomb.Tomb
}

func NewServerAddr(addr string, opt *MMOpt) (*Server, error) {
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, err
	}
	return NewServer(ln, opt), nil
}

func NewServer(ln net.Listener, opt *MMOpt) *Server {
	if opt == nil {
		opt = &MMOpt{
			engine: &Simple{collections: make(map[string]*MemoryCollection)},
		}
	}
	s := &Server{ln: ln, engine: opt.engine}
	return s
}

func (s *Server) Start() {
	s.t.Go(s.run)
	log.Printf("mmdb running pid=%d addr=%q", os.Getpid(), s.ln.Addr())
}

func (s *Server) Wait() error {
	return s.t.Wait()
}

func (s *Server) run() error {
	s.t.Go(func() (err error) {
		var conn net.Conn
		defer s.t.Kill(err)
		for {
			conn, err = s.ln.Accept()
			if err != nil {
				return
			}
			s.t.Go(func() error { s.handle(conn); return nil })
		}
	})
	<-s.t.Dying()
	s.ln.Close()
	return nil
}

func (s *Server) Stop() {
	//TODO stop
}

func (s *Server) handle(c net.Conn) {
	defer c.Close()
	for {
		select {
		case <-s.t.Dying():
			return
		default:
		}
		parse(c, c, s.engine)
	}
}
