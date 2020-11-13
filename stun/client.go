package stun

import (
	"context"
	"crypto/rand"
	"net"
	"time"

	"golang.org/x/xerrors"
)

type Client struct {
	Timeout time.Duration

	serverAddr string
	localAddr  *string

	conn    *net.UDPConn
	context context.Context
}

type Result struct {
	Error error
	Addr  net.UDPAddr
}

// serverAddr is STUN server address.
// If localAddr is nil, a local address is automatically chosen.
func NewClient(ctx context.Context, serverAddr string, localAddr *string) Client {
	return Client{
		serverAddr: serverAddr,
		context:    ctx,
		localAddr:  localAddr,
		Timeout:    1 * time.Second,
	}
}

// Keepalive stun server connection
func (c *Client) Keepalive() error {
	conn, err := c.dial()
	if err != nil {
		return err
	}
	c.conn = conn
	return nil
}

func (c *Client) Close() {
	if c.conn != nil {
		c.conn.Close()
	}
}

// Discover external ip address and port
func (c *Client) Discover() (net.UDPAddr, error) {
	conn := c.conn
	if conn == nil {
		var err error
		conn, err = c.dial()
		if err != nil {
			return net.UDPAddr{}, err
		}
		defer conn.Close()
	}
	data := c.createBindingRequest()
	var txID [TransactionIDLength]byte
	copy(txID[:], data[8:MessageHeaderLength])

	ch := make(chan Result)
	go func() {
		buf := make([]byte, 44)
		err := conn.SetReadDeadline(time.Now().Add(c.Timeout))
		if err != nil {
			ch <- c.error(xerrors.Errorf("dead line setting error: %w", err))
			return
		}
		size, _, err := conn.ReadFromUDP(buf)
		if err != nil {
			ch <- c.error(xerrors.Errorf("read error: %w", err))
			return
		}
		if size < MessageHeaderLength {
			ch <- c.error(xerrors.New("invalid response"))
			return
		}
		result := NewMessage(buf)
		if result.Type == BindingErrorResponse {
			ch <- c.error(xerrors.New("bad request"))
			return
		}
		if result.Type != BindingSuccessResponse {
			ch <- c.error(xerrors.New("type must be binding success response"))
			return
		}
		if result.MagicCookie != MagicCookie {
			ch <- c.error(xerrors.New("invalid magic cookie"))
			return
		}
		if result.TransactionID != txID {
			ch <- c.error(xerrors.New("invalid transaction id"))
			return
		}
		if nbo.Uint16(buf[20:22]) != AttributeXorMappedAddress {
			ch <- c.error(xerrors.New("first address must be xor mapped address"))
			return
		}
		port := nbo.Uint16(buf[26:28]) ^ uint16(MagicCookie>>16)

		if buf[25] == FamilyV4 {
			ch <- Result{nil, net.UDPAddr{IP: c.parseV4(buf[28:32]), Port: int(port)}}
		} else if buf[25] == FamilyV6 {
			ch <- Result{nil, net.UDPAddr{IP: c.parseV6(buf[28:44], txID), Port: int(port)}}
		} else {
			ch <- c.error(xerrors.New("invalid address family"))
		}
	}()

	_, err := conn.Write(data)
	if err != nil {
		_ = conn.SetReadDeadline(time.Now().Add(1 * time.Millisecond))
	}
	result := <-ch
	if err != nil {
		// write error
		return net.UDPAddr{}, xerrors.Errorf("sending error: %w", err)
	}
	return result.Addr, result.Error
}

func (c *Client) dial() (*net.UDPConn, error) {
	var localAddr *net.UDPAddr
	if c.localAddr != nil {
		var err error
		localAddr, err = net.ResolveUDPAddr("udp", *c.localAddr)
		if err != nil {
			return nil, xerrors.Errorf("resolve error: %w", err)
		}
	}
	serverAddr, err := net.ResolveUDPAddr("udp", c.serverAddr)
	if err != nil {
		return nil, xerrors.Errorf("resolve error: %w", err)
	}
	conn, err := net.DialUDP("udp", localAddr, serverAddr)
	if err != nil {
		return nil, xerrors.Errorf("listen error: %w", err)
	}
	return conn, nil
}

func (c *Client) createBindingRequest() []byte {
	data := make([]byte, MessageHeaderLength)
	nbo.PutUint16(data[0:2], BindingRequest)
	nbo.PutUint16(data[2:4], 0)
	nbo.PutUint32(data[4:8], MagicCookie)
	_, _ = rand.Read(data[8:MessageHeaderLength])
	return data
}

func (c *Client) error(err error) Result {
	return Result{err, net.UDPAddr{}}
}

func (c *Client) parseV4(address []byte) net.IP {
	resultIP := make([]byte, 4)
	ip := nbo.Uint32(address) ^ MagicCookie
	nbo.PutUint32(resultIP[:], ip)
	return resultIP
}

func (c *Client) parseV6(address []byte, txID [TransactionIDLength]byte) net.IP {
	resultIP := make([]byte, 16)
	xorValue := make([]byte, net.IPv6len)
	nbo.PutUint32(xorValue[0:4], MagicCookie)
	copy(xorValue[4:], txID[:])
	safeXORBytes(resultIP[:], address, xorValue)
	return resultIP
}
