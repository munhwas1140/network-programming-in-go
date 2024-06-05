package tcpip

import (
	"context"
	"net"
	"syscall"
	"testing"
	"time"
)

func TestDialContext(t *testing.T) {
	dl := time.Now().Add(5 * time.Second)
	ctx, cancel := context.WithDeadline(context.Background(), dl)
	defer cancel()

	var d net.Dialer // DialContext는 Dialer의 메서드이다.

	// Control은 Control이 not nil 이면 실제 dialing 하기 전에 먼저 실행한다.
	d.Control = func(_, _ string, _ syscall.RawConn) error {
		// 콘텍스트의 데드라인이 지나기 위해 충분히 긴 시간 동안 대기한다.
		time.Sleep(5*time.Second + time.Millisecond)
		return nil
	}

	// DialContext로 Context를 사용하여 "10.0.0.0:80" 에
	// tcp 요청하려 하지만
	// Control를 수정하여 context의 DeadLine보다 더 긴 시간 대기하도록 한다.
	// 그 후 연결하려고 하지만 context가 deadline으로 끝나버림
	conn, err := d.DialContext(ctx, "tcp", "10.0.0.0:80")

	if err == nil {
		conn.Close()
		t.Fatal("connection did not time out")
	}
	nErr, ok := err.(net.Error)
	if !ok {
		t.Error(err)
	} else {
		if !nErr.Timeout() {
			t.Errorf("error is not a timeout: %v", err)
		}
	}

	if ctx.Err() != context.DeadlineExceeded {
		t.Errorf("expected deadline exceeded; actual: %v", ctx.Err())
	}

	// 테스트코드가 익숙지 않아 읽기 힘들지만
	// DialContext 의 결과로 timeout 에러가 나고
	// ctx.Err() 는 context.DeadlineExceeded 가 된다.
}
