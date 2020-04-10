package pmtparser

import (
	"encoding/binary"
	"fmt"
)

type component struct {
	streamType  uint8
	pid         []byte //really 13 bits
	descriptors []*descriptor
}

func (c *component) parseDescriptors(payload []byte) int {
	descriptorLen := int(payload[0])
	if *pmtDebug {
		fmt.Printf("descriptors:\n")
	}

	pos := 1
	for pos < descriptorLen {
		d := &descriptor{}
		d.tag = payload[pos]
		pos++
		dSize := int(payload[pos])
		pos++
		d.data = payload[pos : pos+dSize]
		pos += dSize
		c.appendDescriptor(d)
		if *pmtDebug {
			fmt.Printf("\ttag = % #x, size = %v, data = % #x\n", d.tag, dSize, d.data)
		}
	}

	return 1 + descriptorLen
}

func (c *component) output() []byte {
	descriptors := []byte{}
	for _, d := range c.descriptors {
		descriptors = append(descriptors, d.output()...)
	}

	ret := []byte{c.streamType}
	ret = append(ret, c.pid...)
	ret = append(ret, 0xF0)
	ret = append(ret, byte(len(descriptors)))
	ret = append(ret, descriptors...)

	return ret
}

func (c *component) getPid() uint16 {
	return binary.BigEndian.Uint16(c.pid)
}

func (c *component) appendDescriptor(d *descriptor) {
	c.descriptors = append(c.descriptors, d)
}
