package dlmsal

import (
	"fmt"
	"io"
)

// SN func read, for now it should be enough
func (d *dlmsal) Read(items []DlmsSNRequestItem) ([]DlmsData, error) {
	if !d.isopen {
		return nil, fmt.Errorf("connection is not open")
	}

	if len(items) == 0 {
		return nil, fmt.Errorf("no items to read")
	}

	// format request into byte slice and send that to unit
	d.pdu.Reset()
	d.pdu.WriteByte(byte(TagReadRequest))
	encodelength(&d.pdu, uint(len(items)))
	for _, item := range items {
		if item.HasAccess {
			d.pdu.WriteByte(4)
			d.pdu.WriteByte(byte(item.Address >> 8))
			d.pdu.WriteByte(byte(item.Address))
			d.pdu.WriteByte(item.AccessDescriptor)
			err := encodeData(&d.pdu, item.AccessData)
			if err != nil {
				return nil, err
			}
		} else {
			d.pdu.WriteByte(2)
			d.pdu.WriteByte(byte(item.Address >> 8))
			d.pdu.WriteByte(byte(item.Address))
		}
	}

	err := d.transport.Write(d.pdu.Bytes())
	if err != nil {
		return nil, err
	}

	_, err = io.ReadFull(d.transport, d.tmpbuffer[:1])
	if err != nil {
		return nil, err
	}
	if d.tmpbuffer[0] != byte(TagReadResponse) {
		return nil, fmt.Errorf("unexpected tag: %x", d.tmpbuffer[0])
	}
	l, _, err := decodelength(d.transport, d.tmpbuffer)
	if err != nil {
		return nil, err
	}
	if int(l) != len(items) {
		return nil, fmt.Errorf("different amount of data received")
	}
	ret := make([]DlmsData, len(items))
	for i := 0; i < len(ret); i++ {
		_, err = io.ReadFull(d.transport, d.tmpbuffer[:1])
		if err != nil {
			return nil, err
		}
		switch d.tmpbuffer[0] {
		case 0:
			ret[i], _, err = d.decodeDataTag(d.transport)
			if err != nil {
				return nil, err
			}
		case 1:
			_, err = io.ReadFull(d.transport, d.tmpbuffer[:1])
			if err != nil {
				return nil, err
			}
			ret[i] = DlmsData{Tag: TagError, Value: DlmsError{Result: AccessResultTag(d.tmpbuffer[0])}}
		default:
			return nil, fmt.Errorf("unexpected response tag: %x", d.tmpbuffer[0])
		}
	}

	return ret, nil
}
