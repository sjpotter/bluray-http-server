package pmtparser

import (
	"encoding/binary"
	"flag"
	"fmt"

	"github.com/sjpotter/bluray-http-server/pkg/utils"
)

const (
	SyncByte = 0x47
	PmtStart = 0x41
	PaktCont = 0x01
	TableId  = 0x2
	LANG_DSC = 0xA
)

var (
	pmtDebug = flag.Bool("pmtDebug", false, "Debug Output")
)

type PMT struct {
	timestamp     []byte
	continuityBit uint8
	preProgram    []byte
	program       []byte
	components    []*component
}

func (p *PMT) Output() []byte {
	var payload []byte

	payload = append(payload, p.preProgram...)
	payload = append(payload)
	payload = append(payload, byte(len(p.program)))
	payload = append(payload, p.program...)
	for _, c := range p.components {
		payload = append(payload, c.output()...)
	}

	packetData := []byte{2}

	size := len(payload) + 4
	size = size | 0xb000

	sizeBytes := make([]byte, 2)
	binary.BigEndian.PutUint16(sizeBytes, uint16(size))

	packetData = append(packetData, sizeBytes...)

	packetData = append(packetData, payload...)
	crc := utils.CalcCRC(packetData)
	packetData = append(packetData, crc...)

	var output []byte

	output = append(output, p.timestamp...)
	output = append(output, 0x47, 0x41, 0x00, 0x10|p.continuityBit)
	p.continuityBit = (p.continuityBit + 1) % 10
	output = append(output, 00)
	count := 9
	for _, b := range packetData {
		if count%192 == 0 {
			output = append(output, p.timestamp...)
			count += 4
			output = append(output, 0x47, 0x01, 0x00, 0x10|p.continuityBit)
			p.continuityBit = (p.continuityBit + 1) % 10
			count += 4
		}
		output = append(output, b)
		count++
	}

	for count%192 != 0 {
		output = append(output, 0xFF)
		count++
	}

	return output
}

func (p *PMT) loadPayload(payload []byte) {
	pos := 0
	p.preProgram = payload[pos : pos+8] // cheating a little as program size is part of byte 8
	pos += 8
	pos += p.parseProgram(payload[pos:])
	for pos < len(payload)-4 {
		pos += p.parseComponent(payload[pos:])
	}
}

func (p *PMT) parseProgram(payload []byte) int {
	size := payload[0]
	p.program = payload[1 : size+1]

	return int(size) + 1
}

func (p *PMT) parseComponent(payload []byte) int {
	c := &component{
		streamType: payload[0],
		pid:        payload[1:3], //really 13 bits
	}

	if *pmtDebug {
		fmt.Printf("pid bits = % #x\n", c.pid)
	}

	pid := int(binary.BigEndian.Uint16(c.pid)) & 0x1FFF

	if *pmtDebug {
		fmt.Printf("streamType: % #x, pid = % #x\n", c.streamType, pid)
	}

	descriptorLen := c.parseDescriptors(payload[4:])

	p.components = append(p.components, c)

	return 4 + descriptorLen
}

func (p *PMT) SetLanguage(pid uint16, lang string) {
	var seenPIDs []uint16
	for _, c := range p.components {
		cPid := binary.BigEndian.Uint16(c.pid) & 0x1FFF
		if cPid == pid {
			for _, d := range c.descriptors {
				if d.tag == LANG_DSC {
					if cap(d.data) <= len(lang) {
						copy(d.data, lang)
					} else {
						data := make([]byte, 4)
						copy(data, lang)
						for i := len(lang); i < 4; i++ {
							data[i] = 0
						}
						d.data = data
						return
					}
				}
			}

			// didn't find a descriptor for language.  need to add one
			data := make([]byte, 4)
			copy(data, lang)
			for i := len(lang); i < 4; i++ {
				data[i] = 0
			}
			c.descriptors = append(c.descriptors, &descriptor{tag: LANG_DSC, data: data})

			return
		}
		seenPIDs = append(seenPIDs, cPid)
	}

	fmt.Printf("Didn't match an existing component to pid %v in %+v\n", pid, seenPIDs)
}

func ParsePMTPackets(data []byte) (*PMT, int, error) {
	ret := &PMT{}

	if len(data)%192 != 0 {
		return nil, 0, fmt.Errorf("not an even packet number of bytes")
	}

	if data[4] != SyncByte {
		return nil, 0, fmt.Errorf("data isn't aligned, byte 4 (%#x) != %#x", data[4], SyncByte)
	}

	ret.timestamp = data[:4]

	if data[5] != PmtStart {
		return nil, 0, fmt.Errorf("first packet isn't a PMT, byte 5 (%#x) != %#x", data[5], PmtStart)
	}

	if data[6] != 0 {
		return nil, 0, fmt.Errorf("expected byte 6 (%#x) to be 0", data[6])
	}

	ret.continuityBit = data[7] & 0xF

	if *pmtDebug {
		fmt.Printf("continuityBits = %#x\n", ret.continuityBit)
	}

	if data[8] != 0 {
		return nil, 0, fmt.Errorf("expected byte 8 - Pointer field (%#x) to be 0", data[8])
	}

	if data[9] != TableId {
		return nil, 0, fmt.Errorf("expected byte 9 - Table id (%#x) to be %#x", data[9], TableId)
	}

	pmtLen := binary.BigEndian.Uint16([]byte{data[10], data[11]}) & 0xFFF

	if *pmtDebug {
		fmt.Printf("length of PMT data = %v\n", pmtLen)
	}
	pos := 12

	var payload []byte
	first := true

	for pmtLen != 0 {
		var maxBytes uint16 = 184
		if first {
			maxBytes = 184 - 4
			first = false
		}

		if pmtLen > maxBytes {
			payload = append(payload, data[pos:pos+int(maxBytes)]...)
			pmtLen -= maxBytes
			pos += int(maxBytes) + 8
		} else {
			payload = append(payload, data[pos:pos+int(pmtLen)]...)
			pos += int(pmtLen)
			pmtLen = 0
		}
	}

	ret.loadPayload(payload)

	partial := (12+int(pmtLen))%192 != 0
	packets := (12 + int(pmtLen)) / 192
	if partial {
		packets++
	}

	return ret, packets, nil
}
