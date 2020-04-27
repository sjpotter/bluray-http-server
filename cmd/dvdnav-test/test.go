package main

/*
#cgo pkg-config: dvdnav dvdread

#include <stdlib.h>
#include <dvdnav/dvdnav.h>
*/
import "C"
import (
	"flag"
	"fmt"
	"io"
	"os"
	"strconv"
	"unsafe"
)

func main() {
	flag.Args()

	p := C.malloc(C.ulong(2048))
	if p != nil {
		defer C.free(p)
	} else {
		fmt.Printf("couldn't allocate memory space for the read")
		os.Exit(1)
	}

	handle := &C.struct_dvdnav_s{}
	if C.dvdnav_open(&handle, C.CString(os.Args[1])) != C.DVDNAV_STATUS_OK {
		fmt.Printf("Failed to open %v: %v\n", os.Args[1], C.GoString(C.dvdnav_err_to_string(handle)))
		os.Exit(1)
	}
	defer C.dvdnav_close(handle)

	var numTitles C.int

	if C.dvdnav_get_number_of_titles(handle, &numTitles) != C.DVDNAV_STATUS_OK {
		fmt.Printf("Failed to get number of titles %v: %v\n", os.Args[1], C.GoString(C.dvdnav_err_to_string(handle)))
		os.Exit(2)
	}

	fmt.Printf("Number of Titles = %v\n", numTitles)

	for i := 1; i < int(numTitles); i++ {
		fmt.Printf("Title: %v\n", i)

		var parts C.int
		if C.dvdnav_get_number_of_parts(handle, C.int(i), &parts) != C.DVDNAV_STATUS_OK {
			fmt.Printf("Failed to get number of parts for title %v:%v: %v\n", os.Args[1], i, C.GoString(C.dvdnav_err_to_string(handle)))
			continue
		}
		fmt.Printf("\tparts: %v\n", parts)
		var times *C.ulong
		var duration C.ulong

		num := C.dvdnav_describe_title_chapters(handle, C.int(i), &times, &duration)
		if num == 0 {
			fmt.Printf("Failed to get duration of title %v:%v: %v\n", os.Args[1], i, C.GoString(C.dvdnav_err_to_string(handle)))
			continue
		}

		fmt.Printf("\tChapters: %v\n", num)

		timeSlice := (*[1 << 30]C.ulong)(unsafe.Pointer(times))[:num:num]

		var divisor C.ulong = 90000

		for j := 0; j < int(num); j++ {
			fmt.Printf("\t\tChapter %v: %v\n", j, timeSlice[j]/divisor)
		}

		C.free(unsafe.Pointer(times))


		var pos C.uint
		var len C.uint
		if C.dvdnav_get_position_in_title(handle, &pos, &len) != C.DVDNAV_STATUS_OK {
			fmt.Printf("Failed to get length of title %v:%v: %v\n", os.Args[1], i, C.GoString(C.dvdnav_err_to_string(handle)))
		}

		fmt.Printf("title: %v -> parts = %v, duration = %v seconds, block = %v, 	pos = %v\n", i, parts, duration / 90000, len, pos)

		/*
		if C.dvdnav_sector_search(handle, C.long(len/2), io.SeekStart) != C.DVDNAV_STATUS_OK {
			fmt.Printf("failed to seek: %v:%v: %v\n", os.Args[1], i, C.GoString(C.dvdnav_err_to_string(handle)))
		}

		var event C.int
		var bsize C.int

		if C.dvdnav_get_next_block(handle, (*C.uchar)(p), &event, &bsize) != C.DVDNAV_STATUS_OK {
			fmt.Printf("Failed to read block %v:%v: %v\n", os.Args[1], i, C.GoString(C.dvdnav_err_to_string(handle)))
			os.Exit(1)
		}

		if C.dvdnav_get_position_in_title(handle, &pos, &len) != C.DVDNAV_STATUS_OK {
			fmt.Printf("Failed to get length of title %v:%v: %v\n", os.Args[1], i, C.GoString(C.dvdnav_err_to_string(handle)))
		}

		fmt.Printf("title: %v -> parts = %v, duration = %v seconds, block = %v, cur_pos = %v\n", i, parts, duration / 90000, len, pos)
		 */
	}

	val, err := strconv.Atoi(os.Args[2])
	if err != nil {
		fmt.Printf("%v\n", err)
		os.Exit(1)
	}

	if C.dvdnav_title_play(handle, C.int(val)) != C.DVDNAV_STATUS_OK {
		fmt.Printf("Failed to start playing %v:%v: %v\n", os.Args[1], val, C.GoString(C.dvdnav_err_to_string(handle)))
		os.Exit(1)
	}

	file, err := os.Create("output.mpeg")
	if err != nil {
		fmt.Printf("failed to create output file: %v\n", err)
		os.Exit(1)
	}
	defer file.Close()

	var pos C.uint
	var blocks C.uint
	if C.dvdnav_get_position_in_title(handle, &pos, &blocks) != C.DVDNAV_STATUS_OK {
		fmt.Printf("Failed to get length of title %v:%v: %v\n", os.Args[1], val, C.GoString(C.dvdnav_err_to_string(handle)))
	}

	fmt.Printf("current pos = %v\n", pos)

	if C.dvdnav_sector_search(handle, C.long((blocks*2048)/2), io.SeekStart) != C.DVDNAV_STATUS_OK {
		fmt.Printf("failed to seek: %v:%v: %v\n", os.Args[1], C.int(val), C.GoString(C.dvdnav_err_to_string(handle)))
	}

	if C.dvdnav_get_position_in_title(handle, &pos, &blocks) != C.DVDNAV_STATUS_OK {
		fmt.Printf("Failed to get length of title %v:%v: %v\n", os.Args[1], val, C.GoString(C.dvdnav_err_to_string(handle)))
	}

	fmt.Printf("new pos = %v\n", pos)

	var event C.int
	var len   C.int

	total := 0
	count := 0

loop:
	for {
		if C.dvdnav_get_position_in_title(handle, &pos, &blocks) != C.DVDNAV_STATUS_OK {
			fmt.Printf("Failed to get length of title %v:%v: %v\n", os.Args[1], val, C.GoString(C.dvdnav_err_to_string(handle)))
		}

		if count % 10000 == 0 {
			fmt.Printf("Current pos = %v\n", pos)
		}

		count++

		if C.dvdnav_get_next_block(handle, (*C.uchar)(p), &event, &len) != C.DVDNAV_STATUS_OK {
			fmt.Printf("Failed to read block %v:%v: %v\n", os.Args[1], val, C.GoString(C.dvdnav_err_to_string(handle)))
			os.Exit(1)
		}
		switch event {
		case C.DVDNAV_BLOCK_OK:
			if total == 0 {
				var seek C.long = 0
				ok := C.DVDNAV_STATUS_OK + 1

				for ok != C.DVDNAV_STATUS_OK {
					if C.dvdnav_sector_search(handle, seek, io.SeekEnd) != C.DVDNAV_STATUS_OK {
						fmt.Printf("failed to seek (%v): %v:%v: %v\n", seek, os.Args[1], C.int(val), C.GoString(C.dvdnav_err_to_string(handle)))
					} else {
						ok = C.DVDNAV_STATUS_OK
					}
					seek++
				}

				if C.dvdnav_get_position_in_title(handle, &pos, &blocks) != C.DVDNAV_STATUS_OK {
					fmt.Printf("Failed to get length of title %v:%v: %v\n", os.Args[1], val, C.GoString(C.dvdnav_err_to_string(handle)))
				}

				fmt.Printf("Current pos = %v\n", pos)
			}
			total += int(len)
/*			data := C.GoBytes(p, len)
			file.Write(data)
 */
		case C.DVDNAV_STOP:
			fmt.Printf("Break on STOP!\n")
			break loop
		case C.DVDNAV_WAIT:
			fmt.Printf("Break on WAIT!\n")
			break loop
		default:
			fmt.Printf("Got event: %v @ %v\n", event, total)

			if C.dvdnav_get_position_in_title(handle, &pos, &blocks) != C.DVDNAV_STATUS_OK {
				fmt.Printf("Failed to get length of title %v:%v: %v\n", os.Args[1], val, C.GoString(C.dvdnav_err_to_string(handle)))
			}

			fmt.Printf("Current pos = %v\n", pos)
		}
	}

	if C.dvdnav_get_position_in_title(handle, &pos, &blocks) != C.DVDNAV_STATUS_OK {
		fmt.Printf("Failed to get length of title %v:%v: %v\n", os.Args[1], val, C.GoString(C.dvdnav_err_to_string(handle)))
	}

	fmt.Printf("total len = %v, current_pos = %v\n", total, pos)
}