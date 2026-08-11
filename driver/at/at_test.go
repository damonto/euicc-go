package at

import "testing"

func TestNewRejectsEmptyDevice(t *testing.T) {
	if _, err := New(""); err == nil {
		t.Fatal("New() error = nil, want empty device error")
	}
}

func TestNewDoesNotOpenDevice(t *testing.T) {
	channel, err := New("/does/not/exist")
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	if channel.channel != nil {
		t.Fatal("New() opened the serial port")
	}
	if _, err := channel.Transmit(nil); err == nil {
		t.Fatal("Transmit() before Connect error = nil")
	}
	if err := channel.Disconnect(); err != nil {
		t.Fatalf("Disconnect() before Connect error = %v", err)
	}
	if err := channel.Connect(); err == nil {
		t.Fatal("Connect() after Disconnect error = nil")
	}
}

func TestConnectOpensDevice(t *testing.T) {
	channel, err := New("/does/not/exist")
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	if err := channel.Connect(); err == nil {
		t.Fatal("Connect() error = nil for missing device")
	}
	if channel.channel != nil {
		t.Fatal("Connect() retained a channel after open failure")
	}
}
