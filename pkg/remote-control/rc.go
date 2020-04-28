package remote_control

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"k8s.io/klog"
	"net/http"
)

func init() {
	http.HandleFunc("/rc", rc)
}

func rc(w http.ResponseWriter, r *http.Request) {
	var jsonReq map[string]string

	var body []byte
	if r.Body != nil {
		if data, err := ioutil.ReadAll(r.Body); err == nil {
			body = data
		}
	}

	json.Unmarshal(body, &jsonReq)

	for k, v := range jsonReq {
		switch k {
		case "verbosity":
			var level klog.Level
			err := level.Set(v)
			if err != nil {
				klog.Infof("Failed to set verbosity to %v", err)
				w.Write([]byte(fmt.Sprintf("Failed: %v", err)))
			} else {
				klog.Infof("Set verbosity to %v", v)
				w.Write([]byte("OK"))
			}
			return
		default:
			klog.Infof()
			w.Write([]byte(fmt.Sprintf("Unknown: %v", k)))
			return
		}
	}
}