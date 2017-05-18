package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	_ "github.com/go-sql-driver/mysql"
)

type medicine struct {
	Amount        uint
	E_time        string
	Id            uint
	Medicine_name string
	S_time        string
	Times         uint
	BaseObjId     int16
}

var db *sql.DB

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
	w.WriteHeader(http.StatusOK)
	json.Unmarshal([]byte(m), &med)
	fmt.Println(med)
	fmt.Println(len(med))
	stmt, err := db.Prepare("DELETE FROM remind")
	res, err := stmt.Exec()
	checkErr(err)
	for i := 0; i < len(med); i++ {
		stmt, err = db.Prepare("INSERT remind SET medicine_name=?,s_time=?,e_time=?,times=?,amount=?")
		checkErr(err)

		res, err = stmt.Exec(med[i].Medicine_name, med[i].S_time, med[i].E_time, med[i].Times, med[i].Amount)
		checkErr(err)

		id, err := res.LastInsertId()
		fmt.Println(id)
		checkErr(err)
	}
}

func main() {
	var err error
	db, err = sql.Open("mysql", "root:root@tcp(localhost:3306)/cnsoftbei?charset=utf8")
	checkErr(err)
	http.HandleFunc("/", sayhelloName)
	err = http.ListenAndServe(":9090", nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}

func checkErr(err error) {
	if err != nil {
		panic(err)
	}
}
