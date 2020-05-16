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
	"unsafe"

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

	switch len(playlists) {
	case 0:
	case 2:
		if *findOne {
			var ti1 *C.struct_bd_title_info
			var ti2 *C.struct_bd_title_info
			for _, v := range playlists {
				if ti1 == nil {
					ti1 = v
				} else {
					ti2 = v
				}
			}
			if areEqual(ti1, ti2) {
				if ti1.chapter_count != ti2.chapter_count {
					newPlaylists := make(map[int]*C.struct_bd_title_info)
					if ti1.chapter_count > ti2.chapter_count {
						fmt.Printf("Picking %v\n", ti1.playlist)
						newPlaylists[int(ti1.playlist)] = ti1
					} else {
						fmt.Printf("Picking %v\n", ti2.playlist)
						newPlaylists[int(ti2.playlist)] = ti2
					}
					playlists = newPlaylists
					break
				}
			}
		}

		fallthrough
	default:
		if *findOne {
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
	}

	if len(playlists) > 1 && *findOne {
	}

	for playlist, _ := range playlists {
		if writePlaylist(name, playlist, dir) {
			return
		}
	}
}

func areEqual(ti1 *C.struct_bd_title_info, ti2 *C.struct_bd_title_info) bool {
	if ti1.duration != ti2.duration {
		return false
	}

	if ti1.clip_count == 0 || ti2.clip_count == 0 || ti1.clip_count != ti2.clip_count {
		return false
	}

	t1clips := (*[1 << 30]C.struct_bd_clip)(unsafe.Pointer(ti1.clips))[:ti1.clip_count:ti1.clip_count]
	t2clips := (*[1 << 30]C.struct_bd_clip)(unsafe.Pointer(ti2.clips))[:ti2.clip_count:ti2.clip_count]

	return clipsEqual(t1clips, t2clips)
}

func clipsEqual(c1 []C.struct_bd_clip, c2 []C.struct_bd_clip) bool {
	for i, _ := range c1 {
		c1p := c1[i]
		c2p := c2[i]
		if c1p.pkt_count != c2p.pkt_count {
			return false
		}
		if c1p.still_mode != c2p.still_mode {
			return false
		}
		if c1p.still_time != c2p.still_time {
			return false
		}
		if c1p.video_stream_count != c2p.video_stream_count {
			return false
		}
		if c1p.pg_stream_count != c2p.pg_stream_count {
			return false
		}
		if c1p.ig_stream_count != c2p.ig_stream_count {
			return false
		}
		if c1p.sec_audio_stream_count != c2p.sec_audio_stream_count {
			return false
		}
		if c1p.sec_video_stream_count != c2p.sec_video_stream_count {
			return false
		}

		c1pV := (*[1 << 30]C.BLURAY_STREAM_INFO)(unsafe.Pointer(c1p.video_streams))[:c1p.video_stream_count:c1p.video_stream_count]
		c2pV := (*[1 << 30]C.BLURAY_STREAM_INFO)(unsafe.Pointer(c2p.video_streams))[:c2p.video_stream_count:c2p.video_stream_count]

		if !compareStreamInfos(c1pV, c2pV) {
			return false
		}

		c1pA := (*[1 << 30]C.BLURAY_STREAM_INFO)(unsafe.Pointer(c1p.audio_streams))[:c1p.audio_stream_count:c1p.audio_stream_count]
		c2pA := (*[1 << 30]C.BLURAY_STREAM_INFO)(unsafe.Pointer(c2p.audio_streams))[:c2p.audio_stream_count:c2p.audio_stream_count]

		if !compareStreamInfos(c1pA, c2pA) {
			return false
		}

		c1pI := (*[1 << 30]C.BLURAY_STREAM_INFO)(unsafe.Pointer(c1p.ig_streams))[:c1p.ig_stream_count:c1p.audio_stream_count]
		c2PI := (*[1 << 30]C.BLURAY_STREAM_INFO)(unsafe.Pointer(c2p.ig_streams))[:c2p.ig_stream_count:c2p.audio_stream_count]

		if !compareStreamInfos(c1pI, c2PI) {
			return false
		}
	}

	return true
}

func compareStreamInfos(si1 []C.BLURAY_STREAM_INFO, si2 []C.BLURAY_STREAM_INFO) bool {
	for i, _ := range si1 {
		si1p := si1[i]
		si2p := si2[i]

		if si1p.coding_type != si2p.coding_type {
			return false
		}

		if si1p.format != si2p.format {
			return false
		}

		if si1p.rate != si2p.rate {
			return false
		}

		if si1p.char_code != si2p.char_code {
			return false
		}

		if si1p.lang[0] != si2p.lang[0] || si1p.lang[1] != si2p.lang[1] || si1p.lang[2] != si2p.lang[2] || si1p.lang[3] != si2p.lang[3] {
			return false
		}

		if si1p.pid != si2p.pid {
			return false
		}

		if si1p.aspect != si2p.aspect {
			return false
		}

		if si1p.subpath_id != si2p.subpath_id {
			return false
		}
	}

	return true
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
