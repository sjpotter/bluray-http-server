package handlers

import "C"
import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/sjpotter/bluray-http-server/pkg/utils"
)

func init() {
	http.HandleFunc("/metadata", metadata)
}

func metadata(writer http.ResponseWriter, request *http.Request) {
	if request.URL.Query().Get("file") == "" {
		utils.GenericError(writer, errors.New("need to provide a file"))
		return
	}
	file := request.URL.Query().Get("file")

	if request.URL.Query().Get("playlist") == "" {
		utils.GenericError(writer, errors.New("need to provide a playlist"))
		return
	}

	playlist, err := strconv.Atoi(request.URL.Query().Get("playlist"))
	if err != nil {
		utils.GenericError(writer, err)
		return
	}

	bdrs, err := NewBDReadSeeker(file, playlist, 0)
	if err != nil {
		utils.GenericError(writer, err)
		return
	}
	defer bdrs.Close()

	bdTitle, err := bdrs.ParseTile()
	if err != nil {
		utils.GenericError(writer, err)
		return
	} else {
		data, err := json.MarshalIndent(bdTitle, "", "\t")
		if err != nil {
			utils.GenericError(writer, err)
			return
		} else {
			_, err1 := writer.Write(data)
			if err1 != nil {
				fmt.Printf("http writer failed: %v\n", err1)
			}
		}
	}
}
