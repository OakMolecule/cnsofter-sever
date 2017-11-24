package main

import (
	"cnsoftbei/mqtt"
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

type position struct {
	Time      string  // 定位时间
	Latitude  float64 // 纬度
	Longitude float64 // 经度
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

	remindJSON, err := json.Marshal(reminds)
	checkErr(err)
	mqtt.RemindEatMedicine(string(remindJSON))
	remindJSONIndent, err := json.MarshalIndent(reminds, "", "     ")
	checkErr(err)
	fmt.Println(remindJSON)
	fmt.Println(string(remindJSONIndent))
}

func checkMedicine() {
	stmt, err := db.Prepare("DELETE FROM remind WHERE e_time<DATE(NOW())")
	checkErr(err)
	_, err = stmt.Exec()
	checkErr(err)
}

// 更新数据库中的提醒信息
func updateMedicine(w http.ResponseWriter, r *http.Request) {
	var med []medicine
	r.ParseForm()
	fmt.Println(r.Host)
	b, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Println(err)
	}
	log.Println(string(b))
	m := string(b)
	w.WriteHeader(http.StatusOK)
	json.Unmarshal([]byte(m), &med)
	log.Println(med)
	log.Println(len(med))
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

//获取位置信息
func getPosition(w http.ResponseWriter, r *http.Request) {
	// r.ParseForm()
	log.Print(r.RemoteAddr)
	b, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Println(err)
	}
	log.Println(string(b))

	var positionJSON position
	json.Unmarshal(b, &positionJSON)
	stmt, err := db.Prepare("INSERT INTO position (longitude, latitude, time) values(?, ?, ?)")
	checkErr(err)

	var res sql.Result
	res, err = stmt.Exec(positionJSON.Longitude, positionJSON.Latitude, positionJSON.Time)
	checkErr(err)

	var id int64
	id, err = res.LastInsertId()
	checkErr(err)
	log.Println(id)
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
	//specRemind := "0 */1 * * * *" 	//每分钟执行一次

	c.AddFunc(specCheckMedicine, checkMedicine)
	c.AddFunc(specRemind, remindTime)
	c.Start()

	//开始http监听
	http.HandleFunc("/", updateMedicine)
	http.HandleFunc("/updateposition", getPosition)
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
