package stun

import (
	"context"
	"log"
	"testing"
	"time"
)

func TestClientGoogleServer(t *testing.T) {

	laddr := ":50000"
	client := NewClient(context.Background(), "stun.l.google.com:19302", &laddr)
	addr, err := client.Discover()
	if err != nil {
		t.Fatalf("%+v", err)
		return
	}
	if addr.Port != 50000 {
		t.Fatal("invalid port ")
		return
	}
	if addr.IP.To4() == nil {
		t.Fatalf("invalid address")
		return
	}
	log.Print(addr.IP.String())
}

func TestClientAutoLocalAddr(t *testing.T) {

	client := NewClient(context.Background(), "stun.l.google.com:19302", nil)
	addr, err := client.Discover()
	if err != nil {
		t.Fatalf("%+v", err)
		return
	}
	if addr.IP.To4() == nil {
		t.Fatalf("invalid address")
		return
	}
	log.Print(addr.String())
}

func TestClientV4(t *testing.T) {

	server := NewServer(context.Background(), ":3478")
	errs := make(chan error, 1)

	go func() {
		err := server.ListenAndServe()
		errs <- err
	}()

	time.Sleep(time.Second)

	laddr := "127.0.0.1:50000"
	client := NewClient(context.Background(), ":3478", &laddr)
	addr, err := client.Discover()
	if err != nil {
		t.Fatalf("%+v", err)
		return
	}
	if addr.Port != 50000 {
		t.Fatal("invalid port ")
		return
	}
	if addr.IP.String() != "127.0.0.1" {
		t.Fatalf("invalid address: actual=%s", addr.IP.String())
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

	laddr := "[::1]:50000"
	client := NewClient(context.Background(), ":3478", &laddr)
	addr, err := client.Discover()
	if err != nil {
		t.Fatalf("%+v", err)
		return
	}
	if addr.Port != 50000 {
		t.Fatal("invalid port ")
		return
	}
	if addr.IP.String() != "::1" {
		t.Fatalf("invalid address: actual=%s", addr.IP.String())
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

	laddr := "127.0.0.1:50000"
	client := NewClient(context.Background(), "127.0.0.1:3478", &laddr)
	err := client.Keepalive()
	defer client.Close()
	if err != nil {
		t.Fatalf("%+v", err)
		return
	}

	discover := func() {
		addr, err := client.Discover()
		if err != nil {
			t.Fatalf("%+v", err)
			return
		}
		if addr.Port != 50000 {
			t.Fatal("invalid port ")
			return
		}
		if addr.IP.String() != "127.0.0.1" {
			t.Fatalf("invalid address: actual=%s", addr.IP.String())
			return
		}
	}
	discover()
	discover()
	discover()
	discover()
}
