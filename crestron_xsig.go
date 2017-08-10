package avcompat

import "errors"

// ISC Errors
var (
	ErrIndexRange        = errors.New("Transition index exceeds encoding range")
	ErrSerialLength      = errors.New("Serial transition length cannot exceed 252 bytes")
	ErrSerialInvalidByte = errors.New("Serial transition cannot contain \\xFF byte")
	ErrDecodeLength      = errors.New("Cannot decode due to short buffer")
	ErrDecodeIllegal     = errors.New("Cannot decode due to invalid bitstream")
)

type ISCDigitalTransition struct {
	index uint
	value bool
}

type ISCAnalogTransition struct {
	index uint
	value uint16
}

type ISCSerialTransition struct {
	index uint
	value []byte
}

type ISCClearOperation struct{}
type ISCRefreshOperation struct{}

func (t *ISCDigitalTransition) MarshalBinary() ([]byte, error) {
	var buf [2]byte
	if t.index > 4095 {
		return nil, ErrIndexRange
	}
	buf[0] = byte(0x80) | byte(0x1f&(t.index>>7))
	if !t.value {
		buf[0] |= 0x20 // contains the complement of the value
	}
	buf[1] = byte(0x7f & t.index)
	return buf[:], nil
}

func (t *ISCDigitalTransition) UnmarshalBinary(buf []byte) error {
	if len(buf) < 2 {
		return ErrDecodeLength
	}
	if (buf[0]&byte(0xC0) != byte(0x80)) || (buf[1]&byte(0x80) != byte(0x00)) {
		return ErrDecodeIllegal
	}

	t.index = uint(buf[1]) | uint(0x1f&(buf[0])<<7)
	t.value = (buf[0]&byte(0x20) == byte(0x00))
	return nil
}

func (t *ISCAnalogTransition) MarshalBinary() ([]byte, error) {
	var buf [4]byte
	if t.index > 1023 {
		return nil, ErrIndexRange
	}
	buf[0] = byte(0xc0) | byte((t.value>>14)<<4) | byte(t.index>>7)
	buf[1] = byte(0x7f & t.index)
	buf[2] = byte(0x7f & (t.value >> 7))
	buf[3] = byte(0x7f & t.value)
	return buf[0:4], nil
}

func (t *ISCAnalogTransition) UnmarshalBinary(buf []byte) error {
	if len(buf) < 4 {
		return ErrDecodeLength
	}
	if (buf[0]&byte(0xC8) != byte(0xC0)) ||
		(buf[1]&byte(0x80) != byte(0x00)) ||
		(buf[2]&byte(0x80) != byte(0x00)) ||
		(buf[3]&byte(0x80) != byte(0x00)) {
		return ErrDecodeIllegal
	}

	t.index = uint(buf[1]) | uint(0x07&(buf[0])<<7)
	t.value = uint16(0x30&buf[0]<<14) | uint16(buf[2]<<7) | uint16(buf[3])
	return nil
}

func (t *ISCSerialTransition) MarshalBinary() ([]byte, error) {
	if t.index > 1023 {
		return nil, ErrIndexRange
	}
	if len(t.value) > 252 {
		return nil, ErrSerialLength
	}
	for j := range t.value {
		if t.value[j] == byte(0xFF) {
			return nil, ErrSerialInvalidByte
		}
	}
	buf := make([]byte, len(t.value)+3, len(t.value)+3)
	buf[0] = byte(0xc8) | byte(t.index>>7)
	buf[1] = byte(0x7f & t.index)
	copy(buf[2:], t.value)
	buf[len(buf)-1] = 0xff
	return buf, nil
}

func (t *ISCSerialTransition) UnmarshalBinary(buf []byte) error {
	if len(buf) < 3 {
		return ErrDecodeLength
	}
	if buf[len(buf)-1] != 0xff {
		// this has three sane causes:
		// 1: the buffer we have is incomplete, and more data will come (error = ErrDecodeLength)
		// 2: the buffer we have contains more than one packet (error = nil)
		// 3: the buffer contains invalid data (error = ErrDecodeIllegal)
		//
		// we will assume that UnmarshalBinary will always be called with a perfectly
		// framed packet, which renders 1 & 2 impossible.
		return ErrDecodeIllegal
	}

	if (buf[0]&byte(0xF8) != byte(0xC8)) || (buf[1]&byte(0x80) != byte(0x00)) {
		return ErrDecodeIllegal
	}

	t.index = uint(buf[1]) | uint(0x07&(buf[0])<<7)
	t.value = make([]byte, len(buf)-3)
	copy(t.value, buf[2:])
	return nil
}

func (o *ISCClearOperation) MarshalBinary() ([]byte, error) {
	return []byte{0xFC}, nil
}

func (o *ISCClearOperation) UnmarshalBinary(buf []byte) error {
	if len(buf) < 1 {
		return ErrDecodeLength
	}
	if buf[0] != 0xFC {
		return ErrDecodeIllegal
	}
	return nil
}

func (o *ISCRefreshOperation) MarshalBinary() ([]byte, error) {
	return []byte{0xFD}, nil
}

func (o *ISCRefreshOperation) UnmarshalBinary(buf []byte) error {
	if len(buf) < 1 {
		return ErrDecodeLength
	}
	if buf[0] != 0xFD {
		return ErrDecodeIllegal
	}
	return nil
}
