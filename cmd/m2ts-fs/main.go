package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"bazil.org/fuse"
	"bazil.org/fuse/fs"
	"bazil.org/fuse/fs/fstestutil"

	"k8s.io/klog"

	m2ts_fs "github.com/sjpotter/bluray-http-server/pkg/m2ts-fs"
	_ "github.com/sjpotter/bluray-http-server/pkg/remote-control"
)

var (
	progName = filepath.Base(os.Args[0])
	debug = flag.Bool("debug", false, "verbose fuse debugging")
	port  = flag.String("port", "8765", "port for rc to listne on")
)

func usage() {
	fmt.Fprintf(os.Stderr, "Usage of %s:\n", progName)
	fmt.Fprintf(os.Stderr, "  %s [-debug] STORAGE_DIR MOUNTPOINT\n", progName)
	flag.PrintDefaults()
}

func main() {
	defer klog.Flush()

	log.SetFlags(0)
	log.SetPrefix(progName + ": ")

	flag.Usage = usage
	flag.Parse()

	if flag.NArg() != 2 {
		usage()
		os.Exit(2)
	}

	server := &http.Server{
		Addr: fmt.Sprintf("localhost:%v", *port),
	}

	go func() {
		err := server.ListenAndServe()
		if err != nil {
			klog.Errorf("rc server failed: %v", err)
		}
	}()

	if *debug {
		fstestutil.DebugByDefault()
	}

	path := flag.Arg(0)
	mountpoint := flag.Arg(1)

	klog.Infof("Starting Up")

	if err := mount(path, mountpoint); err != nil {
		log.Fatal(err)
	}

	server.Shutdown(context.Background())
}

func mount(path, mountpoint string) error {
	err := os.Chdir(path)
	if err != nil {
		return err
	}

	c, err := fuse.Mount(mountpoint, fuse.AllowOther())
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

