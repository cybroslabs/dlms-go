package dlmsal

import (
	"bytes"
	"fmt"
	"io"
)

func codedlength(len uint) int {
	if len < 128 {
		return 1
	}
	if len < 256 {
		return 2
	}
	if len < 65536 {
		return 3
	}
	if len < 16777216 {
		return 4
	}
	return 5
}

func EncodeCaptureObject(classId uint16, obis *DlmsObis, attribute int8, version uint16) DlmsData {
	ch := make([]DlmsData, 4) // this bellow should be done using some function with type checking... poor maintainability
	ch[0] = DlmsData{Tag: TagLongUnsigned, Value: classId}
	ch[1] = DlmsData{Tag: TagOctetString, Value: obis}
	ch[2] = DlmsData{Tag: TagInteger, Value: attribute}
	ch[3] = DlmsData{Tag: TagLongUnsigned, Value: version}
	return DlmsData{Tag: TagStructure, Value: ch}
}

func EncodeSimpleRangeAccess(from *DlmsDateTime, to *DlmsDateTime) DlmsData {
	ch := make([]DlmsData, 4)
	ch[0] = EncodeCaptureObject(8, &DlmsObis{A: 0, B: 0, C: 1, D: 0, E: 0, F: 255}, 2, 0)
	ch[1] = DlmsData{Tag: TagOctetString, Value: *from}
	ch[2] = DlmsData{Tag: TagOctetString, Value: *to}
	ch[3] = DlmsData{Tag: TagArray, Value: nil}
	return DlmsData{Tag: TagStructure, Value: ch}
}

func encodelength(dst *bytes.Buffer, len uint) {
	if len < 128 {
		dst.WriteByte(byte(len))
		return
	}
	if len < 256 {
		dst.WriteByte(0x81)
		dst.WriteByte(byte(len))
		return
	}
	if len < 65536 {
		dst.WriteByte(0x82)
		dst.WriteByte(byte(len >> 8))
		dst.WriteByte(byte(len))
		return
	}
	if len < 16777216 {
		dst.WriteByte(0x83)
		dst.WriteByte(byte(len >> 16))
		dst.WriteByte(byte(len >> 8))
		dst.WriteByte(byte(len))
		return
	}
	dst.WriteByte(0x84)
	dst.WriteByte(byte(len >> 24))
	dst.WriteByte(byte(len >> 16))
	dst.WriteByte(byte(len >> 8))
	dst.WriteByte(byte(len))
}

func encodetag(dst *bytes.Buffer, tag byte, data []byte) {
	dst.WriteByte(tag)
	encodelength(dst, uint(len(data)))
	dst.Write(data)
}

func encodetag2(dst *bytes.Buffer, tag byte, innertag byte, data []byte) {
	if codedlength(uint(len(data))) > 1 { // so not exactly effective for memory copy and so on, but exceptional case
		var buf bytes.Buffer
		buf.WriteByte(innertag)
		encodelength(&buf, uint(len(data)))
		buf.Write(data)
		encodetag(dst, tag, buf.Bytes())
	} else {
		dst.WriteByte(tag)
		encodelength(dst, uint(len(data)+2))
		dst.WriteByte(innertag)
		encodelength(dst, uint(len(data)))
		dst.Write(data)
	}
}

func decodelength(src io.Reader, tmp []byte) (uint, int, error) {
	_, err := io.ReadFull(src, tmp[:1])
	if err != nil {
		return 0, 0, err
	}
	b := tmp[0]
	if b < 128 {
		return uint(b), 1, nil
	}
	if b == 128 {
		return 0, 0, fmt.Errorf("unsupported infinite length")
	}
	r := uint(0)
	c := int(b & 0x7f)
	if c > 4 {
		return 0, 0, fmt.Errorf("too much bytes for length")
	}
	_, err = io.ReadFull(src, tmp[:c])
	if err != nil {
		return 0, 0, err
	}
	for i := 0; i < c; i++ {
		r = (r << 8) | uint(tmp[i])
	}
	return r, c + 1, nil
}

func decodetag(src []byte, tmp []byte) (byte, int, []byte, error) {
	if len(src) < 2 {
		return 0, 0, nil, fmt.Errorf("no data available")
	}
	tag := src[0]
	t := bytes.NewBuffer(src[1:])
	dlen, c, err := decodelength(t, tmp)
	if err != nil {
		return 0, 0, nil, err
	}

	if len(src) < c+1+int(dlen) {
		return 0, 0, nil, fmt.Errorf("no data left in source")
	}
	return tag, c + 1 + int(dlen), src[1+c : 1+c+int(dlen)], nil
}

var _units = [...]string{"unknown",
	// 1
	"a",
	"mo",
	"wk",
	"d",
	"h",
	"min.",
	"s",
	"deg",
	"degC",
	// 10
	"currency",
	"m",
	"m/s",
	"m3",
	"m3",
	"m3/h",
	"m3/h",
	"m3/d",
	"m3/d",
	"l",
	// 20
	"kg",
	"N",
	"Nm",
	"Pa",
	"bar",
	"J",
	"J/h",
	"W",
	"VA",
	"var",
	// 30
	"Wh",
	"VAh",
	"varh",
	"A",
	"C",
	"V",
	"V/m",
	"F",
	"Ohm",
	"Ohm m2/m",
	// 40
	"Wb",
	"T",
	"A/m",
	"H",
	"Hz",
	"1/(Wh)",
	"1/(varh)",
	"1/(VAh)",
	"V2h",
	"A2h",
	// 50
	"kg/s",
	"S, mho",
	"K",
	"1/(V2h)",
	"1/(A2h)",
	"1/m3",
	"%",
	"Ah",
	"unknown",
	"unknown",
	// 60
	"Wh/m3",
	"J/m3",
	"Mol %",
	"g/m3",
	"Pa s",
	"unknown",
	"unknown",
	"unknown",
	"unknown",
	"unknown",
	// 70
	"dBm",
	"dbuV",
	"dB"}

func GetUnit(u uint8) string {
	if int(u) >= len(_units) {
		return _units[0]
	}
	return _units[u]
}
