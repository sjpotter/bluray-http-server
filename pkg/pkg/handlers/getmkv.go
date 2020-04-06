package handlers

import (
	"errors"
	"fmt"
	"github.com/sjpotter/bluray-http-server/pkg/utils"
	"io"
	"net/http"
	"os"
	"os/exec"
	"strconv"
)

func init() {
	http.HandleFunc("/getmkv", getmkv)
}

func getmkv(writer http.ResponseWriter, request *http.Request) {
	if request.URL.Query().Get("file") == "" {
		utils.GenericError(writer, errors.New("need to provide a file"))
		return
	}
	file := request.URL.Query().Get("file")

	if request.URL.Query().Get("playlist") == "" {
		utils.GenericError(writer, errors.New("need to provide a playlist"))
		return
	}

	playlistString := request.URL.Query().Get("playlist")
	playlist, err := strconv.Atoi(playlistString)
	if err != nil {
		utils.GenericError(writer, err)
		return
	}

	var seekTime int
	timeString := request.URL.Query().Get("time")
	if timeString != "" {
		seekTime, err = strconv.Atoi(timeString)
		if err != nil {
			utils.GenericError(writer, err)
			return
		}
	}

	bdrs, err := NewBDReadSeeker(file, playlist, seekTime)
	if err != nil {
		utils.GenericError(writer, err)
		return
	}
	defer bdrs.Close()

	info, err := bdrs.ParseTile()
	if err != nil {
		utils.GenericError(writer, err)
		return
	}

	writer.Header().Add("Content-Type", "application/octet-stream")
	writer.Header().Add("Content-Disposition", fmt.Sprintf("attachment; filename=\"%v.mkv\"", playlistString))
	writer.Header().Add("Content-Length", "-1")

	cmdLine := []string{"-i", "-", "-map", "0"}
	for i, a := range info.Audio {
		cmdLine = append(cmdLine, fmt.Sprintf("-metadata:s:a:%v", i), fmt.Sprintf("language=%v", a.AudioLang))
	}
	for i, s := range info.PG {
		cmdLine = append(cmdLine, fmt.Sprintf("-metadata:s:s:%v", i), fmt.Sprintf("language=%v", s.PGLang))
	}

	cmdLine = append(cmdLine, "-codec", "copy", "-f", "matroska", "-")

	cmd := exec.Command("ffmpeg", cmdLine...)
	cmd.Stdin = bdrs
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		fmt.Println(err)
		os.Exit(4)
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		fmt.Println(err)
		os.Exit(5)
	}

	fmt.Printf("cmd = %+v\n", cmd)

	if err := cmd.Start(); err != nil {
		fmt.Println(err)
		os.Exit(6)
	}

	io.Copy(writer, stdout)

	if err := cmd.Wait(); err != nil {
		fmt.Println(err)
	}

	buf := make([]byte, 100000)
	len, err := stderr.Read(buf)
	if err != nil {
		fmt.Println(err)
	}

	fmt.Printf("stderr = \n%v\n", string(buf[0:len]))
}

