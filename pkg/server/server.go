package server

import (
	"strings"
	"bytes"
	"io"
	"log"
	"net"
	"sync"
)

type HandlerFunc func(conn net.Conn)

type Server struct {
	addr string
	mu sync.RWMutex
	handlers map[string]HandlerFunc
} 

func NewServer(addr string) *Server {
	return &Server{addr: addr, handlers: make(map[string]HandlerFunc)}
} 

func (s *Server) Register(path string, handler HandlerFunc) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.handlers[path] = handler
}

func (s *Server) Start() (err error) {

	listener, err := net.Listen("tcp", s.addr)
	if err != nil {
		log.Println(err)
		return err
	}
	
	defer func() {
		if cerr := listener.Close(); cerr != nil {
			err = cerr
			return
		}
	}()

	for {
		conn, err := listener.Accept()
		if err != nil {
			continue
		}
		go s.handle(conn)
	}
}

func (s *Server) handle(conn net.Conn) {
	defer conn.Close()  
	
	buf := make([]byte, 4096)
	
	for {
	  	n, err := conn.Read(buf)
	  	if err == io.EOF {
			log.Printf("%s", buf[:n])
	  	}
	  	if err != nil {
			log.Println(err)
			return
	  	}

	  	data := buf[:n]
	  	requestLineDelim := []byte{'\r','\n'}
		
		requestLineEnd :=  bytes.Index(data, requestLineDelim)
		if requestLineEnd == -1 {
			return
		} 
		
		requestLine := string(data[:requestLineEnd])
		
		parts := strings.Split(requestLine," ")
		if len(parts) != 3 {
			return 
		}
		
		path, version := parts[1], parts[2]

		if version != "HTTP/1.1" {
			return
		}

		var handler = func(conn net.Conn) {
			conn.Close()
		}
		s.mu.RLock()
		for i := 0; i < len(s.handlers); i++ {
			if value, ok := s.handlers[path]; ok {
				handler = value
				break
			}
		}
		s.mu.RUnlock()
		
		handler(conn)
	}
}