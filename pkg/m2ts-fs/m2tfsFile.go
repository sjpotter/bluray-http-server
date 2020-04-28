package m2ts_fs

import (
	"context"
	"flag"
	"io"
	"k8s.io/klog"

	"bazil.org/fuse"
	"bazil.org/fuse/fs"

	"github.com/sjpotter/bluray-http-server/pkg/readers"
)

var (
	insertLang = flag.Bool("insert-lang", false, "Insert language tags")
)

type M2TSInfo struct {
	Name     string
	File     string
	Playlist int
}

var _ = fs.Node(&m2tsFile{})
var _ = fs.NodeOpener(&m2tsFile{})

type m2tsFile struct {
	path string
}

var _ = fs.Handle(&m2tsFileHandle{})
var _ = fs.HandleReleaser(&m2tsFileHandle{})
var _ = fs.HandleReader(&m2tsFileHandle{})

type m2tsFileHandle struct {
	f      readers.BluRayReader
	path   string
	offset int64
}

func getM2TSFile(path string) (fs.Node, error) {
	return &m2tsFile{path: path}, nil
}

func (m *m2tsFile) Attr(ctx context.Context, attr *fuse.Attr) error {
	err := _attr(m.path, attr)
	if err != nil {
		return err
	}

	klog.Infof("Attr: %v", m.path)

	rs, err := getM2TSRemuxer(m.path, *insertLang)
	if err != nil {
		return err
	}
	defer rs.Close()

	attr.Size = rs.Size()

	return err
}

func (m *m2tsFile) Open(ctx context.Context, req *fuse.OpenRequest, resp *fuse.OpenResponse) (fs.Handle, error) {
	klog.Infof("Open: %v by %v", m.path, req.Pid)
	rs, err := getM2TSRemuxer(m.path, *insertLang)
	if err != nil {
		return nil, err
	}

	return &m2tsFileHandle{f: rs, path: m.path}, nil
}

func (m m2tsFileHandle) Release(ctx context.Context, req *fuse.ReleaseRequest) error {
	klog.Infof("Release: %v", m.path)
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
