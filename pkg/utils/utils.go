package utils

import (
	"fmt"
	"net/http"
)

func GenericError(writer http.ResponseWriter, err error) {
	writer.WriteHeader(500)
	_, err1 := writer.Write([]byte(fmt.Sprintf("%v", err)))
	if err1 != nil {
		fmt.Printf("http writer failed: %v\n", err1)
	}
}
