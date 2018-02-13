package protocol

import (
	"fmt"
)

type ProtocolError struct {
	msg string
	code int
}

func (err *ProtocolError) Error() string {
	return fmt.Sprintf("Error [%03d] : %s", err.code, err.msg)
}

func NewProtocolError(code int, msg string) error {
	err := ProtocolError {
		msg: msg,
		code: code,
	}
	return &err
}

const (
	HashLength	= 32
	UUIDLength	= 16
	EnodeLength	= 64
)

type ByteArray interface {
	Bytes() []byte
}

type Byte32 	[HashLength]byte
type Byte16 	[UUIDLength]byte
type Byte64	[EnodeLength]byte

func (bytes *Byte32) Bytes() []byte {
	return bytes[:]
}

func (bytes *Byte16) Bytes() []byte {
	return bytes[:]
}

func (bytes *Byte64) Bytes() []byte {
	return bytes[:]
}

func copyBytes(source []byte, dest []byte, size int) {
	i := len(source)
	if i > size {
		i = size
	}
	for ;i > 0; i-- {
		dest[i-1] = source[i-1]
	}
}

func BytesToByte16(source []byte) *Byte16 {
	var byte16 Byte16
	copyBytes(source, byte16.Bytes(), UUIDLength)
	return &byte16
}

func BytesToByte32(source []byte) *Byte32 {
	var byte32 Byte32
	copyBytes(source, byte32.Bytes(), HashLength)
	return &byte32
}

func BytesToByte64(source []byte) *Byte64 {
	var byte64 Byte64
	copyBytes(source, byte64.Bytes(), EnodeLength)
	return &byte64
}
