package server

import (
	"encoding/binary"
	"fmt"
	"io"

	"log"

	"github.com/golang/protobuf/proto"
)

const (
	// MaxAllowedSize is maximum allowed size for the incoming message (not counting 4 bytes for length encoding)
	MaxAllowedSize = 1024 * 1024
)

// LengthEncodedProtoReaderWriter is a writer/reader that wraps length-encoded protobuff stream.
// This kind of stream has 2-part communication: first a length is sent as littlendian 4-byte integer
// and then the protobuff message is sent of that length
type LengthEncodedProtoReaderWriter struct {
	UnderlyingReadWriter io.ReadWriteCloser
	readBuffer           []byte
}

func (lep *LengthEncodedProtoReaderWriter) readSize() (size int, err error) {
	var preciseSize int32
	err = binary.Read(lep.UnderlyingReadWriter, binary.LittleEndian, &preciseSize)
	if err != nil {
		if err.Error() == io.EOF.Error() {
			return
		}
		err = fmt.Errorf("Failure reading length: %v", err)
		return
	}
	size = int(preciseSize)
	return
}

func (lep *LengthEncodedProtoReaderWriter) Read(data []byte) (n int, err error) {
	size, err := lep.readSize()
	if err != nil {
		return
	}
	if size > len(data) {
		err = fmt.Errorf("Provided slice too small. %v is the size of data, only %v provided", size, len(data))
		return
	}
	n, err = lep.UnderlyingReadWriter.Read(data)
	if err != nil {
		return
	}
	return
}

// ReadProto method allows direct reading of a protobuff object, with length as a prefix
func (lep *LengthEncodedProtoReaderWriter) ReadProto(msg proto.Message) (err error) {
	size, err := lep.readSize()
	if err != nil {
		return
	}
	if size > MaxAllowedSize {
		err = fmt.Errorf("Encoded size waiting on channel too big: %v", size)
		return
	} else if size > len(lep.readBuffer) {
		lep.readBuffer = make([]byte, size)
		log.Printf("Buffer resized to: %v", size)
	}
	n, err := lep.UnderlyingReadWriter.Read(lep.readBuffer)
	if err != nil {
		return
	}
	err = proto.Unmarshal(lep.readBuffer[:n], msg)
	if err != nil {
		err = fmt.Errorf("Could not unmarshal message: %v", err)
		return
	}
	return
}

func (lep *LengthEncodedProtoReaderWriter) Write(data []byte) (n int, err error) {
	var size int32
	size = int32(len(data))
	err = binary.Write(lep.UnderlyingReadWriter, binary.LittleEndian, size)
	if err != nil {
		err = fmt.Errorf("write bytes for length encoding failed: %v", err)
		return
	}
	return lep.UnderlyingReadWriter.Write(data)
}

// WriteProto method allows direct writing of a protobuff object, with length as a prefix
func (lep *LengthEncodedProtoReaderWriter) WriteProto(msg proto.Message) (err error) {
	marshalled, err := proto.Marshal(msg)
	if err != nil {
		return
	}
	_, err = lep.Write(marshalled)
	return
}
