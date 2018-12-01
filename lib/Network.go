package lib

import (
	"errors"
	"fmt"
	"github.com/commonscan/fasthttp"
	"io/ioutil"
	"net/http"
	//	"fmt"
	log "github.com/sirupsen/logrus"
	"net"
	"time"
)

func IsIPv4(ip net.IP) bool {
	return ip.To4() != nil
}
func ClientGetURLDeadline(client fasthttp.Client, req *fasthttp.Request, response *fasthttp.Response) {
	timeout := -time.Since(time.Now().Add(time.Second * 30))
	ch := make(chan bool, 1)
	go func() {
		_ = client.Do(req, response)
		ch <- true
	}()

	tc := time.NewTimer(timeout)
	select {
	case _ = <-ch:
		//fmt.Println(response.Body())
	case <-tc.C:
		log.Error("Timeout")
	}
}

// get IPv4 Address
func Addr(domain string) (ipv4 string, ipv6 string, err error) {
	ips, err := net.LookupIP(domain)
	if err != nil {
		return "", "", err
	}
	for _, i := range ips {
		if IsIPv4(i) {
			ipv4 = i.String()
		} else {
			ipv6 = i.String()
		}
	}
	return ipv4, ipv6, nil
}

// 请求一个URL 返回hash
func RespHash(rawUrl string, ipv6Only bool) string {
	resp, err := DefaultRequest(rawUrl, ipv6Only)
	if err != nil {
		fmt.Println(err)
		return ""
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println(err)
		return ""
	}
	if len(body) > 0 { // 明确请求成功了
		return Md5Sum(body)
	} else if resp.StatusCode != 0 { // 跳转类的URL
		return fmt.Sprintf("%d", resp.StatusCode)
	} else { // 请求失败
		return ""
	}
}

func httpGetWithTimeout(rawUrl string, IPv6only bool) (resp *http.Response, err error) {
	timeout := -time.Since(time.Now().Add(time.Second * 30))
	ch := make(chan bool, 1)
	go func() {
		client := http.Client{Transport: MyTransportWrapper{IPv6Only: IPv6only}}
		resp, err = client.Get(rawUrl)
		ch <- true
	}()

	tc := time.NewTimer(timeout)
	select {
	case _ = <-ch:
		return resp, err
	case <-tc.C:
		return nil, errors.New("timeout error")
	}
}
func DefaultRequest(rawUrl string, IPv6only bool) (*http.Response, error) {
	return httpGetWithTimeout(rawUrl, IPv6only)
}
