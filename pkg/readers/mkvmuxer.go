package readers

import (
	"bytes"
	"fmt"
	"io"
	"os/exec"
	"strconv"
	"sync"
	"time"

	ffprobe "github.com/sjpotter/go-ffprobe"
)

type MKVMuxer struct {
	b      *BDReadSeeker
	cmd    *exec.Cmd
	stdout io.ReadCloser
	stderr io.ReadCloser
}

func NewMKVMuxer(b *BDReadSeeker) (*MKVMuxer, error) {
	info, err := b.ParseTile()
	if err != nil {
		return nil, err
	}

	probeData, err := ffprobe.GetProbeData(b, 10*time.Second)
	if err != nil {
		return nil, err
	}
	b.Seek(0, io.SeekStart)

	cmdLine := []string{"-i", "-", "-map", "0"}
	audioCount := 0
	subtitleCount := 0

	fmt.Printf("audio = %+v\n", info.Audio)
	fmt.Printf("pgs = %+v\n", info.PG)

	for _, s := range probeData.Streams {
		idx, _ := strconv.Atoi(s.Id)

		if s.CodecType == "audio" {
			lang := info.Audio[idx].AudioLang
			cmdLine = append(cmdLine, fmt.Sprintf("-metadata:s:a:%v", audioCount), fmt.Sprintf("language=%v", lang))
			audioCount++
		} else if s.CodecType == "subtitle" {
			lang := info.PG[idx].PGLang
			cmdLine = append(cmdLine, fmt.Sprintf("-metadata:s:s:%v", subtitleCount), fmt.Sprintf("language=%v", lang))
			subtitleCount++
		}
	}

	cmdLine = append(cmdLine, "-codec", "copy", "-f", "matroska", "-")

	cmd := exec.Command("ffmpeg", cmdLine...)
	cmd.Stdin = b

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		return nil, err
	}

	fmt.Printf("cmd = %+v\n", cmd)

	return &MKVMuxer{
		b:      b,
		cmd:    cmd,
		stdout: stdout,
		stderr: stderr,
	}, nil
}

func (m *MKVMuxer) Read(p []byte) (n int, err error) {
	if m.cmd.Process == nil {
		if err := m.cmd.Start(); err != nil {
			return 0, err
		}
	}

	return m.stdout.Read(p)
}

func (m *MKVMuxer) Seek(offset int64, whence int) (int64, error) {
	return m.b.Seek(offset, whence)
}

func (m *MKVMuxer) Close() {

	buf := bytes.Buffer{}
	wg := sync.WaitGroup{}
	wg.Add(1)

	go func() {
		io.Copy(&buf, m.stderr)
		wg.Done()
	}()

	m.cmd.Process.Kill()
	m.cmd.Process.Wait()

	m.stdout.Close()
	m.stderr.Close()

	wg.Wait()
	fmt.Printf("cmd output = \n%v\n", string(buf.Bytes()))
}
