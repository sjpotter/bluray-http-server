package types

type AudioInfo struct {
	AudioType string
	AudioLang string
}

type PGInfo struct {
	PGType string
	PGLang string
}

type ChapterInfo struct {
	StartTime uint64
	StartByte uint64
}

type BDTitle struct {
	Title     string
	Playlist  int
	Duration  int64
	VideoType string
	Audio     map[int]*AudioInfo
	PG        map[int]*PGInfo
	Chapters  []ChapterInfo
}
