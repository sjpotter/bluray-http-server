package m2ts_fs

import (
	"github.com/sjpotter/bluray-http-server/pkg/readers"
	"io/ioutil"
	"path/filepath"

    "github.com/goccy/go-yaml"
)

func getM2TS(path string) (*M2TSInfo, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var info M2TSInfo
	err = yaml.Unmarshal(data, &info)
	if err != nil {
		return nil, err
	}

	info.Name = filepath.Base(path)

	return &info, nil
}

func getBDRS(path string) (*readers.BDReadSeeker, error) {
	info, err := getM2TS(path)
	if err != nil {
		return nil, err
	}

	parent := filepath.Dir(path)
	iso := filepath.Join(parent, info.File)

	return readers.NewBDReadSeeker(iso, info.Playlist, 0)
}

func getM2TSRemuxer(path string) (*readers.M2TSRemuxer, error) {
	info, err := getM2TS(path)
	if err != nil {
		return nil, err
	}

	parent := filepath.Dir(path)
	iso := filepath.Join(parent, info.File)

	return readers.NewM2TSRemuxer(iso, info.Playlist)
}