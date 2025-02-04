package util

import (
	"fmt"
	"testing"
)

func TestName(t *testing.T) {
	ips, _ := ParseIP("172.20.42.5/24", "")
	fmt.Println(ips)
	for _, ip := range ips {
		fmt.Println(ip)
	}
}

func TestIsUpper(t *testing.T) {
	r := IsUpper("ALL")
	fmt.Println(r)
}

//func TestCheckAlive(t *testing.T) {
//	ips, _ := ParseIP("172.20.40.22/24", "")
//	plugins := []string{"ssh"}
//	//r := CheckAlive(ips, plugins, "")
//	fmt.Println(r)
//}

func TestStrXor(t *testing.T) {
	r := StrXor("", "1")
	fmt.Println(r)
	print(r)

}

func TestReadipfile(t *testing.T) {
	r, _ := Readipfile("/tmp/ip.txt")
	fmt.Println(r)
}
