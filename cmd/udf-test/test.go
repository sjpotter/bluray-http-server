package main

/*
#cgo pkg-config: libudf
#include <stdlib.h>

#include <cdio/udf.h>
#include <cdio/udf_file.h>
 */
import "C"
import (
	"encoding/base64"
	"fmt"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Printf("usage: %v <device> [directory or file to list]\n", os.Args[0])
		os.Exit(1)
	}

	udf := C.udf_open(C.CString(os.Args[1]))
	if udf == nil {
		fmt.Printf("failed to open %v\n", os.Args[1])
		os.Exit(1)
	}
	defer C.udf_close(udf)

	id := getVolumeId(udf)
	if id == nil {
		fmt.Printf("failed to get Volume ID\n")
	} else {
		fmt.Printf("volume id = %v\n", *id)
	}

	id = getVolumesetId(udf)
	if id == nil {
		fmt.Printf("failed to get Volumeset ID\n")
	} else {
		fmt.Printf("Base64 encoded Volumeset ID = %v\n", *id)
	}

	//anyPartition := 0
	partNumber := C.udf_get_part_number(udf)
	if partNumber == -1 {
		fmt.Printf("faild to get partNumber\n")
		//anyPartition = 1
	}

	fmt.Printf("partNumber = %v\n", partNumber)

	udf_root := C.udf_get_root(udf, C.bool(1), C.ushort(partNumber))
	if udf_root == nil {
		fmt.Printf("failed to get root directory entry\n")
		os.Exit(1)
	}

	fmt.Printf("%+v\n", udf_root)
	for {
		dirent := C.udf_readdir(udf_root)
		if dirent == nil {
			break
		}
		fmt.Printf("\t%+v\n", dirent)
	}
}

func getVolumesetId(udf *C.udf_t) *string {
	data := C.malloc(128)

	retVal := C.udf_get_volumeset_id(udf, (*C.uchar)(data), 128)
	if retVal == 0 {
		fmt.Printf("getVolumesetId failed\n")
		return nil
	}

	encoded := base64.StdEncoding.EncodeToString(C.GoBytes(data, 128))
	return &encoded
}

func getVolumeId(udf *C.udf_t) *string {
	volIdSize := C.udf_get_volume_id(udf, nil, 0)
	if volIdSize <= 0 {
		fmt.Printf("volIdSize (%v) was <= 0\n", volIdSize)
		return nil
	}

	data := C.malloc(C.ulong(volIdSize))
	if data == nil {
		fmt.Printf("getVolumeId: malloc failed\n")
		return nil
	}
	defer C.free(data)

	C.udf_get_volume_id(udf, (*C.char)(data), C.uint(volIdSize))
	str := C.GoString((*C.char)(data))

	return &str
}