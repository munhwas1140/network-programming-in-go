package echo

import (
	"context"
	"fmt"
	"net"
)

func echoServerUDP(ctx context.Context, addr string) (net.Addr, error) {
	s, err := net.ListenPacket("udp", addr)
	if err != nil {
		return nil, fmt.Errorf("binding to upd %s: %w", addr, err)
	}

	go func() {
		go func() {
			<-ctx.Done()
			_ = s.Close()
		}()

		buf := make([]byte, 1024)

		for {
			n, clientAddr, err := s.ReadFrom(buf) // 클라이언트에서 서버로
			if err != nil {
				return
			}

			_, err = s.WriteTo(buf[:n], clientAddr) // 서버에서 클라이언트로
			if err != nil {
				return
			}

		}
	}()
	return s.LocalAddr(), nil
}
