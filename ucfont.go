package ucfont

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"golang.org/x/text/encoding"
	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/transform"
	"io"
	"io/ioutil"
)

type CharacterIndex struct {
	Offset uint32
	Length uint16
}

type PSFontFile struct {
	io.ReadSeeker
	IsGBK bool
}

type FontError struct {
	error
}

func NewPSFontFile(f io.ReadSeeker, isGBK bool) *PSFontFile {
	fontFile := &PSFontFile{f, isGBK}
	return fontFile
}

func ConvertToPath(data []byte) string {
	bf := NewBitFile(bytes.NewReader(data))
	var cx, cy int32
	path := new(bytes.Buffer)
	for {
		cmd, err := bf.ReadBit(4)
		if err != nil {
			break
		}
		switch cmd {
		case 0:
			cx, _ = bf.ReadBit(8)
			cy, _ = bf.ReadBit(8)
			fmt.Fprintf(path, "M%d,%d ", cx, cy)
		case 1:
			cx, _ = bf.ReadBit(8)
			fmt.Fprintf(path, "H%d ", cx)
		case 2:
			cy, _ = bf.ReadBit(8)
			fmt.Fprintf(path, "V%d ", cy)
		case 3:
			cx, _ = bf.ReadBit(8)
			cy, _ = bf.ReadBit(8)
			fmt.Fprintf(path, "L%d,%d ", cx, cy)
		case 4:
			x1, _ := bf.ReadBit(8)
			y1, _ := bf.ReadBit(8)
			cx, _ = bf.ReadBit(8)
			cy, _ = bf.ReadBit(8)
			fmt.Fprintf(path, "Q%d,%d,%d,%d ", x1, y1, cx, cy)
		case 5:
			x1, _ := bf.ReadBit(8)
			y1, _ := bf.ReadBit(8)
			x2, _ := bf.ReadBit(8)
			y2, _ := bf.ReadBit(8)
			cx, _ = bf.ReadBit(8)
			cy, _ = bf.ReadBit(8)
			fmt.Fprintf(path, "C%d,%d,%d,%d,%d,%d ", x1, y1, x2, y2, cx, cy)
		case 6:
			x1, _ := bf.ReadBit(8)
			y1, _ := bf.ReadBit(8)
			x2, _ := bf.ReadBit(8)
			y2, _ := bf.ReadBit(8)
			fmt.Fprintf(path, "M%d,%d H%d V%d H%d V%d M%d,%d ", x1, y1, x2, y2, x1, y1, cx, cy)
		case 7:
			dx, _ := bf.ReadBitWithSig(4)
			cx += dx
			cy, _ = bf.ReadBit(8)
			fmt.Fprintf(path, "L%d,%d ", cx, cy)
		case 8:
			cx, _ = bf.ReadBit(8)
			dy, _ := bf.ReadBitWithSig(4)
			cy += dy
			fmt.Fprintf(path, "L%d,%d ", cx, cy)
		case 9:
			dx, _ := bf.ReadBitWithSig(4)
			dy, _ := bf.ReadBitWithSig(4)
			cx += dx
			cy += dy
			fmt.Fprintf(path, "L%d,%d ", cx, cy)
		case 10:
			dx, _ := bf.ReadBitWithSig(6)
			dy, _ := bf.ReadBitWithSig(6)
			cx += dx
			cy += dy
			fmt.Fprintf(path, "L%d,%d ", cx, cy)
		case 11:
			dx1, _ := bf.ReadBitWithSig(4)
			dy1, _ := bf.ReadBitWithSig(4)
			dx2, _ := bf.ReadBitWithSig(4)
			dy2, _ := bf.ReadBitWithSig(4)
			x1 := cx + dx1
			y1 := cy + dy1
			cx = x1 + dx2
			cy = y1 + dy2
			fmt.Fprintf(path, "Q%d,%d,%d,%d ", x1, y1, cx, cy)
		case 12:
			dx1, _ := bf.ReadBitWithSig(6)
			dy1, _ := bf.ReadBitWithSig(6)
			dx2, _ := bf.ReadBitWithSig(6)
			dy2, _ := bf.ReadBitWithSig(6)
			x1 := cx + dx1
			y1 := cy + dy1
			cx = x1 + dx2
			cy = y1 + dy2
			fmt.Fprintf(path, "Q%d,%d,%d,%d ", x1, y1, cx, cy)
		case 13:
			dx1, _ := bf.ReadBitWithSig(4)
			dy1, _ := bf.ReadBitWithSig(4)
			dx2, _ := bf.ReadBitWithSig(4)
			dy2, _ := bf.ReadBitWithSig(4)
			dx3, _ := bf.ReadBitWithSig(4)
			dy3, _ := bf.ReadBitWithSig(4)
			x1 := cx + dx1
			y1 := cy + dy1
			x2 := x1 + dx2
			y2 := y1 + dy2
			cx = x2 + dx3
			cy = y2 + dy3
			fmt.Fprintf(path, "C%d,%d,%d,%d,%d,%d ", x1, y1, x2, y2, cx, cy)
		case 14:
			dx1, _ := bf.ReadBitWithSig(6)
			dy1, _ := bf.ReadBitWithSig(6)
			dx2, _ := bf.ReadBitWithSig(6)
			dy2, _ := bf.ReadBitWithSig(6)
			dx3, _ := bf.ReadBitWithSig(6)
			dy3, _ := bf.ReadBitWithSig(6)
			x1 := cx + dx1
			y1 := cy + dy1
			x2 := x1 + dx2
			y2 := y1 + dy2
			cx = x2 + dx3
			cy = y2 + dy3
			fmt.Fprintf(path, "C%d,%d,%d,%d,%d,%d ", x1, y1, x2, y2, cx, cy)
		case 15:
			bf.ReadBit(12)
		}
	}
	return path.String()
}

