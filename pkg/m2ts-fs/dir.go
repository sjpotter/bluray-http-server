package m2ts_fs

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"bazil.org/fuse"
	"bazil.org/fuse/fs"
)

var _ = fs.Node(&dir{})
var _ = fs.NodeRequestLookuper(&dir{})
var _ = fs.HandleReadDirAller(&dir{})
var _ = fs.NodeCreater(&dir{})

type dir struct {
	path string
}

func (d *dir) Attr(ctx context.Context, attr *fuse.Attr) error {
	return _attr(d.path, attr)
}

func _attr(path string, attr *fuse.Attr) error {
	info, err := os.Stat(path)
	if err != nil {
		return err
	}

	stat, ok := info.Sys().(*syscall.Stat_t)
	if !ok {
		panic("Not what I expected")
	}

	attr.Inode = stat.Ino
	attr.Size = uint64(stat.Size)
	attr.Blocks = uint64(stat.Blocks)
	attr.Atime = time.Unix(stat.Atim.Unix())
	attr.Mtime = time.Unix(stat.Mtim.Unix())
	attr.Ctime = time.Unix(stat.Ctim.Unix())
	attr.Mode = info.Mode()
	attr.Nlink = uint32(stat.Nlink)
	attr.Uid = stat.Uid
	attr.Gid = stat.Gid

	return nil
}

func (d *dir) Lookup(ctx context.Context, req *fuse.LookupRequest, resp *fuse.LookupResponse) (fs.Node, error) {
	if strings.HasSuffix(req.Name, ".iso") || strings.HasSuffix(req.Name, ".ISO") {
		return nil, syscall.ENOENT
	}

	dirEntName := filepath.Join(d.path, req.Name)

	var child fs.Node

	stat, err := os.Stat(dirEntName)
	if err != nil {
		return nil, err
	}

	switch stat.IsDir() {
	case true:
		child = &dir{path: dirEntName}
	case false:
		if strings.HasSuffix(req.Name, ".m2ts") {
			child, err = getM2TSFile(dirEntName)
		} else {
			child, err = getPassthroughFile(dirEntName)
		}
	}

	return child, err
}

func (d *dir) ReadDirAll(ctx context.Context) ([]fuse.Dirent, error) {
	var ret []fuse.Dirent

	dirents, err := ioutil.ReadDir(d.path)
	if err != nil {
		return nil, err
	}

	for _, d := range dirents {
		if strings.HasSuffix(d.Name(), ".iso") || strings.HasSuffix(d.Name(), ".ISO") {
			continue
		}

		stat, ok := d.Sys().(*syscall.Stat_t)
		if !ok {
			panic("Not what I expected")
		}
		var de = fuse.Dirent{
			Inode: stat.Ino,
			Name:  d.Name(),
		}

		switch (stat.Mode & syscall.S_IFMT) {
		case syscall.S_IFDIR:
			de.Type = fuse.DT_Dir
		case syscall.S_IFLNK:
			de.Type = fuse.DT_Link
		case syscall.S_IFREG:
			de.Type = fuse.DT_File
		default:
			fmt.Printf("%v is not recognized type: %#x\n", de.Name, stat.Mode & syscall.S_IFMT)
			continue
		}

		ret = append(ret, de)
	}

	return ret, nil
}

func (d *dir) Create(ctx context.Context, req *fuse.CreateRequest, resp *fuse.CreateResponse) (fs.Node, fs.Handle, error) {
	panic("implement me")
}