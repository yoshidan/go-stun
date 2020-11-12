package stun

import (
	"context"
	"encoding/binary"
	"net"
	"sync"
	"testing"
	"time"

	"golang.org/x/xerrors"
)

func TestBindingRequestV4(t *testing.T) {

	server := NewServer(context.Background(), ":3478")
	errs := make(chan error, 1)

	go func() {
		err := server.ListenAndServe()
		errs <- err
	}()

	time.Sleep(time.Second)

	laddr, err := net.ResolveUDPAddr("udp", "127.0.0.1:50000")
	raddr, err := net.ResolveUDPAddr("udp", ":3478")
	conn, err := net.DialUDP("udp", laddr, raddr)
	if err != nil {
		t.Fatal(err)
		return
	}

	txID := [TransactionIDLength]byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x10, 0x11, 0x12}
	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		buf := make([]byte, MessageHeaderLength+12)
		_, err = conn.Read(buf)
		if err != nil {
			errs <- err
			return
		}
		result := MessageHeader{
			Type:        nbo.Uint16(buf[0:2]),
			Length:      nbo.Uint16(buf[2:4]),
			MagicCookie: nbo.Uint32(buf[4:8]),
		}
		copy(result.TransactionID[:], buf[8:20])
		if result.Type != BindingSuccessResponse {
			errs <- xerrors.New("type must be binding success response")
		}
		if result.Length != 12 {
			errs <- xerrors.New("body length must be 12")
		}
		if result.MagicCookie != MagicCookie {
			errs <- xerrors.New("invalid magic cookie")
		}
		if result.TransactionID != txID {
			errs <- xerrors.New("invalid transaction id")
		}
		if binary.BigEndian.Uint16(buf[20:22]) != AttributeXorMappedAddress {
			errs <- xerrors.New("first attribute type must be xor mapped address")
		}
		if binary.BigEndian.Uint16(buf[22:24]) != 8 {
			errs <- xerrors.New("attribute value length must be 8 ")
		}
		if buf[24] != Offset {
			errs <- xerrors.New("offset must be zero")
		}
		if buf[25] != FamilyV4 {
			errs <- xerrors.New("family must be ipv4")
		}
		if binary.BigEndian.Uint16(buf[26:28])^uint16(MagicCookie>>16) != 50000 {
			errs <- xerrors.New("port must be laddr port")
		}
		if binary.BigEndian.Uint32(buf[28:32])^MagicCookie != binary.BigEndian.Uint32(laddr.IP.To4()) {
			errs <- xerrors.New("ip must be laddr ip")
		}
	}()

	data := make([]byte, MessageHeaderLength)
	binary.BigEndian.PutUint16(data[0:2], BindingRequest)
	binary.BigEndian.PutUint16(data[2:4], 0)
	binary.BigEndian.PutUint32(data[4:8], MagicCookie)
	copy(data[8:20], txID[:])

	_, err = conn.Write(data)
	if err != nil {
		t.Fatal(err)
	}

	wg.Wait()
	server.Shutdown()

}

func TestBindingRequestV6(t *testing.T) {

	server := NewServer(context.Background(), ":3478")
	errs := make(chan error, 1)

	go func() {
		err := server.ListenAndServe()
		errs <- err
	}()

	time.Sleep(time.Second)

	laddr, err := net.ResolveUDPAddr("udp", "[::1]:50000")
	raddr, err := net.ResolveUDPAddr("udp", ":3478")
	conn, err := net.DialUDP("udp", laddr, raddr)
	if err != nil {
		t.Fatal(err)
	}

	txID := [TransactionIDLength]byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x10, 0x11, 0x12}
	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		buf := make([]byte, MessageHeaderLength+24)
		_, err = conn.Read(buf)
		if err != nil {
			errs <- err
			return
		}
		result := MessageHeader{
			Type:        nbo.Uint16(buf[0:2]),
			Length:      nbo.Uint16(buf[2:4]),
			MagicCookie: nbo.Uint32(buf[4:8]),
		}
		copy(result.TransactionID[:], buf[8:20])
		if result.Type != BindingSuccessResponse {
			errs <- xerrors.New("type must be binding success response")
		}
		if result.Length != 24 {
			errs <- xerrors.New("body length must be 24")
		}
		if result.MagicCookie != MagicCookie {
			errs <- xerrors.New("invalid magic cookie")
		}
		if result.TransactionID != txID {
			errs <- xerrors.New("invalid transaction id")
		}
		if binary.BigEndian.Uint16(buf[20:22]) != AttributeXorMappedAddress {
			errs <- xerrors.New("first attribute type must be xor mapped address")
		}
		if binary.BigEndian.Uint16(buf[22:24]) != 20 {
			errs <- xerrors.New("attribute value length must be 20")
		}
		if buf[24] != Offset {
			errs <- xerrors.New("offset must be zero")
		}
		if buf[25] != FamilyV6 {
			errs <- xerrors.New("family must be ipv6")
		}
		if binary.BigEndian.Uint16(buf[26:28])^uint16(MagicCookie>>16) != 50000 {
			errs <- xerrors.New("port must be laddr port")
		}

		resultIP := make([]byte, 16)
		xorValue := make([]byte, net.IPv6len)
		nbo.PutUint32(xorValue[0:4], MagicCookie)
		copy(xorValue[4:], txID[:])
		safeXORBytes(resultIP[:], buf[28:44], xorValue)
		if laddr.IP.String() != net.IP(resultIP[:]).String() {
			errs <- xerrors.New("ip must be laddr ip")
		}
	}()

	data := make([]byte, MessageHeaderLength)
	binary.BigEndian.PutUint16(data[0:2], BindingRequest)
	binary.BigEndian.PutUint16(data[2:4], 0)
	binary.BigEndian.PutUint32(data[4:8], MagicCookie)
	copy(data[8:20], txID[:])

	_, err = conn.Write(data)
	if err != nil {
		t.Fatal(err)
	}

	wg.Wait()
	server.Shutdown()

}
