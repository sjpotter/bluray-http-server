package m2ts_fs

import (
	"io/ioutil"
	"path/filepath"
	"sync"

	"github.com/goccy/go-yaml"

	"k8s.io/klog"

	"github.com/sjpotter/bluray-http-server/pkg/readers"
)

var (
	openLock sync.Mutex
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

func getM2TSRemuxer(path string, insertLang bool) (readers.BluRayReader, error) {
	// have had issues with crashes in libbluray, wondering if its not thread safe on opening, maybe serializing helps?
	openLock.Lock()
	defer openLock.Unlock()

	info, err := readM2TSInfoFile(path)
	if err != nil {
		return nil, err
	}

	parent := filepath.Dir(path)
	iso := filepath.Join(parent, info.File)

	klog.V(2).Infof("Opening %v:%v", iso, info.Playlist)

	if insertLang {
		return readers.NewM2TSRemuxer(iso, info.Playlist)
	} else {
		return readers.NewBDReadSeeker(iso, info.Playlist)
	}
}