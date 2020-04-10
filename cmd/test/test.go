package main

/*
#cgo pkg-config: libbluray
#include <stdlib.h>

#include <libbluray/bluray.h>
*/
import "C"
import (
	"fmt"
	"os"
	"strconv"
	"unsafe"
)

func main() {
	if len(os.Args) < 3 {
		fmt.Printf("Usage:\n   %s <media_path> <title_number>\n\n", os.Args[0])
		os.Exit(1)
	}

	bd := C.bd_open(C.CString(os.Args[1]), nil)
	if bd == nil {
		fmt.Printf("Error opening %s\n", os.Args[1])
		os.Exit(1)
	}
	defer C.bd_close(bd)

	title, err := strconv.Atoi(os.Args[2])
	if err != nil {
		fmt.Printf("%v is not an int\n", os.Args[2])
		os.Exit(1)
	}

	bdTitleInfo, err := getTitleInfo(bd, title)
	if err != nil {
		fmt.Printf("%v\n", err)
		os.Exit(1)
	}
	defer C.bd_free_title_info(bdTitleInfo)

	dumpBDTitle(bdTitleInfo)
}

func dumpBDTitle(info *C.struct_bd_title_info) {
	fmt.Printf("%+v\n", info)
	dumpChapters(info.chapters, info.chapter_count)
	if info.clip_count != 0 {
		clips := (*[1 << 30]C.struct_bd_clip)(unsafe.Pointer(info.clips))[:info.clip_count:info.clip_count]
		for _, clip := range clips {
			dumpClips(&clip)
		}

	}
	dumpMarks(info.marks, info.mark_count)
}

func dumpChapters(chapters *C.struct_bd_chapter, count C.uint) {
	fmt.Println("Chapters:")
	fmt.Printf("  %+v (count = %v)\n", chapters, count)
}

func dumpClips(clip *C.struct_bd_clip) {
	fmt.Println("Clips:")
	if clip != nil {
		fmt.Printf("  %+v\n", clip)
		if clip.video_streams != nil {
			fmt.Printf("    video: %+v\n", clip.video_streams)
		}
		if clip.audio_streams != nil {
			size := int(clip.audio_stream_count)
			audioStreams := (*[1 << 30]C.BLURAY_STREAM_INFO)(unsafe.Pointer(clip.audio_streams))[:size:size]
			for i := 0; i < size; i++ {
				fmt.Printf("    audio: %+v\n", audioStreams[i])
			}
		}
		if clip.pg_streams != nil {
			size := int(clip.pg_stream_count)
			pgStreams := (*[1 << 30]C.BLURAY_STREAM_INFO)(unsafe.Pointer(clip.pg_streams))[:size:size]
			for i := 0; i < size; i++ {
				fmt.Printf("    pg: %+v\n", pgStreams[i])
			}
		}
		if clip.ig_streams != nil {
			fmt.Printf("   ig: %+v\n", clip.ig_streams)
		}
	}
}

func dumpMarks(marks *C.struct_bd_mark, count C.uint) {
	fmt.Println("Marks:")
	if marks != nil {
		fmt.Printf("  %+v (count = %v)\n", marks, count)
	}
}

func getTitleInfo(bd *C.BLURAY, t int) (*C.struct_bd_title_info, error) {
	numTitles := C.bd_get_titles(bd, C.TITLES_RELEVANT, 0)

	fmt.Printf("numTitles = %v\n", numTitles)

	ti := C.bd_get_title_info(bd, C.uint(t), 0)

	if ti == nil {
		return nil, fmt.Errorf("Failed to get title_info for title %v", t)
	}

	return ti, nil
}
