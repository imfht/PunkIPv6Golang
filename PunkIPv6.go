package PunkIPv6

import (
	"PunkIPv6/lib"
	"bufio"
	"database/sql"
//	"github.com/cheggaaa/pb"
	//	"github.com/valyala/fasthttp"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	log "github.com/sirupsen/logrus"
	"net/url"
	"os"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

var logger log.Logger
var totalCount int
var finisedCount int64
var DEBUG = true

type Record struct {
	domain        string `json:"domain"`
	unit_code     string `json:"unit_code"`
	up_unitcode_0 string `json:"up_unitcode_0"`
	up_unitcode_1 string `json:"up_unitcode_1"`
	up_unitcode_2 string `json:"up_unitcode_2"`
	up_unitcode_3 string `json:"up_unitcode_3"`
	result        Result `json:"result"`
}

type Result struct {
	ipv4      string `json:"ipv_4"`
	ipv6      string `json:"ipv_6"`
	html_ipv4 string `json:"html_ipv_4"`
	html_ipv6 string `json:"html_ipv_6"`
}

// send a http request via IPv4
func IPv4Fetcher(domain string) {
}

// send a http request via IPv6
func IPv6Fetcher(domain string) {

}

// save result to pipeline(such as nsq)
func Pipeline() {
	// pipeline
}
func init() {
	file, _ := os.OpenFile("logrus.log", os.O_CREATE|os.O_WRONLY, 0666)

	log.SetFormatter(&log.JSONFormatter{})
	log.SetLevel(log.DebugLevel)
	log.SetOutput(file)
	//logger.Formatter = &log.JSONFormatter{}
}
func Worker(conn chan Record, pipeline chan Record, wg *sync.WaitGroup) { // work is the really worker. DNS query && do request
	sTime := time.Now()
	defer wg.Done()
	defer atomic.AddInt64(&finisedCount, 1)
	defer log.Info("time spend" + fmt.Sprintf("%f", time.Now().Sub(sTime).Seconds()))
	var record = <-conn
	result := Result{}
	record.result = result
	parsedURL, err := url.Parse(record.domain)
	if err != nil {
		log.Error(err, record.domain)
	}
	host := parsedURL.Host
	if strings.Contains(parsedURL.Host, ":") {
		host = strings.Split(host, ":")[0]
	}
	ipv4, ipv6, err := lib.Addr(host)
	if err != nil {
		for i := 0; i < 1; i++ { //retry
			ipv4, ipv6, err = lib.Addr(host)
			if err == nil {
				break
			}
		}
	}
	if err != nil {
		log.Error("dns error" + err.Error())
		return
	}
	result.ipv6 = ipv6
	result.ipv4 = ipv4
	if len(ipv6) > 0 {
		result.html_ipv6 = lib.RespHash(record.domain, true)
	}
	result.html_ipv4 = lib.RespHash(record.domain, false)
	if len(result.html_ipv4) > 0 && DEBUG {
		fmt.Println(result.html_ipv4)
	}
	record.result = result
	pipeline <- record
	//record.logResult()
}

func (r Record) logResult() {
	log.WithFields(log.Fields{
		"ipv_4":     r.result.html_ipv4,
		"ipv_6":     r.result.html_ipv6,
		"domain":    r.domain,
		"html_ipv4": r.result.html_ipv4,
		"html_ipv6": r.result.html_ipv6,
	}).Info()
}

// FetchWorker will call worker until main process get into exit.
func FetchWorker(conn chan Record, pipeline chan Record, wg *sync.WaitGroup, ) {
	for ; ; {
		Worker(conn, pipeline, wg)
	}
}

// load file and rock it
func RockFile(path string) {
	file, err := os.Open(path)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()
	var wg = sync.WaitGroup{}
	scanner := bufio.NewScanner(file)
	var count = 0
	conn := make(chan Record, 1000)
	pipeline := make(chan Record, 1000)
	for i := 0; i < 128; i++ {
		go FetchWorker(conn, pipeline, &wg)
	}
	for scanner.Scan() {
		count += 1
		wg.Add(1)
		record := Record{}
		record.domain = scanner.Text()
		conn <- record
	}
	wg.Wait()
}
func panicIfError(err error) {
	if err != nil {
		fmt.Println("ops", err)
		log.Fatal(err)
	}
}
func WorkerManager(record []Record) {
	var wg = sync.WaitGroup{}
	conn := make(chan Record, 10000)
	pipelineChan := make(chan Record, 10000)
	wg.Add(len(record))
	totalCount = len(record)
	go showBar()
	for i := 0; i < 10; i++ { // 启动插入mysql
		go pipeline(pipelineChan, i)
	}
	for i := 0; i < 32; i++ { // 启动fetcher
		go FetchWorker(conn, pipelineChan, &wg)
	}
	for i := 0; i < len(record); i++ {
		conn <- record[i]
	}
	wg.Wait()
	time.Sleep(time.Second * 10)
}
func mysqlLoader(query string) []Record {
	db, err := sql.Open("mysql", "root:@tcp(202.120.7.205:3308)/web")
	defer db.Close()
	panicIfError(err)
	rows, err := db.Query(query)
	panicIfError(err)
	var allRecords []Record

	columns, err := rows.Columns()
	panicIfError(err)
	values := make([]sql.RawBytes, len(columns))
	scanArgs := make([]interface{}, len(values))
	for i := range values {
		scanArgs[i] = &values[i]
	}
	for rows.Next() {
		err = rows.Scan(scanArgs...)
		if err != nil {
			panic(err.Error()) // proper error handling instead of panic in your app
		}
		//var value string
		domain, unitCode, up_unitcode_0, up_unitcode_1, up_unitcode_2, up_unitcode_3 := string(values[0]), string(values[1]), string(values[2]), string(values[3]), string(values[4]), string(values[5])
		totalCount += 1

		record := Record{}
		record.unit_code = unitCode
		record.domain = domain
		record.up_unitcode_0 = up_unitcode_0
		record.up_unitcode_1 = up_unitcode_1
		record.up_unitcode_2 = up_unitcode_2
		record.up_unitcode_3 = up_unitcode_3
		allRecords = append(allRecords, record)
	}
	return allRecords
}
func cleanIPv4Empty() {
	db, err := sql.Open("mysql", "root:@tcp(202.120.7.205:3308)/web")
	defer db.Close()
	panicIfError(err)
	a, err := db.Query("DELETE from api_detectlog where (ipv4_dns!='' && ipv4_html_hash='')  or (ipv6_html_hash='' && ipv6_dns!='')")
	fmt.Println(a, err)
}
func cleanMysqlTableLog() {
	db, err := sql.Open("mysql", "root:@tcp(202.120.7.205:3308)/web")
	defer db.Close()
	panicIfError(err)
	a, err := db.Exec("TRUNCATE  api_detectlog")
	fmt.Println(a, err)
}

func RetryMysql() { // 对MySQL中请求失败的网站进行retry
	atomic.StoreInt64(&finisedCount, 0)
	allRecord := mysqlLoader("SELECT domain,unit_code FROM api_detectlog where ipv4_html_hash='' or (ipv6_html_hash='' && ipv6_dns!='')")
	cleanIPv4Empty()
	WorkerManager(allRecord)
}

func RetryIPv4Empty(limit int) {
	allRecord := mysqlLoader("SELECT domain,unit_code FROM api_detectlog where ipv4_html_hash='' limit " + fmt.Sprintf("%d", limit))
	WorkerManager(allRecord)
}

func RockMySQL() {
	atomic.StoreInt64(&finisedCount, 0)
	cleanMysqlTableLog()
	fmt.Println("cleaned")
	var allRecord = mysqlLoader("SELECT domain,unit_code,up_unitcode_0,up_unitcode_1,up_unitcode_2,up_unitcode_3 FROM api_domain ")
	WorkerManager(allRecord)
}

func pipeline(recordChan chan Record, pipelineID int) {
	db, err := sql.Open("mysql", "root:9d23ebd61179@tcp(202.120.7.205:3308)/web")
	panicIfError(err)
	defer db.Close()
	err = db.Ping()
	panicIfError(err)
	stmtIns, err := db.Prepare("INSERT INTO api_detectlog (domain,ipv4_dns,ipv6_dns,ipv4_html_hash,ipv6_html_hash,unit_code,up_unitcode_0,up_unitcode_1,up_unitcode_2,up_unitcode_3) VALUES(?, ?, ?, ?, ?, ?, ?, ?, ?, ?)") // ? = placeholder
	if err != nil {
		panic(err.Error()) // proper error handling instead of panic in your app
	}
	defer stmtIns.Close() // Close the statement when we leave main() / the program terminates
	count := 0
	for ; ; {
		record := <-recordChan
		_, err = stmtIns.Exec(record.domain, record.result.ipv4, record.result.ipv6, record.result.html_ipv4, record.result.html_ipv6, record.unit_code, record.up_unitcode_0, record.up_unitcode_1, record.up_unitcode_2, record.up_unitcode_3)
		if err != nil {
			log.Error(err)
		}
		count++
		if count%100 == 0 {
			log.Info(pipelineID, ":", count)
		}
	}
}

func showBar() {
	var oldTotalCount = totalCount
//	bar := pb.StartNew(totalCount)
	for ; ; {
		if oldTotalCount != totalCount {
			//bar = pb.StartNew(totalCount)
			oldTotalCount = totalCount
		}
//		bar.SetCurrent(finisedCount)
		if int(totalCount) == int(finisedCount) {
			break
		}
		time.Sleep(time.Second)
	}
	//bar.FinishPrint("The End!")
}