func (f *PSFontFile) readCharData(id int) []byte {
	idx := f.readIndex(id)
	buf := make([]byte, idx.Length)
	f.Seek(int64(idx.Offset), 0)
	f.Read(buf)
	return buf
}

func (f *PSFontFile) readIndex(id int) (idx *CharacterIndex) {
	idx = &CharacterIndex{}
	f.Seek(int64(id*6), 0)
	binary.Read(f, binary.LittleEndian, idx)
	if idx.Offset > 0x10000000 {
		idx.Offset -= 0x10000000
	}
	return idx
}

func (f *PSFontFile) getCharID(c rune) (int, error) {
	encoder := encoding.ReplaceUnsupported(simplifiedchinese.GBK.NewEncoder())
	sourceReader := bytes.NewReader([]byte(string(c)))
	data, _ := ioutil.ReadAll(transform.NewReader(sourceReader, encoder))
	if len(data) != 2 {
		return 0, FontError{fmt.Errorf("Invalid GBK Character")}
	}

	c1, c2 := int(data[0]), int(data[1])
	if c1 >= 0xb0 && c1 <= 0xf7 && c2 >= 0xa1 && c2 <= 0xfe {
		if f.IsGBK {
			c1 -= 0xa1
		} else {
			c1 -= 0xb0
		}
		return c1*94 + c2 - 0xa1, nil
	}
	if !f.IsGBK {
		return 0, FontError{fmt.Errorf("Invalid GB2312 Character")}
	}
	if c2 >= 0xa1 && c2 <= 0xfe {
		if c1 >= 0xa1 {
			c1 -= 0xa1
		} else {
			c1 -= 0x23
		}
		return c1*94 + c2 - 0xa1, nil
	} else {
		if c2 >= 0x80 {
			c2 -= 1
		}
		return 11844 + (c1-0x81)*96 + c2 - 0x40, nil
	}
}

func (f *PSFontFile) GetCharPath(c rune) (string, error) {
	id, err := f.getCharID(c)
	if err != nil {
		return "", err
	}
	data := f.readCharData(id)
	return ConvertToPath(data), nil
}
