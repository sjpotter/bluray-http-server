package m2ts_fs

import (
	"io/ioutil"
	"path/filepath"

    "github.com/goccy/go-yaml"

	"github.com/sjpotter/bluray-http-server/pkg/readers"
)

func readM2TSInfoFile(path string) (*M2TSInfo, error) {
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
	info, err := readM2TSInfoFile(path)
	if err != nil {
		return nil, err
	}

	parent := filepath.Dir(path)
	iso := filepath.Join(parent, info.File)

	return readers.NewBDReadSeeker(iso, info.Playlist)
}

func getM2TSRemuxer(path string, insertLang bool) (readers.BluRayReader, error) {
	info, err := readM2TSInfoFile(path)
	if err != nil {
		return nil, err
	}

	parent := filepath.Dir(path)
	iso := filepath.Join(parent, info.File)

	if insertLang {
		return readers.NewM2TSRemuxer(iso, info.Playlist)
	} else {
		return readers.NewBDReadSeeker(iso, info.Playlist)
	}
}