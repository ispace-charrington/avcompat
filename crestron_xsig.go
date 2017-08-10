package avcompat

import "errors"

// ISC Errors
var (
	ErrIndexRange        = errors.New("Transition index exceeds encoding range")
	ErrSerialLength      = errors.New("Serial transition length cannot exceed 252 bytes")
	ErrSerialInvalidByte = errors.New("Serial transition cannot contain \\xFF byte")
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

func (o *ISCClearOperation) MarshalBinary() ([]byte, error) {
	return []byte{0xFC}, nil
}

func (o *ISCRefreshOperation) MarshalBinary() ([]byte, error) {
	return []byte{0xFD}, nil
}
