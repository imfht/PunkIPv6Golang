package PunkIPv6

import (
	"testing"
)

func TestRockIt(t *testing.T) {
	RockFile("./urls.txt")
}
func TestRockMySQL(t *testing.T) {
	RockMySQL()
}
func TestIPv4Fetcher(t *testing.T) {
	//resp := lib.DefaultRequest("http://www.sjtu.edu.cn", true)
	//fmt.Println(resp)
}

func TestWorkerManager(t *testing.T) {
	sdu := Record{}
	sdu.domain = "http://www.baidu.com"
	sjtu := Record{}
	sjtu.domain = "http://www.sjtu.edu.cn"
	var records = []Record{sdu, sjtu}
	WorkerManager(records)
}

func TestRetryIPv4Empty(t *testing.T) {
	RetryIPv4Empty(100)
}
