package m2ts_fs

import (
	"context"
	"io"
	"os"

	"bazil.org/fuse"
	"bazil.org/fuse/fs"
)

var _ = fs.Node(&passthrough{})
var _ = fs.NodeOpener(&passthrough{})

type passthrough struct {
	path string
}

var _ = fs.Handle(&plainFileHandle{})
var _ = fs.HandleReleaser(&plainFileHandle{})
var _ = fs.HandleReader(&plainFileHandle{})

type plainFileHandle struct {
	f      *os.File
	offset int64
}

func getPassthroughFile(path string) (fs.Node, error) {
	return &passthrough{path: path}, nil
}

func (p *passthrough) Attr(ctx context.Context, attr *fuse.Attr) error {
	err := _attr(p.path, attr)

	return err
}

func (f *passthrough) Open(ctx context.Context, req *fuse.OpenRequest, resp *fuse.OpenResponse) (fs.Handle, error) {
	r, err := os.Open(f.path)
	if err != nil {
		return nil, err
	}

	return &plainFileHandle{f: r}, nil
}

func (p plainFileHandle) Release(ctx context.Context, req *fuse.ReleaseRequest) error {
	return p.f.Close()
}

func (p *plainFileHandle) Read(ctx context.Context, req *fuse.ReadRequest, resp *fuse.ReadResponse) error {
	var err error
	if req.Offset != p.offset {
		p.offset, err = p.f.Seek(req.Offset, io.SeekStart)
		if err != nil {
			return err
		}
	}

	buf := make([]byte, req.Size)
	n, err := p.f.Read(buf)
	p.offset += int64(n)

	resp.Data = buf

	return err
}