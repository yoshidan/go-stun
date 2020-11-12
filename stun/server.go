package stun

import (
	"context"
	"log"
	"net"
	"sync"
	"syscall"
	"time"

	"golang.org/x/xerrors"
)

type Server struct {
	serverAddr string
	conn       *net.UDPConn
	context    context.Context
	cancel     context.CancelFunc
}

func NewServer(ctx context.Context, serverAddr string) Server {
	c, cancel := context.WithCancel(ctx)
	return Server{
		serverAddr: serverAddr,
		context:    c,
		cancel:     cancel,
	}
}

func (s *Server) Shutdown() {
	s.cancel()
	err := s.conn.SetReadDeadline(time.Now().Add(1 * time.Millisecond))
	if err != nil {
		log.Fatalf("%+v", err)
	}
}

func (s *Server) ListenAndServe() error {
	listenConfig := &net.ListenConfig{
		Control: func(network, address string, c syscall.RawConn) (err error) {
			return c.Control(func(fd uintptr) {
				err = syscall.SetsockoptInt(int(fd), syscall.SOL_SOCKET, syscall.SO_REUSEADDR, 1)
				if err != nil {
					return
				}
				err = syscall.SetsockoptInt(int(fd), syscall.SOL_SOCKET, syscall.SO_REUSEPORT, 1)
			})
		},
	}
	ln, err := listenConfig.ListenPacket(s.context, "udp", s.serverAddr)
	if err != nil {
		return xerrors.Errorf("listening error: %w", err)
	}
	defer ln.Close()
	s.conn = ln.(*net.UDPConn)

	log.Printf("Server is litening on %s", s.serverAddr)

	wg := sync.WaitGroup{}
	for {
		select {
		case <-s.context.Done():
			log.Println("Waiting for all the process finished...")
			// Graceful Shutdown.
			wg.Wait()
			return nil
		default:
			buf := make([]byte, MessageHeaderLength)
			n, addr, err := s.conn.ReadFromUDP(buf)
			if err != nil {
				log.Printf("%+v", xerrors.Errorf("receive failed: %w", err))
				continue
			}
			// STUN header only message is supported.
			if n != MessageHeaderLength {
				log.Println("illegal message")
				continue
			}
			log.Print("receive from : " + addr.String())

			wg.Add(1)
			go func() {
				defer wg.Done()
				header := NewMessage(buf[:n])
				err := header.ValidateAsBindingRequest()
				if err != nil {
					log.Println(err)
					writeBuf := header.CreateErrorResponse()
					s.send(writeBuf, addr)
				} else {
					writeBuf := header.CreateSuccessResponse(uint16(addr.Port), addr.IP)
					s.send(writeBuf, addr)
				}
			}()
		}
	}
}

func (s *Server) send(writeBuf []byte, addr *net.UDPAddr) {
	_, err := s.conn.WriteTo(writeBuf, addr)
	if err != nil {
		log.Println(err)
	}
}
