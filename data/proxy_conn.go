package main

import (
	"io"
	"net"
)

func proxyConn(source, destination string) error {
	connSource, err := net.Dial("tcp", source)
	if err != nil {
		return err
	}
	defer connSource.Close()

	connDestination, err := net.Dial("tcp", destination)
	if err != nil {
		return err
	}
	defer connDestination.Close()

	// connSource에 대응하는 connDestination
	go func() { _, _ = io.Copy(connSource, connDestination) }()

	// connDestination으로 메시지를 보내는 connSource
	_, err = io.Copy(connDestination, connSource)
	return err
}
