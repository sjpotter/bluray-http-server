package main

import (
	"flag"
	"fmt"
	bd_fs "github.com/sjpotter/bluray-http-server/pkg/bd-fs"
	"log"
	"os"
	"path/filepath"
	"syscall"
	"time"

	"bazil.org/fuse"
	"bazil.org/fuse/fs"
	"bazil.org/fuse/fs/fstestutil"
)

var (
	progName = filepath.Base(os.Args[0])
	debug = flag.Bool("debug", false, "verbose fuse debugging")
)

func usage() {
	fmt.Fprintf(os.Stderr, "Usage of %s:\n", progName)
	fmt.Fprintf(os.Stderr, "  %s [-debug] <block device> <mount point>\n", progName)
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

	go func() {
		for {
			time.Sleep(5*time.Second)
			pid, err := syscall.Wait4(-1, nil, 0, nil)
			if err == nil {
				fmt.Printf("wait4 returned pid: %v\n", pid)
			}
		}
	}()

	if *debug {
		fstestutil.DebugByDefault()
	}

	device := flag.Arg(0)
	mountpoint := flag.Arg(1)

	if err := mount(device, mountpoint); err != nil {
		log.Fatal(err)
	}
}

func mount(device, mountpoint string) error {
	c, err := fuse.Mount(mountpoint, fuse.AllowOther())
	if err != nil {
		return err
	}
	defer c.Close()

	filesys := &bd_fs.FS{Device: device}

	if err := fs.Serve(c, filesys); err != nil {
		return err
	}

	return nil
}