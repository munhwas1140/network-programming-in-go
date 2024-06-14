package logmetrics

import (
	"io"

	"go.uber.org/multierr"
)

// log생성 시 io.Writer를 받는데, io.MultiWriter를 활용하여
// 로그를 동시에 여러 wirter로 쓸 수 있다.
// io.MultiWriter는 여러 writer에 순차적으로 쓰기를 시도하는데,
// 한 번이라도 Wirte호출이 실패하면 그대로 종료된다.
//
// 에러가 발생하더라도 writer가 실패하지 않고 동작하도록 구현
type sustainedMultiWriter struct {
	writers []io.Writer
}

func (s *sustainedMultiWriter) Write(p []byte) (n int, err error) {
	for _, w := range s.writers {
		i, wErr := w.Write(p)
		n += i
		err = multierr.Append(err, wErr)
	}
	return n, err
}

func SustainedMultiWriter(writers ...io.Writer) io.Writer {
	mw := &sustainedMultiWriter{writers: make([]io.Writer, 0, len(writers))}
	for _, w := range writers {
		if m, ok := w.(*sustainedMultiWriter); ok {
			mw.writers = append(mw.writers, m.writers...)
			continue
		}

		mw.writers = append(mw.writers, w)
	}

	return mw
}
