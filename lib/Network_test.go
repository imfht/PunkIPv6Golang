package lib

import (
	"fmt"
	"testing"
)

func TestIPv4Addr(t *testing.T) {
	ipv4, ipv6, err := Addr("www.sjtu.edu.cn")
	fmt.Println(ipv4, ipv6, err)
	if err != nil {
		t.Fail()
	}
	if !(len(ipv4) > 0 && len(ipv6) > 0) {
		t.Failed()
	}
}
func TestDefaultRequest(t *testing.T) {
	resp, err := DefaultRequest("http://signal.seu.edu12313.cn", true) // should return false
	if err == nil {
		fmt.Println(resp.Header, err)
	} else {
		fmt.Println("error")
	}
}

func TestIPv6Addr(t *testing.T) {

}
func TestSlowDomain(t *testing.T) {
	//resp := DefaultRequest("http://palm.seu.edu.cn/", false, 0) // should return false
	//	resp := DefaultRequest("http://www.sjtu.edu.cn", false, 0) // should return false
	//var url = "http://sbcdagl.ruc.edu.cn/" 官方库不能正常处理的302 URL
	var url = "http://hqc.tit.edu.cn"
	resp, err := DefaultRequest(url, false) // should return false
	fmt.Println(resp.Header, err)
}

func TestClientGetURLDeadline(t *testing.T) {

}
func TestDefaultRequest2(t *testing.T) {
	resp, err := DefaultRequest("http://www.baidu.com", false)
	fmt.Println(resp.Header, err)
}
func TestRespHash(t *testing.T) {
	fmt.Println(RespHash("http://www.e12rrox1231321123.edu.cn", false))
	fmt.Println(RespHash("http://www.sdu.edu.cn", false))
	fmt.Println(RespHash("http://loginsuccessful.wxit.edu.cn", false))
}
