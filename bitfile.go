package ucfont

import (
	"fmt"
	"io"
)

type BitFileError struct {
	error
}

type BitFile struct {
	io.ReadSeeker
	bitCache       uint64
	bitCacheLength uint
}

func NewBitFile(f io.ReadSeeker) *BitFile {
	return &BitFile{f, 0, 0}
}

func (f *BitFile) ReadOctet() (byte, error) {
	b := make([]byte, 1)
	n, err := f.Read(b)
	switch {
	case err != nil:
		return 0, err
	case n < 1:
		return 0, io.ErrUnexpectedEOF
	}
	return (b[0]&0xf)<<4 | b[0]>>4, nil
}

func (f *BitFile) ReadBitWithSig(bits uint) (int32, error) {
	result, err := f.ReadBit(bits)
	if err != nil {
		return 0, err
	}
	mask := int32(1<<(bits-1) - 1)
	absv := result & mask
	if result != absv {
		return -int32(absv), nil
	}
	return int32(absv), nil
}

func (f *BitFile) ReadBit(bits uint) (int32, error) {
	if bits > 32 {
		return 0, BitFileError{fmt.Errorf("bits should <= 32")}
	}
	for f.bitCacheLength < bits {
		b, err := f.ReadOctet()
		if err != nil {
			return 0, err
		}
		f.bitCacheLength += 8
		f.bitCache = f.bitCache<<8 | uint64(b)
	}
	f.bitCacheLength -= bits
	mask := uint64((1<<bits - 1) << f.bitCacheLength)
	result := (f.bitCache & mask) >> f.bitCacheLength
	f.bitCache = f.bitCache &^ mask
	return int32(result), nil
}

func (f *BitFile) Seek(offset int64, whence int) (int64, error) {
	v, err := f.ReadSeeker.Seek(offset, whence)
	f.bitCache = 0
	f.bitCacheLength = 0
	return v, err
}
