package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
)

type medicine struct {
	Amount        int16
	E_time        string
	Id            uint
	Medicine_name string
	S_time        string
	Times         int16
	BaseObjId     int16
}

func sayhelloName(w http.ResponseWriter, r *http.Request) {
	var med []medicine
	r.ParseForm()
	fmt.Println(r.Form)
	fmt.Println(r.Host)
	b, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Println(err)
	}
	fmt.Println(string(b))
	m := string(b)
	json.Unmarshal([]byte(m), &med)
	fmt.Println(med)
	w.WriteHeader(http.StatusOK)
}

func main() {
	http.HandleFunc("/", sayhelloName)
	err := http.ListenAndServe(":9090", nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
