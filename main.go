package main

import (
	//	_ "cnsoftbei/mqtt"
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	_ "github.com/go-sql-driver/mysql"
	"github.com/robfig/cron"
)

var httpport = ":9999"

type medicine struct {
	Amount       uint
	Etime        string
	ID           uint
	MedicineName string
	Stime        string
	Times        uint
	BaseObjID    int16
}

var db *sql.DB

func remind() {
}

func checkMedicine() {
	_, err := db.Prepare("DELETE * FROM remind WHERE e_time<DATE(NOW())")
	checkErr(err)
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

		res, err = stmt.Exec(med[i].MedicineName, med[i].Stime, med[i].Etime, med[i].Times, med[i].Amount)
		checkErr(err)

		id, err := res.LastInsertId()
		fmt.Println(id)
		checkErr(err)
	}
}

func main() {
	var err error

	c := cron.New()
	specCheckMedicine := "0 0 0 1/1 * * *" //设定每天早上检查药品是否过期
	specRemind := "0 0 7,12,19 * * *"      //设定每天提醒时间

	c.AddFunc(specCheckMedicine, checkMedicine)
	c.AddFunc(specRemind, remind)
	c.Start()

	//链接数据库
	db, err = sql.Open("mysql", "root:root@tcp(localhost:3306)/cnsoftbei?charset=utf8")
	checkErr(err)

	//开始http监听
	http.HandleFunc("/", sayhelloName)
	err = http.ListenAndServe(httpport, nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}

func checkErr(err error) {
	if err != nil {
		panic(err)
	}
}
