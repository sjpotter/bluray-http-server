package main

import (
	"flag"
	"fmt"
	m2ts_fs "github.com/sjpotter/bluray-http-server/pkg/m2ts-fs"
	"log"
	"os"
	"path/filepath"

	"bazil.org/fuse"
	"bazil.org/fuse/fs"
)

var progName = filepath.Base(os.Args[0])

func usage() {
	fmt.Fprintf(os.Stderr, "Usage of %s:\n", progName)
	fmt.Fprintf(os.Stderr, "  %s STORAGE_DIR MOUNTPOINT\n", progName)
	flag.PrintDefaults()
}

func main() {
	log.SetFlags(0)
	log.SetPrefix(progName + ": ")

	flag.Usage = usage
	flag.Parse()

	if flag.NArg() != 2 {
		usage()
		os.Exit(2)
	}

	path := flag.Arg(0)
	mountpoint := flag.Arg(1)

	if err := mount(path, mountpoint); err != nil {
		log.Fatal(err)
	}
}

func mount(path, mountpoint string) error {
	err := os.Chdir(path)
	if err != nil {
		return err
	}

	c, err := fuse.Mount(mountpoint)
	if err != nil {
		return err
	}
	defer c.Close()

	filesys := &m2ts_fs.FS{}

	if err := fs.Serve(c, filesys); err != nil {
		return err
	}

	return nil
}