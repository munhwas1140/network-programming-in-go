package main

import (
	"io"
	"log"
	"net"
	"os"
)

// 모니터 구조체는 네트워크 트래픽을 로깅하기 위한 log.Logger를 임베딩한다.
type Monitor struct {
	*log.Logger
}

// Write메서드는 io.Wirter 인터페이스를 구현한다.
func (m *Monitor) Write(p []byte) (int, error) {
	return len(p), m.Output(2, string(p))
}

func ExampleMonitor() {
	monitor := &Monitor{Logger: log.New(os.Stdout, "monitor: ", 0)}

	listener, err := net.Listen("tcp", "127.0.0.1:")
	if err != nil {
		monitor.Fatal(err)
	}
	done := make(chan struct{})

	go func() {
		defer close(done)

		conn, err := listener.Accept()
		if err != nil {
			return
		}
		defer conn.Close()

		b := make([]byte, 1024)

		// TeeReader는 읽은 데이터를 지정된 Writer에게 전달한다.
		// 즉 conn에서 데이터를 읽고 monitor로 전달한다.
		r := io.TeeReader(conn, monitor)

		n, err := r.Read(b)
		if err != nil && err != io.EOF {
			monitor.Println(err)
			return
		}

		// MultiWriter 는 데이터를 여러 Writer에 동시에 전달한다.
		//
		// 주의
		// io.TeeReader와 io.MultiWriter함수 모두 writer로 데이터를 쓰는 동안 블로킹된다.
		// writer에서 발생한 에러는 io.TeeWriter, io.MultiWriter역시 에러를 발생하게 하며
		// 네트워크 데이터의 흐름을 중단시킨다.
		// 따라서 reader는 항상 nil 에러를 반환하도록 구현하고,
		// 하위 레벨에서 발생하는 에러는 처리할 수 있도록 로깅하는 식으로 구현하길 권장한다.
		w := io.MultiWriter(conn, monitor)

		_, err = w.Write(b[:n]) // 메시지를 에코잉한다.
		if err != nil && err != io.EOF {
			monitor.Println(err)
			return
		}
	}()

	conn, err := net.Dial("tcp", listener.Addr().String())
	if err != nil {
		monitor.Fatal(err)
	}

	_, err = conn.Write([]byte("Test\n"))
	if err != nil {
		monitor.Fatal(err)
	}

	_ = conn.Close()
	<-done

	// output
	// monitor: Test
	// monitor: Test
}
