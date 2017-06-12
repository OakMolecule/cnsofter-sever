package main

import (
	//	_ "cnsoftbei/mqtt"
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/robfig/cron"
)

var httpport = ":9999"

type medicine struct {
	Amount       uint   //每次服用的片数
	Etime        string //结束服用时间
	ID           uint   //ID号
	MedicineName string //药品名
	Stime        string //开始服用时间
	Times        uint   //次数
	BaseObjID    int16
}

type remind struct {
	MedicineName string //药品名称
	Times        uint   //每天服用次数
	Amount       uint   //每次服用片数
}

var db *sql.DB

func remindTime() {
	fmt.Println(time.Now())
	var reminds []remind
	rows, err := db.Query("SELECT medicine_name,times,amount FROM remind WHERE s_time<=DATE(NOW())")
	checkErr(err)

	// Fetch rows
	for rows.Next() {
		var amount uint
		var medicineName string
		var times uint

		var newRemind remind

		err = rows.Scan(&medicineName, &times, &amount)
		if err != nil {
			panic(err.Error()) // proper error handling instead of panic in your app
		}

		newRemind.Amount = amount
		newRemind.MedicineName = medicineName
		newRemind.Times = times
		reminds = append(reminds, newRemind)

	}

	log.Println(reminds)

	remindJson, err := json.Marshal(reminds)
	checkErr(err)
	remindJsonIndent, err := json.MarshalIndent(reminds, "", "     ")
	checkErr(err)
	fmt.Println(remindJson)
	fmt.Println(string(remindJsonIndent))
}

func checkMedicine() {
	stmt, err := db.Prepare("DELETE FROM remind WHERE e_time<DATE(NOW())")
	checkErr(err)
	_, err = stmt.Exec()
	checkErr(err)
}

func sayhelloName(w http.ResponseWriter, r *http.Request) {
	var med []medicine
	r.ParseForm()
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

	//链接数据库
	db, err = sql.Open("mysql", "root:root@tcp(localhost:3306)/cnsoftbei?charset=utf8")
	checkErr(err)
	defer db.Close()

	err = db.Ping()
	if err != nil {
		panic(err.Error()) // proper error handling instead of panic in your app
	}

	c := cron.New()
	specCheckMedicine := "0 0 0 * * *" //设定每天早上检查药品是否过期
	specRemind := "0 0 7,12,19 * * *"  //设定每天提醒时间
	//specRemind := "0 */1 * * * *" //设定每天提醒时间

	c.AddFunc(specCheckMedicine, checkMedicine)
	c.AddFunc(specRemind, remindTime)
	c.Start()

	//开始http监听
	http.HandleFunc("/", sayhelloName)
	err = http.ListenAndServe(httpport, nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}

func checkErr(err error) {
	if err != nil {
		log.Fatal("", err)
		panic(err)
	}
}
