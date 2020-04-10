package readers

import (
	"fmt"
	"io"

	"github.com/sjpotter/bluray-http-server/pkg/pmtparser"
	"github.com/sjpotter/bluray-http-server/pkg/types"
)

type M2TSRemuxer struct {
	b               *BDReadSeeker
	info            *types.BDTitle
	editPMT         *pmtparser.PMT
	origPacketCount int
	deltaSize       int
	edited          bool
}

// Generates the packet at setup time to minimize speed disruption during reading
func NewM2TSRemuxer(b *BDReadSeeker) (*M2TSRemuxer, error) {
	info, err := b.ParseTile()
	if err != nil {
		return nil, err
	}

	buffer := make([]byte, 192*200)

	n, err := b.Read(buffer)
	if err != nil {
		return nil, err
	}

	b.Seek(0, io.SeekStart)

	pos := 0
	packetNum := 0

	for pos < n && pos+192 < len(buffer) {
		if buffer[pos+4] == 0x47 {
			if buffer[pos+5] == 0x41 && buffer[pos+6] == 00 {
				pmt, packets, err := pmtparser.ParsePMTPackets(buffer[pos:])
				if err != nil {
					return nil, err
				}

				for k, v := range info.Audio {
					pmt.SetLanguage(uint16(k), v.AudioLang)
				}
				for k, v := range info.PG {
					pmt.SetLanguage(uint16(k), v.PGLang)
				}

				deltaSize := len(pmt.Output()) - (packets * 192)

				return &M2TSRemuxer{
					b:               b,
					info:            info,
					editPMT:         pmt,
					origPacketCount: packets,
					deltaSize:       deltaSize,
				}, nil
			}
		}

		packetNum++
		pos = packetNum * 192
	}

	return nil, fmt.Errorf("didn't find the PMT packet")
}

// reads buffer, if already modified PMT packet once, just return
// if haven't read in PMT packet, see if its in ths block and overwrite it with language added version
func (m *M2TSRemuxer) Read(p []byte) (int, error) {
	n, err := m.b.Read(p)
	if err != nil {
		return n, err
	}

	if m.edited {
		return n, err
	}

	packetNum := 0
	pos := packetNum * 192

	for pos < n && pos+192 < len(p) {
		if p[pos+4] == 0x47 && p[pos+5] == 0x41 && p[pos+6] == 00 {
			m.edited = true

			return m.overwritePMT(p, pos), nil
		}

		packetNum++
		pos = packetNum * 192
	}

	return n, err
}

// fakes out "seek end" due to added bytes of course, one could break it (SeekEnd with negative bytes and won't be
// accurate, but simplifying for now
func (m *M2TSRemuxer) Seek(offset int64, whence int) (int64, error) {
	n, err := m.b.Seek(offset, whence)
	if err != nil {
		return n, err
	}

	if whence == io.SeekEnd && offset == 0 {
		n += int64(m.deltaSize)
	}

	return n, err
}

// Doesn't do anything yet, needed if part of the readercloser interface
func (m *M2TSRemuxer) Close() {
	return
}

// This an probably be done more efficiently, but it's clear enough
// we are basically inserting the "modified" PMT into the current block of bytes, overwriting the existing block
// which might entail pushing the rest of the bytes down a bit.
// as the buffer is fixed size, will have to remove the last few bytes - will be reread due to the backwards seek
func (m *M2TSRemuxer) overwritePMT(buffer []byte, pos int) int {
	prePMT := buffer[:pos]
	pmt := m.editPMT.Output()
	postPMT := buffer[pos+m.origPacketCount*192 : len(buffer)-m.deltaSize]
	m.b.Seek(int64(-m.deltaSize), io.SeekCurrent) // to reread the bytes we cut off above

	var output []byte
	output = append(output, prePMT...)
	output = append(output, pmt...)
	output = append(output, postPMT...)
	copy(buffer, output)

	return len(output)
}
