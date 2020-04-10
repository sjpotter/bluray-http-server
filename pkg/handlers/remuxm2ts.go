package handlers

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httputil"
	"strconv"
	"time"

	"github.com/sjpotter/bluray-http-server/pkg/readers"
	"github.com/sjpotter/bluray-http-server/pkg/utils"
)

func init() {
	http.HandleFunc("/remuxm2ts", remuxm2ts)
}

func remuxm2ts(writer http.ResponseWriter, request *http.Request) {
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

	requestDump, err := httputil.DumpRequest(request, false)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(string(requestDump))

	bdrs, err := readers.NewBDReadSeeker(file, playlist, seekTime)
	if err != nil {
		utils.GenericError(writer, err)
		return
	}
	defer bdrs.Close()

	remuxer, err := readers.NewM2TSRemuxer(bdrs)
	if err != nil {
		utils.GenericError(writer, err)
		return
	}
	defer remuxer.Close()

	writer.Header().Add("Content-Type", "application/octet-stream")
	writer.Header().Add("Content-Disposition", fmt.Sprintf("attachment; filename=\"%v.m2ts\"", playlistString))
	http.ServeContent(writer, request, fmt.Sprintf("%v.m2ts", playlistString), time.Now(), remuxer)
}
