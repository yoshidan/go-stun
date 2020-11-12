package stun

import (
	"encoding/binary"
	"errors"
	"net"
)

const (
	BindingRequest         uint16 = 0x0001
	BindingSuccessResponse uint16 = 0x0101
	BindingErrorResponse   uint16 = 0x0111
)

const (
	MessageHeaderLength = 20
	TransactionIDLength = 12
)

const MagicCookie uint32 = 0x2112A442
const FamilyV4 uint8 = 0x01
const FamilyV6 uint8 = 0x02
const Offset uint8 = 0x00

const ErrorCodePadding uint16 = 0x0000
const ClazzBadRequest uint8 = 0x04
const ErrNumBadRequest uint8 = 0x00

const (
	AttributeXorMappedAddress uint16 = 0x0020
	AttributeErrorCode        uint16 = 0x0009
)

type MessageHeader struct {
	Type          uint16
	Length        uint16
	MagicCookie   uint32
	TransactionID [TransactionIDLength]byte
}

type XorMappedAddressAttribute struct {
	Type   uint16
	Length uint16
	Offset uint8
	Family uint8
	XPort  uint16
}

type IPV4Attribute struct {
	XorMappedAddressAttribute
	XAddress uint32
}

type IPV6Attribute struct {
	XorMappedAddressAttribute
	XAddress [16]uint8
}

type SuccessV4Response struct {
	Header    MessageHeader
	Attribute IPV4Attribute
}

type SuccessV6Response struct {
	Header    MessageHeader
	Attribute IPV6Attribute
}

type ErrorResponse struct {
	Header MessageHeader
}

var nbo = binary.BigEndian

func NewMessage(buffer []byte) (MessageHeader, error) {
	message := MessageHeader{
		Type:        nbo.Uint16(buffer[0:2]),
		Length:      nbo.Uint16(buffer[2:4]),
		MagicCookie: nbo.Uint32(buffer[4:8]),
	}
	copy(message.TransactionID[:], buffer[8:20])
	if message.Type != BindingRequest {
		return message, errors.New("invalid message type")
	}
	if message.Length != 0 {
		return message, errors.New("body must be zero")
	}
	if message.MagicCookie != MagicCookie {
		return message, errors.New("invalid magic cookie")
	}
	return message, nil
}

func (m *MessageHeader) CreateSuccessResponse(port uint16, address net.IP) []byte {
	var attributeValueLength uint16 = 8
	var family uint8 = FamilyV4
	if address.To4() == nil {
		attributeValueLength = 20
		family = FamilyV6
	}
	bodyLength := 4 + attributeValueLength
	bufferSize := MessageHeaderLength + bodyLength
	writeBuf := make([]byte, bufferSize)

	nbo.PutUint16(writeBuf[0:2], BindingSuccessResponse)
	nbo.PutUint16(writeBuf[2:4], bodyLength)
	nbo.PutUint32(writeBuf[4:8], m.MagicCookie)
	copy(writeBuf[8:20], m.TransactionID[:])

	//RFC 5389 15.2 xor mapped address
	nbo.PutUint16(writeBuf[20:22], AttributeXorMappedAddress)
	nbo.PutUint16(writeBuf[22:24], attributeValueLength)
	writeBuf[24] = Offset
	writeBuf[25] = family
	nbo.PutUint16(writeBuf[26:28], port^uint16(m.MagicCookie>>16))

	if address.To4() == nil {
		xorValue := make([]byte, net.IPv6len)
		nbo.PutUint32(xorValue[0:4], m.MagicCookie)
		copy(xorValue[4:], m.TransactionID[:])
		safeXORBytes(writeBuf[28:44], address.To16(), xorValue)
	} else {
		nbo.PutUint32(writeBuf[28:32], nbo.Uint32(address.To4())^m.MagicCookie)
	}
	return writeBuf
}

func (m *MessageHeader) CreateErrorResponse() []byte {
	var bodyLength uint16 = 8
	writeBuf := make([]byte, MessageHeaderLength+bodyLength)
	nbo.PutUint16(writeBuf[0:2], BindingErrorResponse)
	nbo.PutUint16(writeBuf[2:4], bodyLength)
	nbo.PutUint32(writeBuf[4:8], MagicCookie)
	copy(writeBuf[8:20], m.TransactionID[:])

	//RFC 5389 15.6. ERROR-CODE
	//Bad Request Only
	nbo.PutUint16(writeBuf[20:22], AttributeErrorCode)
	nbo.PutUint16(writeBuf[22:24], 4)
	nbo.PutUint16(writeBuf[22:26], ErrorCodePadding)
	writeBuf[26] = ClazzBadRequest
	writeBuf[27] = ErrNumBadRequest
	return writeBuf
}

func safeXORBytes(dst, a, b []byte) {
	n := len(a)
	if len(b) < n {
		n = len(b)
	}
	if n == 0 {
		return
	}
	for i := 0; i < n; i++ {
		dst[i] = a[i] ^ b[i]
	}
}
