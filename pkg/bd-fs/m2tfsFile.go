package bd_fs

import (
	"bazil.org/fuse"
	"bazil.org/fuse/fs"
	"context"
	"github.com/sjpotter/bluray-http-server/pkg/readers"
	"io"
	"time"
)

var _ = fs.Node(&m2tsFile{})
var _ = fs.NodeOpener(&m2tsFile{})

type m2tsFile struct {
	device string
	playlist int
	size uint64
}

var _ = fs.Handle(&m2tsFileHandle{})
var _ = fs.HandleReleaser(&m2tsFileHandle{})
var _ = fs.HandleReader(&m2tsFileHandle{})

type m2tsFileHandle struct {
	f      readers.BluRayReader
	offset int64
}

func getM2TSFile(device string, playlist int) (fs.Node, error) {
	return &m2tsFile{device: device, playlist: playlist}, nil
}

func (m *m2tsFile) getM2TSRemuxer() (*readers.M2TSRemuxer, error) {
	return readers.NewM2TSRemuxer(m.device, m.playlist)
}

func (m *m2tsFile) Attr(ctx context.Context, attr *fuse.Attr) error {
	attr.Size = m.size
	attr.Mode = 0644
	attr.Valid = 24*7*365*time.Hour

	return nil
}

func (m *m2tsFile) Open(ctx context.Context, req *fuse.OpenRequest, resp *fuse.OpenResponse) (fs.Handle, error) {
	rs, err := m.getM2TSRemuxer()
	if err != nil {
		return nil, err
	}

	return &m2tsFileHandle{f: rs}, nil
}

func (m m2tsFileHandle) Release(ctx context.Context, req *fuse.ReleaseRequest) error {
	m.f.Close()

	return nil
}

func (m *m2tsFileHandle) Read(ctx context.Context, req *fuse.ReadRequest, resp *fuse.ReadResponse) error {
	var err error
	if req.Offset != m.offset {
		m.offset, err = m.f.Seek(req.Offset, io.SeekStart)
		if err != nil {
			return err
		}
	}

	buf := make([]byte, req.Size)
	n, err := m.f.Read(buf)
	m.offset += int64(n)

	resp.Data = buf

	return err
}