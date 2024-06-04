package tcpip

import (
	"io"
	"net"
	"testing"
)

func TestDial(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:")
	if err != nil {
		t.Fatal(err)
	}

	done := make(chan struct{})
	go func() {
		defer func() { done <- struct{}{} }()

		for {
			// Accept() 메서드는 리스너가 수신 연결을 감지하고
			// 클라이언트와 서버 간의 TCP 핸드셰이크 절차가 완료될 떄까지 블로킹된다.
			conn, err := listener.Accept()
			if err != nil {
				// 57번 줄에서 listener가 Close 되면
				// 블로킹이 해제되면서 에러를 반환한다.
				// ex) accept tcp 127.0.0.1:50388: use of closed network connection
				t.Log(err)
				return
			}

			go func(c net.Conn) {
				defer func() {
					c.Close()
					done <- struct{}{}
				}()

				buf := make([]byte, 1024)
				for {
					n, err := c.Read(buf)
					if err != nil {
						if err != io.EOF {
							t.Error(err)
						}
						return
					}
					t.Logf("received: %q", buf[:n])
				}
			}(conn)
		}
	}()

	conn, err := net.Dial("tcp", listener.Addr().String())
	if err != nil {
		t.Fatal(err)
	}

	conn.Close()
	<-done
	listener.Close()
	<-done
}
