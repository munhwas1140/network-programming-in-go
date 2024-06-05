package main

import (
	"io"
	"net"
	"sync"
	"testing"
)

func proxy(from io.Reader, to io.Writer) error {
	fromWriter, fromIsWriter := from.(io.Writer)
	toReader, toIsReader := to.(io.Reader)

	if toIsReader && fromIsWriter {
		// 필요한 인터페이스르 ㄹ모두 구현하였으니
		// from 과 to에 대응하는 프락시 생성
		go func() { _, _ = io.Copy(fromWriter, toReader) }()
	}

	_, err := io.Copy(to, from)
	return err
}

func TestProxy(t *testing.T) {
	var wg sync.WaitGroup

	// 서버는 "ping" 메시지를 대기하고 "pong" 메시지로 응답한다.
	// 그 외의 메시지는 동일하게 클라이언트로 에코잉된다.

	server, err := net.Listen("tcp", "127.0.0.1:")
	if err != nil {
		t.Fatal(err)
	}

	wg.Add(1)

	go func() {
		defer wg.Done()

		for {
			conn, err := server.Accept()
			if err != nil {
				return
			}

			go func(c net.Conn) {
				defer c.Close()

				for {
					buf := make([]byte, 1024)
					n, err := c.Read(buf)
					if err != nil {
						if err != io.EOF {
							t.Error(err)
						}
						return
					}

					switch msg := string(buf[:n]); msg {
					case "ping":
						_, err = c.Write([]byte("pong"))
					default:
						_, err = c.Write(buf[:n])
					}

					if err != nil {
						if err != io.EOF {
							t.Error(err)
						}
						return
					}
				}
			}(conn)
		}
	}()

	// proxyServer는 메시지를 클라이언트 연결로부터 destinationServer로 프락시한다.
	// destinationServer 서버에서 온 응답 메시지는 역으로 클라이언트에게 프락시된다.
	proxyServer, err := net.Listen("tcp", "127.0.0.1:")
	if err != nil {
		t.Fatal(err)
	}

	wg.Add(1)

	go func() {
		defer wg.Done()

		for {
			conn, err := proxyServer.Accept()
			if err != nil {
				return
			}

			go func(from net.Conn) {
				defer from.Close()

				to, err := net.Dial("tcp", server.Addr().String())
				if err != nil {
					t.Error(err)
					return
				}
				defer to.Close()

				err = proxy(from, to)
				if err != nil && err != io.EOF {
					t.Error(err)
				}
			}(conn)
		}
	}()
}
