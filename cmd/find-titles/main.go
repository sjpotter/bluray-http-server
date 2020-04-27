package main

/*
#cgo pkg-config: libbluray
#include <stdlib.h>

#include <libbluray/bluray.h>
*/
import "C"
import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/goccy/go-yaml"

	m2ts_fs "github.com/sjpotter/bluray-http-server/pkg/m2ts-fs"
)

var (
	iso = flag.String("iso", "", "iso file to inspect")
	minLength = flag.Int64("minimum length (minutes)", 60, "minimum length to enumerate")
	findOne   = flag.Bool("find-one", false, "only enumerate one")
)

func main() {
	flag.Parse()

	if *iso == "" {
		flag.Usage()
		return
	}

	dir := filepath.Dir(*iso)
	name := filepath.Base(*iso)

	bd := C.bd_open(C.CString(*iso), nil)
	if bd == nil {
		fmt.Printf("Error opening %s\n", *iso)
		os.Exit(1)
	}
	defer C.bd_close(bd)

	numTitles := C.bd_get_titles(bd, C.TITLES_RELEVANT, 0)

	var playlists = make(map[int]int64)

	for i := 0; i < int(numTitles); i++ {
		ti := C.bd_get_title_info(bd, C.uint(i), 0)
		if ti == nil {
			continue
		}
		if (int64(ti.duration) / 90000) / 60 > *minLength {
			playlists[int(ti.playlist)] = int64(ti.duration) / 90000
		}
		C.bd_free_title_info(ti)
	}

	if len(playlists) > 1 && *findOne {
		fmt.Printf("%v: found %v titles over %v minutes long\n", *iso, len(playlists), *minLength)
		for playlist, len := range playlists {
			hours := len / 3600
			len %= 3600
			minutes := len / 60
			secs := len % 60
			fmt.Printf("\t%05d - %02d:%02d:%02d\n", playlist, hours, minutes, secs)
		}
		return
	}

	for playlist, _ := range playlists {
		info := m2ts_fs.M2TSInfo{
			File: name,
			Playlist: int(playlist),
		}

		data, err := yaml.Marshal(info)
		if err != nil {
			fmt.Printf("failed to marshal %+v: %v\n", info, err)
			return
		}

		m2tsFileName := fmt.Sprintf("%v-%05d.m2ts", name, playlist)
		m2tsFileName = filepath.Join(dir, m2tsFileName)
		f, err := os.Create(m2tsFileName)
		if err != nil {
			fmt.Printf("failed to create %v: %v\n", m2tsFileName, err)
			return
		}

		f.Write(data)
		f.Close()
	}
}
