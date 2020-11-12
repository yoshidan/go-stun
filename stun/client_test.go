package stun

import (
	"context"
	"log"
	"net"
	"testing"
	"time"
)

func TestClientGoogleServer(t *testing.T) {

	laddr, _ := net.ResolveUDPAddr("udp", ":50000")
	client := NewClient(context.Background(), "stun.l.google.com:19302", laddr)
	result := client.Discover()
	if result.Error != nil {
		t.Fatalf("%+v", result.Error)
		return
	}
	if result.Addr.Port != 50000 {
		t.Fatal("invalid port ")
		return
	}
	if result.Addr.IP.To4() == nil {
		t.Fatalf("invalid address")
		return
	}
	log.Print(result.Addr.IP.String())
}

func TestClientV4(t *testing.T) {

	server := NewServer(context.Background(), ":3478")
	errs := make(chan error, 1)

	go func() {
		err := server.ListenAndServe()
		errs <- err
	}()

	time.Sleep(time.Second)

	laddr, _ := net.ResolveUDPAddr("udp", "127.0.0.1:50000")
	client := NewClient(context.Background(), ":3478", laddr)
	result := client.Discover()
	if result.Error != nil {
		t.Fatalf("%+v", result.Error)
		return
	}
	if result.Addr.Port != 50000 {
		t.Fatal("invalid port ")
		return
	}
	if result.Addr.IP.String() != "127.0.0.1" {
		t.Fatalf("invalid address: actual=%s", result.Addr.IP.String())
		return
	}

}

func TestClientV6(t *testing.T) {

	server := NewServer(context.Background(), ":3478")
	errs := make(chan error, 1)

	go func() {
		err := server.ListenAndServe()
		errs <- err
	}()

	time.Sleep(time.Second)

	laddr, _ := net.ResolveUDPAddr("udp", "[::1]:50000")
	client := NewClient(context.Background(), ":3478", laddr)
	result := client.Discover()
	if result.Error != nil {
		t.Fatalf("%+v", result.Error)
		return
	}
	if result.Addr.Port != 50000 {
		t.Fatal("invalid port ")
		return
	}
	if result.Addr.IP.String() != "::1" {
		t.Fatalf("invalid address: actual=%s", result.Addr.IP.String())
		return
	}
}

func TestClientKeepalive(t *testing.T) {

	server := NewServer(context.Background(), ":3478")
	errs := make(chan error, 1)

	go func() {
		err := server.ListenAndServe()
		errs <- err
	}()

	time.Sleep(time.Second)

	laddr, _ := net.ResolveUDPAddr("udp", "127.0.0.1:50000")
	client := NewClient(context.Background(), "127.0.0.1:3478", laddr)
	err := client.Keepalive()
	defer client.Close()
	if err != nil {
		t.Fatalf("%+v", err)
		return
	}

	discover := func() {
		result := client.Discover()
		if result.Error != nil {
			t.Fatalf("%+v", result.Error)
			return
		}
		if result.Addr.Port != 50000 {
			t.Fatal("invalid port ")
			return
		}
		if result.Addr.IP.String() != "127.0.0.1" {
			t.Fatalf("invalid address: actual=%s", result.Addr.IP.String())
			return
		}
	}
	discover()
	discover()
	discover()
	discover()
}
