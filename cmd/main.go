package main

import "PunkIPv6"
import "fmt"

func main() {
	//PunkIPv6.RockFile(os.Args[1])
	PunkIPv6.RockMySQL()
	var i = 0;
	for i = 0; i < 5; i++ {
		PunkIPv6.RetryMysql()
		fmt.Print("\n")
	}
}
