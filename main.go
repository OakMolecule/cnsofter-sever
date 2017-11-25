package main

import (
	"cnsoftbei/mqtt"
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
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

type basicCity struct {
	Cid        string
	Location   string
	ParentCity string `json:"parent_city"`
	AdminArea  string `json:"admin_area"`
	Cnty       string
	Lat        string
	Lon        string
	Tz         string
}

type updateTime struct {
	Loc string
	Utc string
}

type nowWeather struct {
	Cloud    string
	CondCode string `json:"cond_code"`
	CondTxt  string `json:"cond_txt"`
	Fl       string
	Hum      string
	Pcpn     string
	Pres     string
	Tmp      string
	Vis      string
	WindDeg  string `json:"wind_deg"`
	WindDir  string `json:"wind_dir"`
	WindSc   string `json:"wind_sc"`
	WindSpd  string `json:"wind_spd"`
}

type weather struct {
	Basic  basicCity
	Update updateTime
	Status string
	Now    nowWeather
}

type weather6 struct {
	HeWeather6 []weather
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

// 获取新位置信息
func updatePosition(w http.ResponseWriter, r *http.Request) {
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

// 返回位置信息
func getPosition(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()

	var sql string
	if len(r.Form) != 0 {
		start := r.FormValue("starttime")
		end := r.FormValue("endtime")

		fmt.Println(start)
		fmt.Println(end)
		sql = fmt.Sprintf("SELECT latitude, longitude, time FROM position WHERE time > '%s' AND time < '%s' ORDER BY time ASC", start, end)
		log.Println(sql)
	} else {
		sql = fmt.Sprint("SELECT latitude, longitude, time FROM position ORDER BY time ASC")
	}

	var positions []position
	rows, err := db.Query(sql)
	checkErr(err)

	for rows.Next() {
		var latitude float64
		var longitude float64
		var time string

		var newpositionJSON position

		err = rows.Scan(&latitude, &longitude, &time)
		if err != nil {
			panic(err.Error()) // proper error handling instead of panic in your app
		}

		newpositionJSON.Latitude = latitude
		newpositionJSON.Longitude = longitude
		newpositionJSON.Time = time
		positions = append(positions, newpositionJSON)
	}

	positionJSON, err := json.Marshal(positions)
	checkErr(err)

	w.Header().Set("Content-Type", "json; charset=utf-8")
	fmt.Fprintf(w, string(positionJSON))

	log.Println(positions)
}

// 返回位置信息
func getPositionnow(w http.ResponseWriter, r *http.Request) {

	sql := fmt.Sprint("SELECT latitude, longitude, time FROM position ORDER BY time ASC LIMIT 1")

	var positions position
	rows, err := db.Query(sql)
	checkErr(err)

	for rows.Next() {
		var latitude float64
		var longitude float64
		var time string

		var newpositionJSON position

		err = rows.Scan(&latitude, &longitude, &time)
		if err != nil {
			panic(err.Error()) // proper error handling instead of panic in your app
		}

		newpositionJSON.Latitude = latitude
		newpositionJSON.Longitude = longitude
		newpositionJSON.Time = time
		positions = newpositionJSON
		break
	}

	positionJSON, err := json.Marshal(positions)
	checkErr(err)

	w.Header().Set("Content-Type", "json; charset=utf-8")
	fmt.Fprintf(w, string(positionJSON))

	log.Println(positions)
}

func getWeather(w http.ResponseWriter, r *http.Request) {
	form := url.Values{}
	form.Add("location", "xiqing")
	form.Add("lang", "cn")
	form.Add("key", "4b94ed0b862d4f4689cf94c2d4fe507f")

	resp, err := http.PostForm("https://free-api.heweather.com/s6/weather/now", form)

	log.Println(form.Encode())

	if err != nil {
		// handle error
		log.Fatalf("post unmarshaling failed: %s", err)
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("ioutil.ReadAll failed: %s", err)
		// handle error
	}

	var weathernow weather6
	json.Unmarshal(body, &weathernow)

	fmt.Fprintf(w, weathernow.HeWeather6[0].Basic.Cid+",")
	fmt.Fprintf(w, weathernow.HeWeather6[0].Basic.Location+",")
	fmt.Fprintf(w, weathernow.HeWeather6[0].Now.Tmp+",")
	fmt.Fprintf(w, weathernow.HeWeather6[0].Now.WindDir+",")
	fmt.Fprintf(w, weathernow.HeWeather6[0].Now.WindSc)

	log.Println(weathernow.HeWeather6[0].Status)
	log.Println(weathernow.HeWeather6[0].Basic.Location)
	log.Println(weathernow.HeWeather6[0].Basic.Cid)
	log.Println(weathernow.HeWeather6[0].Now.CondTxt)
	log.Println(weathernow.HeWeather6[0].Now.Fl)
	log.Println(weathernow.HeWeather6[0].Now.WindDir)
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
	http.HandleFunc("/updateposition", updatePosition)
	http.HandleFunc("/getposition", getPosition)
	http.HandleFunc("/getpositionnow", getPositionnow)
	http.HandleFunc("/getweather", getWeather)
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
