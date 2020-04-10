package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/sjpotter/bluray-http-server/pkg/pmtparser"
)

func main() {
	flag.Parse()

	data, err := ioutil.ReadFile(os.Args[1])
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	pmt, _, err := pmtparser.ParsePMTPackets(data)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	pmtOut := pmt.Output()

	f, _ := os.Create("output")
	f.Write(pmtOut)
}
