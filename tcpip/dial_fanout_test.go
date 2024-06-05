package tcpip

import (
	"context"
	"net"
	"sync"
	"testing"
	"time"
)

func TestDialContextCancelFanout(t *testing.T) {
	ctx, cancel := context.WithDeadline(
		context.Background(),
		time.Now().Add(10*time.Second),
	)

	listener, err := net.Listen("tcp", "127.0.0.1:")
	if err != nil {
		t.Fatal(err)
	}
	defer listener.Close()

	go func() {
		// 하나의 연결만을 수락한다.
		// 이 고루틴을 연결 요청이 올때까지 블로킹된다.
		// 따라서 아래의 dial함수는 한 번만 동작하는 것임
		conn, err := listener.Accept()
		if err == nil {
			conn.Close()
		}
	}()

	dial := func(ctx context.Context, address string, response chan int, id int, wg *sync.WaitGroup) {
		defer wg.Done()

		var d net.Dialer
		c, err := d.DialContext(ctx, "tcp", address)
		if err != nil {
			return
		}
		c.Close()

		select {
		case <-ctx.Done():
		case response <- id:
		}
	}

	res := make(chan int)
	var wg sync.WaitGroup

	for i := 0; i < 10; i++ {
		wg.Add(1)
		go dial(ctx, listener.Addr().String(), res, i+1, &wg)
	}

	// 10개의 고루틴 중 접속에 성공한 첫 번째 고루틴이의 번호(i + 1) 이 res 채널에들어온다.
	// 나머지는 select 에서 대기하다 cancel()이 호출되면 <-ctx.Done()에 의해 종료된다.
	response := <-res
	cancel()
	wg.Wait()
	close(res)

	if ctx.Err() != context.Canceled {
		t.Errorf("expected canceled context; actual %s", ctx.Err())
	}

	t.Logf("diaer %d retrieved the resource", response)
}
