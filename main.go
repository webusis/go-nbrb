package main
import (
    "encoding/json"
    "net/http"
    "io/ioutil"
    "fmt"
    "time"
    "os"
    "database/sql"
	"github.com/go-sql-driver/mysql"
	"strings"
)

var db *sql.DB

type Currencies struct {
   Cur_ID int `json:"Cur_ID"`
   Code	string `json:"Cur_Code"`
   Abbr	string `json:"Cur_Abbreviation"`
   Name string `json:"Cur_Name"`
   Name_eng string `json:"Cur_Name_Eng"`
   Sc_name string `json:"Cur_QuotName"`
   Sc_name_eng string `json:"Cur_QuotName_Eng"`
   Scale int `json:"Cur_Scale"`
   Sdate string `json:"Cur_DateStart"`
   Edate string `json:"Cur_DateEnd"`
}

type Currency struct {
	CId int	`json:"Cur_ID"`
	Date string	`json:"Date"`
	Abbr string	`json:"Cur_Abbreviation"`
	Scale int `json:"Cur_Scale"`
	Name string	`json:"Cur_Name"`
	Rate float32 `json:"Cur_OfficialRate"`
}

type Rate struct {
	CId int
	Date string
	Rate float32 
	Scale int
}

func db_connect() *sql.DB {
	cfg := mysql.Config{
        User:   "tema",
        Passwd: "tema",
        Net:    "tcp",
        Addr:   "127.0.0.1:3306",
        AllowNativePasswords: true,
        DBName: "nbrb",
    }
	db, err := sql.Open("mysql", cfg.FormatDSN())
	if err != nil {
        panic(err)
    }
    //defer db.Close()
	return db
}

func init() {
	db = db_connect()
	
	var jsn string
	if len(os.Args) != 0 {
		switch os.Args[1]  {
			case "currencies":
				jsn = cUpdate()
			case "today":
				if len(os.Args) > 1 {
					jsn = rToday(os.Args[2])
				}
			case "date":
				if len(os.Args) > 3 {
					jsn = rDate(os.Args[2],os.Args[3],os.Args[4])
				}
			case "tsync":
				jsn = rUpdate()
		}
		
		fmt.Println(jsn)
	}
}

func main() {
    http.HandleFunc("/", handler)
    http.ListenAndServe(":8888", nil)
}

func handler(output http.ResponseWriter, input *http.Request) {
	output.Header().Set("Content-Type", "application/json; charset=utf-8")
	var jsn string
	path := strings.Split(input.URL.Path[1:], "/")

    switch path[0] {
    	case "currencies":
    		jsn = cUpdate()
    	case "today":
    		jsn = rToday(path[1])
    	case "date":
    		jsn = rDate(path[1],path[2],path[3])
    	case "tsync":
    		jsn = rUpdate()
    }
	
	fmt.Println(jsn)
	fmt.Fprintf(output, jsn)
}

func reqNbrb(path string) string {
	resp, err := http.Get(path)
    defer resp.Body.Close()
    
    body, err := ioutil.ReadAll(resp.Body)
    
    if err != nil {
    	panic(err)
    }
    
    return string(body)
}

func rDate(cId string,sDate string,eDate string) string {
	var r Rate
	path := "https://api.nbrb.by/ExRates/Rates/Dynamics/"+string(cId)+"?startDate="+sDate+"&endDate="+eDate
	rates := reqNbrb(path)
	var data []Currency
	err := json.Unmarshal([]byte(rates), &data)
	

	if err != nil {
		panic(err)
	}
	
	for _, cur := range data {
		date := cur.Date[0:10]

		db.QueryRow("select cId, date, rate from rates where cId = ? and date = ?", cur.CId, date).Scan(&r.CId, &r.Date, &r.Rate)

		if r.CId == 0{
			_, err := db.Exec("insert into rates (`cId`, `date`, `rate`) values (?, ?, ?)", cur.CId, cur.Date, cur.Rate)
			if err != nil {
				panic(err)
			}
		}
	}
	
	dt, err := json.Marshal(data)
	if err != nil {
		panic(err)
	}

	return string(dt)
}

func rToday(cId string) string {
	var r Rate
	var c Currency
	tm := time.Now()
	t := fmt.Sprintf("%d-%02d-%02d",
        tm.Year(), tm.Month(), tm.Day())
	err := db.QueryRow("select cId, date, rate from rates where cId = ? and date = ?", cId, t).Scan(&r.CId, &r.Date, &r.Rate, &r.Scale)

	if err != nil {
		path := "https://api.nbrb.by/exrates/rates/" +string(cId)
		currency := reqNbrb(path)
		
		err = json.Unmarshal([]byte(currency), &c)

		if err != nil {
			panic(err)
		}
		
		_, err := db.Exec("insert into rates (`cId`, `date`, `rate`, `scale`) values (?, ?, ?, ?)", c.CId, c.Date, c.Rate, c.Scale)
	    if err != nil {
			panic(err)
		}else{
			dt, err := json.Marshal(c)
			if err != nil {
				panic(err)
			}
			return string(dt)
		}
	}else{
		dt, err := json.Marshal(r)
		if err != nil {
			panic(err)
		}
		return string(dt)
	}

}

func rUpdate() string {
	var r Rate
	path := "https://api.nbrb.by/exrates/rates?periodicity=0"
	rates := reqNbrb(path)
	var data []Currency
	err := json.Unmarshal([]byte(rates), &data)
	
	tm := time.Now()
	t := fmt.Sprintf("%d-%02d-%02d",
        tm.Year(), tm.Month(), tm.Day())

	if err != nil {
		panic(err)
	}
	
	for _, cur := range data {
		db.QueryRow("select cId, date, rate from rates where cId = ? and date = ?", cur.CId, t).Scan(&r.CId, &r.Date, &r.Rate, &r.Scale)

		if r.CId == 0{
			_, err := db.Exec("insert into rates (`cId`, `date`, `rate`, `scale`) values (?, ?, ?, ?)", cur.CId, cur.Date, cur.Rate, cur.Scale)
			if err != nil {
				panic(err)
			}
		}
	}
	
	dt, err := json.Marshal(data)
	if err != nil {
		panic(err)
	}

	return string(dt)
}

func cUpdate() string {
	path := "https://api.nbrb.by/exrates/currencies"
	currencies := reqNbrb(path)
	var data []Currencies
	err := json.Unmarshal([]byte(currencies), &data)

	if err != nil {
		panic(err)
	}
	
	db.Exec("drop table if exists currencies ")
	
	for _, cur := range data {
		_, err := db.Exec("insert into currencies (`code`, `name`, `name_eng`, `sc_name`, `sc_name_eng`, `abbr`, `scale`, `sdate`, `edate`) values (?, ?, ?, ?, ?, ?, ?, ?, ?)", cur.Code, cur.Name, cur.Name_eng, cur.Sc_name, cur.Sc_name_eng, cur.Abbr, cur.Scale, cur.Sdate, cur.Edate)
	    if err != nil {
			panic(err)
		}
	}
	
	dt, err := json.Marshal(data)
	if err != nil {
		panic(err)
	}

	return string(dt)
}
