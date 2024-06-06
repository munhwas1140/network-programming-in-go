// tftp 의 일부를 제한적인 기능으로 구현함
// https://datatracker.ietf.org/doc/html/rfc783
package tftp

import (
	"bytes"
	"encoding/binary"
	"errors"
	"io"
	"strings"
)

const (
	DatagramSize = 516              // 최대 지원하는 데이터그램 크기
	BlockSize    = DatagramSize - 4 // DatagramSize - 4바이트 헤더
)

type OpCode int16

const (
	OpRRQ OpCode = iota + 1 // Read Request
	_                       // 쓰기요청인 OpWRQ 미지원
	OpData
	OpAck
	OpErr
)

type ErrCode uint16

const (
	ErrUnknown ErrCode = iota + 1
	ErrNotFound
	ErrAccessViolation
	ErrDiskFull
	ErrIllegalOp
	ErrUnknownID
	ErrFileExists
	ErrNoUser
)

type ReadReq struct {
	Filename string
	Mode     string // netascii, octet
}

// 서버에서 사용되지는 않지만 클라이언트가 이 메서드를 사용한다.
func (q ReadReq) MarshalBinary() ([]byte, error) {
	mode := "octet"
	if q.Mode != "" {
		mode = q.Mode
	}

	// OP코드 + 파일명 + 0바이트 + 모드 정보 + 0 바이트

	cap := 2 + 2 + len(q.Filename) + 1 + len(q.Mode) + 1

	b := new(bytes.Buffer)
	b.Grow(cap)

	err := binary.Write(b, binary.BigEndian, OpRRQ) // OP코드 쓰기
	if err != nil {
		return nil, err
	}

	_, err = b.WriteString(q.Filename)
	if err != nil {
		return nil, err
	}

	err = b.WriteByte(0)
	if err != nil {
		return nil, err
	}

	_, err = b.WriteString(mode)
	if err != nil {
		return nil, err
	}

	err = b.WriteByte(0)
	if err != nil {
		return nil, err
	}

	return b.Bytes(), nil
}

func (q *ReadReq) UnmarshalBinary(p []byte) error {
	r := bytes.NewBuffer(p)

	var code OpCode

	err := binary.Read(r, binary.BigEndian, &code)
	if err != nil {
		return err
	}

	if code != OpRRQ {
		return errors.New("invalid RRQ")
	}

	q.Filename, err = r.ReadString(0)
	if err != nil {
		return errors.New("invalid RRQ")
	}

	q.Filename = strings.TrimRight(q.Filename, "\x00")
	if len(q.Filename) == 0 {
		return errors.New("invalid RRQ")
	}

	q.Mode, err = r.ReadString(0)
	if err != nil {
		return errors.New("invalid RRQ")
	}

	q.Mode = strings.TrimRight(q.Mode, "\x00")
	if len(q.Mode) == 0 {
		return errors.New("invalid RRQ")
	}

	actual := strings.ToLower(q.Mode) // 강제로 octet 모드 설정
	if actual != "octet" {
		return errors.New("only binary transfers supported")
	}

	return nil
}

type Data struct {
	Block   uint16
	Payload io.Reader
}

func (d *Data) MarshalBinary() ([]byte, error) {
	b := new(bytes.Buffer)
	b.Grow(DatagramSize)

	d.Block++

	err := binary.Write(b, binary.BigEndian, OpData)
	if err != nil {
		return nil, err
	}

	err = binary.Write(b, binary.BigEndian, d.Block)
	if err != nil {
		return nil, err
	}

	// BlockSize만큼 쓰기
	_, err = io.CopyN(b, d.Payload, BlockSize)
	if err != nil && err != io.EOF {
		return nil, err
	}

	return b.Bytes(), nil
}

func (d *Data) UnmarshalBinary(p []byte) error {
	if l := len(p); l < 4 || l > DatagramSize {
		return errors.New("invalid DATA")
	}

	var opcode OpCode

	err := binary.Read(bytes.NewReader(p[:2]), binary.BigEndian, &opcode)
	if err != nil || opcode != OpData {
		return errors.New("invalid DATA")
	}

	err = binary.Read(bytes.NewReader(p[2:4]), binary.BigEndian, &d.Block)
	if err != nil {
		return errors.New("invalid DATA")
	}
	d.Payload = bytes.NewBuffer(p[4:])

	return nil
}

type Ack uint16

func (a Ack) MarshalBinary() ([]byte, error) {
	cap := 2 + 2 // OP code, Block number

	b := new(bytes.Buffer)
	b.Grow(cap)

	err := binary.Write(b, binary.BigEndian, OpAck) // Op 코드 쓰기
	if err != nil {
		return nil, err
	}

	err = binary.Write(b, binary.BigEndian, a) // 블록 번호 쓰기
	if err != nil {
		return nil, err
	}

	return b.Bytes(), nil
}

func (a *Ack) UnmarshalBinary(p []byte) error {
	var opcode OpCode

	r := bytes.NewReader(p)

	err := binary.Read(r, binary.BigEndian, &opcode)
	if err != nil {
		return err
	}

	if opcode != OpAck {
		return errors.New("invalid ACK")
	}

	return binary.Read(r, binary.BigEndian, a)
}

type Err struct {
	Error   ErrCode
	Message string
}

func (e Err) MarshalBinary() ([]byte, error) {
	// OP코드 + 에러 코드 + 메시지 + 0바이트

	cap := 2 + 2 + len(e.Message) + 1

	b := new(bytes.Buffer)
	b.Grow(cap)

	err := binary.Write(b, binary.BigEndian, OpErr) // OP 코드 쓰기
	if err != nil {
		return nil, err
	}

	err = binary.Write(b, binary.BigEndian, e.Error) // 에러코드 쓰기
	if err != nil {
		return nil, err
	}

	_, err = b.WriteString(e.Message) // 메시지 쓰기
	if err != nil {
		return nil, err
	}

	err = b.WriteByte(0) // 0바이트 쓰기
	if err != nil {
		return nil, err
	}

	return b.Bytes(), nil
}

func (e *Err) UnmarshalBinary(p []byte) error {
	r := bytes.NewBuffer(p)

	var code OpCode

	err := binary.Read(r, binary.BigEndian, &code)
	if err != nil {
		return err
	}

	if code != OpErr {
		return errors.New("invalid ERROR")
	}

	err = binary.Read(r, binary.BigEndian, &e.Error)
	if err != nil {
		return err
	}

	e.Message, err = r.ReadString(0)
	e.Message = strings.TrimRight(e.Message, "\x00")

	return err
}
