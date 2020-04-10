package pmtparser

type descriptor struct {
	tag  uint8
	data []byte
}

func (d *descriptor) output() []byte {
	ret := []byte{d.tag, byte(len(d.data))}
	ret = append(ret, d.data...)

	return ret
}
