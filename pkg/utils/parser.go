package utils

/*
#cgo pkg-config: libbluray

#include <libbluray/bluray.h>
*/
import "C"

import (
	"fmt"
	"unsafe"

	"github.com/sjpotter/bluray-http-server/pkg/types"
)

func ParseTitle(p unsafe.Pointer) (*types.BDTitle, error) {
	ti := (*C.struct_bd_title_info)(p)

	videoType, err := parseVideo(ti)
	if err != nil {
		return nil, err
	}

	audioInfo, err := parseAudio(ti)
	if err != nil {
		return nil, err
	}
	pgInfo, err := parsePG(ti)
	if err != nil {
		return nil, err
	}
	chapterInfo, err := parseChapters(ti)
	if err != nil {
		return nil, err
	}

	bdTitle := &types.BDTitle{
		Playlist:  int(ti.playlist),
		Duration:  int64(ti.duration) / 90000,
		VideoType: videoType,
		Audio:     audioInfo,
		PG:        pgInfo,
		Chapters:  chapterInfo,
	}

	return bdTitle, nil
}

func parseVideo(ti *C.struct_bd_title_info) (string, error) {
	var videoType string
	switch int(ti.clips.video_streams.coding_type) {
	case 0xea:
		videoType = "vc1"
	case 0x1b:
		videoType = "h264"
	case 0x24:
		videoType = "hevc"
	default:
		fmt.Printf("unknown videotype %#x", int(ti.clips.video_streams.coding_type))
		videoType = "unknown"
	}

	return videoType, nil
}

func parseChapters(ti *C.struct_bd_title_info) ([]types.ChapterInfo, error) {
	var chapterInfos []types.ChapterInfo
	size := int(ti.chapter_count)
	chapters := (*[1 << 30]C.BLURAY_TITLE_CHAPTER)(unsafe.Pointer(ti.chapters))[:size:size]

	for i := 0; i < size; i++ {
		chapterInfos = append(chapterInfos, types.ChapterInfo{
			StartTime: uint64(chapters[i].start) / 9000,
			StartByte: uint64(chapters[i].offset),
		})
	}

	return chapterInfos, nil
}

func parsePG(ti *C.struct_bd_title_info) (map[int]*types.PGInfo, error) {
	pgInfos := make(map[int]*types.PGInfo)

	size := int(ti.clips.pg_stream_count)
	if ti.clips.pg_streams != nil {
		pgStreams := (*[1 << 30]C.BLURAY_STREAM_INFO)(unsafe.Pointer(ti.clips.pg_streams))[:size:size]
		for i := 0; i < size; i++ {
			pgInfo, err := parsePGStream(&pgStreams[i])
			if err != nil {
				return nil, err
			}
			pgInfos[int(pgStreams[i].pid)] = pgInfo
		}
	}

	return pgInfos, nil
}

func parseAudio(ti *C.struct_bd_title_info) (map[int]*types.AudioInfo, error) {
	audioInfos := make(map[int]*types.AudioInfo)

	size := int(ti.clips.audio_stream_count)
	if ti.clips.audio_streams != nil {
		audioStreams := (*[1 << 30]C.BLURAY_STREAM_INFO)(unsafe.Pointer(ti.clips.audio_streams))[:size:size]
		for i := 0; i < size; i++ {
			audioInfo, err := parseAudioStream(&audioStreams[i])
			if err != nil {
				return nil, err
			}
			audioInfos[int(audioStreams[i].pid)] = audioInfo
		}
	}

	return audioInfos, nil
}

func parseLang(clang [4]C.uchar) string {
	var blang []byte
	for i := 0; i < 4; i++ {
		if clang[i] == 0 {
			break
		}
		blang = append(blang, byte(clang[i]))
	}

	return string(blang)
}

func parsePGStream(stream *C.BLURAY_STREAM_INFO) (*types.PGInfo, error) {
	lang := parseLang(stream.lang)

	var pgType string
	switch int(stream.coding_type) {
	case 0x90:
		pgType = "pgs"
	case 0x91:
		pgType = "ig"
	case 0x92:
		pgType = "text"
	default:
		fmt.Printf("unknown pg type %#x", int(stream.coding_type))
		pgType = "unknown"
	}

	return &types.PGInfo{PGType: pgType, PGLang: lang}, nil
}

func parseAudioStream(stream *C.BLURAY_STREAM_INFO) (*types.AudioInfo, error) {
	lang := parseLang(stream.lang)

	var audioType string
	switch int(stream.coding_type) {
	case 0x80:
		audioType = "lpcm"
	case 0x81:
		audioType = "ac3"
	case 0x82:
		audioType = "dts"
	case 0x83:
		audioType = "truhd"
	case 0x84:
		audioType = "eac3"
	case 0x85:
		audioType = "dtshd"
	case 0x86:
		audioType = "dtshdma"
	default:
		fmt.Printf("unknown audio type %#x", int(stream.coding_type))
		audioType = "unknown"
	}

	return &types.AudioInfo{AudioType: audioType, AudioLang: lang}, nil
}
