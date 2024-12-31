package tcpserver

import (
	"encoding/binary"
	"fmt"
	"io"

	"github.com/spf13/cast"
	"github.com/wuqunyong/file_storage/pkg/easytcp"
)

// PBPacker treats packet as:
//
// totalSize(4)|idSize(2)|id(n)|data(n)
//
// | segment     | type   | size    | remark                |
// | ----------- | ------ | ------- | --------------------- |
// | `totalSize` | uint32 | 4       | the whole packet size |
// | `idSize`    | uint16 | 2       | length of id          |
// | `id`        | string | dynamic |                       |
// | `data`      | []byte | dynamic |                       |
type PBPacker struct{}

func NewPBPacker() *PBPacker {
	return &PBPacker{}
}

func (p *PBPacker) bytesOrder() binary.ByteOrder {
	return binary.LittleEndian
}

func (p *PBPacker) Pack(msg *easytcp.Message) ([]byte, error) {
	dataSize := len(msg.Data())

	iHead := 4 + 1 + 1 + 2 + 4 + 4
	buffer := make([]byte, iHead+dataSize)
	p.bytesOrder().PutUint32(buffer[:4], 0)
	buffer[4] = 0
	buffer[5] = 0
	id, err := cast.ToUint16E(msg.ID())
	if err != nil {
		return nil, fmt.Errorf("invalid type of msg.ID: %s", err)
	}
	p.bytesOrder().PutUint16(buffer[6:8], id)
	p.bytesOrder().PutUint32(buffer[8:12], uint32(dataSize))
	p.bytesOrder().PutUint32(buffer[12:16], 0)
	copy(buffer[iHead:], msg.Data()) // write data
	return buffer, nil
}

func (p *PBPacker) Unpack(reader io.Reader) (*easytcp.Message, error) {
	headerBuff := make([]byte, 4+1+1+2+4+4)
	if _, err := io.ReadFull(reader, headerBuff); err != nil {
		if err == io.EOF {
			return nil, err
		}
		return nil, fmt.Errorf("read header err: %s", err)
	}
	id := int(p.bytesOrder().Uint16(headerBuff[6:8]))        // read totalSize
	bodySize := int(p.bytesOrder().Uint32(headerBuff[8:12])) // read idSize

	bodyBuff := make([]byte, bodySize)
	if _, err := io.ReadFull(reader, bodyBuff); err != nil {
		return nil, fmt.Errorf("read body err: %s", err)
	}

	// ID is a string, so we should use a string-type id to register routes.
	// eg: server.AddRoute("string-id", handler)
	msg := easytcp.NewMessage(id, bodyBuff)
	msg.Set("fullSize", bodySize)
	return msg, nil
}
