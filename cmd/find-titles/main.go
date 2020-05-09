package main

/*
#cgo pkg-config: libbluray, udfread
#include <stdlib.h>

#include <udfread/udfread.h>
#include <libbluray/bluray.h>
*/
import "C"
import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

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

	anyDVDPlaylists := findTitles(iso)
	if len(anyDVDPlaylists) > 0 {
		fmt.Printf("Discovered %v playlists from AnyDVD's disc.inf\n", len(anyDVDPlaylists))
		for _, p := range anyDVDPlaylists {
			writePlaylist(name, p, dir)
		}
		return
	}

	bd := C.bd_open(C.CString(*iso), nil)
	if bd == nil {
		fmt.Printf("Error opening %s\n", *iso)
		os.Exit(1)
	}
	defer C.bd_close(bd)

	numTitles := C.bd_get_titles(bd, C.TITLES_RELEVANT, 0)

	var playlists = make(map[int]*C.struct_bd_title_info)

	for i := 0; i < int(numTitles); i++ {
		ti := C.bd_get_title_info(bd, C.uint(i), 0)
		if ti == nil {
			continue
		}
		if (int64(ti.duration) / 90000) / 60 > *minLength {
			playlists[i] = ti
		} else {
			C.bd_free_title_info(ti)
		}
	}

	if len(playlists) > 1 && *findOne {
		fmt.Printf("%v: found %v titles over %v minutes long\n", *iso, len(playlists), *minLength)
		for title, ti := range playlists {
			len := ti.duration / 90000
			hours := len / 3600
			len %= 3600
			minutes := len / 60
			secs := len % 60
			fmt.Printf("\t%03d(%05d) - %02d:%02d:%02d\n", title, ti.playlist, hours, minutes, secs)

			C.bd_free_title_info(ti)
		}
		return
	}

	for playlist, _ := range playlists {
		if writePlaylist(name, playlist, dir) {
			return
		}
	}
}

func writePlaylist(fileName string, playlist int, dir string) bool {
	info := m2ts_fs.M2TSInfo{
		File:     fileName,
		Playlist: int(playlist),
	}

	data, err := yaml.Marshal(info)
	if err != nil {
		fmt.Printf("failed to marshal %+v: %v\n", info, err)
		return true
	}

	m2tsFileName := fmt.Sprintf("%v-%05d.m2ts", fileName, playlist)
	m2tsFileName = filepath.Join(dir, m2tsFileName)
	f, err := os.Create(m2tsFileName)
	if err != nil {
		fmt.Printf("failed to create %v: %v\n", m2tsFileName, err)
		return true
	}

	f.Write(data)
	f.Close()
	return false
}

func findTitles(iso *string) []int {
	udfread := C.udfread_init()
	if udfread == nil {
		fmt.Printf("couldn't initalize udfread\n")
		return nil
	}

	ret := C.udfread_open(udfread, C.CString(*iso))
	if ret != 0 {
		fmt.Printf("failed to open %v\n", *iso)
		return nil
	}
	defer C.udfread_close(udfread)

	udfRoot := C.udfread_opendir(udfread, C.CString("/"))
	if udfRoot == nil {
		fmt.Printf("failed to get root directory entry\n")
		os.Exit(1)
	}

	dirent := &C.struct_udfread_dirent{}
	for {
		dirent := C.udfread_readdir(udfRoot, dirent)
		if dirent == nil {
			break
		}
		switch dirent.d_type {
		case C.UDF_DT_REG:
			name := C.GoString(dirent.d_name)
			fmt.Printf("\tFile: %+v\n", name)
			if name == "disc.inf" {
				data := dumpFile(udfread, "disc.inf")
				if data != nil {
					return getPlaylists(data)
				}
				return nil
			}
		}
	}

	return nil
}

func dumpFile(udfread *C.udfread, s string) *string {
	fh := C.udfread_file_open(udfread, C.CString(s))
	if fh == nil {
		fmt.Printf("failed to open %v\n", s)
		return nil
	}
	defer C.udfread_file_close(fh)

	data := C.malloc(4096)
	defer C.free(data)

	size := C.udfread_file_read(fh, data, 4096)
	if size <= 0 {
		fmt.Printf("file_read failed\n")
		return nil
	}
	str := C.GoString((*C.char)(data))
	fmt.Println(str)

	return &str
}

func getPlaylists(data *string) []int {
	var ret []int
	lines := strings.Split(*data, "\r\n")
	for _, line := range lines {
		kv := strings.Split(line, "=")
		if len(kv) != 2 {
			continue
		}

		if kv[0] != "playlists" {
			continue
		}

		playlists := strings.Split(kv[1], ",")
		for _, playlist := range playlists {
			p, err := strconv.Atoi(playlist)
			if err != nil {
				fmt.Printf("couldn't convert %v to an int: %v", playlist, err)
				continue
			}
			ret = append(ret, p)
		}

		return ret
	}

	return nil
}
