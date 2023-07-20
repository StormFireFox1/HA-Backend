package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"sync"
)

var fileMutex sync.Mutex

var lineMetadata = map[int]string{
	0: "turGeneral",
	1: "turTavane",
	2: "returTavane",
	3: "turSemineu",
	4: "turPardoseala",
}

func returnTemps(w http.ResponseWriter, _ *http.Request) {
	// Open file that contains info. Assume file over NFS share.
	fileMutex.Lock()
	infoBytes, err := ioutil.ReadFile("/ha/info.txt")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Could not open file for reading information: " + err.Error()))
		return
	}
	fileMutex.Unlock()

	// Remove last newline
	infoContent := strings.TrimSuffix(string(infoBytes), "\n")
	infoLines := strings.Split(infoContent, "\n")
	info := make(map[string]float64, 0)

	for i, line := range infoLines {
		if number, err := strconv.Atoi(line); err == nil {
			info[lineMetadata[i]] = float64(number) / 100.0
		} else {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("Line " + fmt.Sprint(i+1) + " seems to not contain a number."))
			return
		}
	}

	result, err := json.MarshalIndent(info, "", "  ")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Could not convert information to JSON!"))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(result)
}

func main() {
	fileMutex = sync.Mutex{}

	http.HandleFunc("/temps", returnTemps)

	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		panic(err)
	}
}
