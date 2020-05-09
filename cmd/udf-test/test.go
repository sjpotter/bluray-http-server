package main

/*
#cgo pkg-config: udfread
#include <stdlib.h>

#include <udfread/udfread.h>
 */
import "C"
import (
	"encoding/base64"
	"fmt"
	"os"
	"strconv"
	"strings"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Printf("usage: %v <device> [directory or file to list]\n", os.Args[0])
		os.Exit(1)
	}

	udfread := C.udfread_init()
	if udfread == nil {
		fmt.Printf("couldn't initalize udfread\n")
		os.Exit(2)
	}

	ret := C.udfread_open(udfread, C.CString(os.Args[1]))
	if ret != 0 {
		fmt.Printf("failed to open %v\n", os.Args[1])
		os.Exit(3)
	}
	defer C.udfread_close(udfread)

	id := getVolumeId(udfread)
	if id == nil {
		fmt.Printf("failed to get Volume ID\n")
	} else {
		fmt.Printf("volume id = %v\n", *id)
	}

	id = getVolumesetId(udfread)
	if id == nil {
		fmt.Printf("failed to get Volumeset ID\n")
	} else {
		fmt.Printf("Base64 encoded Volumeset ID = %v\n", *id)
	}

	udfRoot := C.udfread_opendir(udfread, C.CString("/"))
	if udfRoot == nil {
		fmt.Printf("failed to get root directory entry\n")
		os.Exit(1)
	}

	fmt.Printf("%+v\n", udfRoot)
	dirent := &C.struct_udfread_dirent{}
	for {
		dirent := C.udfread_readdir(udfRoot, dirent)
		if dirent == nil {
			break
		}
		switch dirent.d_type {
		case C.UDF_DT_DIR:
			fmt.Printf("\tDir: %+v\n", C.GoString(dirent.d_name))
		case C.UDF_DT_REG:
			name := C.GoString(dirent.d_name)
			fmt.Printf("\tFile: %+v\n", name)
			if name == "disc.inf" {
				data := dumpFile(udfread, "disc.inf")
				if data != nil {
					playLists := getPlaylists(data)
					fmt.Printf("# of playlists listed = %v\n", len(playLists))
					for _, p := range playLists {
						fmt.Printf("\t%v\n", p)
					}
				}
			}
		default:
			fmt.Printf("\tUnknown: %+v\n", C.GoString(dirent.d_name))
		}
	}
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

func getVolumesetId(udfread *C.udfread) *string {
	data := C.malloc(128)

	retVal := C.udfread_get_volume_set_id(udfread, data, 128)
	if retVal == 0 {
		fmt.Printf("getVolumesetId failed\n")
		return nil
	}

	encoded := base64.StdEncoding.EncodeToString(C.GoBytes(data, 128))
	return &encoded
}

func getVolumeId(udfread *C.udfread) *string {
	data := C.udfread_get_volume_id(udfread)
	if data == nil {
		return nil
	}

	str := C.GoString((*C.char)(data))

	return &str
}