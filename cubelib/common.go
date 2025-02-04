package cubelib

import (
	"bufio"
	"crypto/md5"
	"cube/log"
	"cube/model"
	"fmt"
	"io"
	"os"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"
)

func ValidIp(ip string) bool {
	addr := strings.Trim(ip, " ")
	regStr := `^(([1-9]|[1-9][0-9]|1[0-9]{2}|2[0-4][0-9]|25[0-5])\.)(([0-9]|[1-9][0-9]|1[0-9]{2}|2[0-4][0-9]|25[0-5])\.){2}([0-9]|[1-9][0-9]|1[0-9]{2}|2[0-4][0-9]|25[0-5])$`
	if match, _ := regexp.MatchString(regStr, addr); match {
		return true
	}
	return false
}

func Split(r rune) bool {
	return strings.ContainsRune("://:", r)
}

func ParseService(str string) (service model.Service, err error) {
	a := strings.FieldsFunc(str, Split)
	l := len(a)
	if l < 2 || l > 3 {
		return service, fmt.Errorf("invalid target: %s (eg: cube sqlcmd -x mssql1://172.16.157.163:1434 -usa -p123456aa -e \"whoami\")", str)
	}

	service.Schema = a[0]
	service.Ip = a[1]
	if !ValidIp(service.Ip) {
		return service, fmt.Errorf("invalid ip: %s", service.Ip)
	}

	if len(a) == 2 {
		service.Port = model.CommonPortMap[service.Schema]
	} else {
		service.Port, _ = strconv.Atoi(a[2])
	}

	return service, nil
}

func SliceContain(str string, slice []string) bool {
	for _, value := range slice {
		if str == value {
			return true
		}
	}
	return false
}

func FileReader(filename string) ([]string, error) {
	file, err := os.Open(filename)
	if err != nil {
		log.Errorf("Open file %s error, %v\n", filename, err)
	}
	defer file.Close()
	var content []string
	scanner := bufio.NewScanner(file)
	scanner.Split(bufio.ScanLines)
	for scanner.Scan() {
		text := strings.TrimSpace(scanner.Text())
		if text != "" {
			content = append(content, scanner.Text())
		}
	}
	return content, nil
}

func SameStringSlice(x, y []string) bool {
	if len(x) != len(y) {
		return false
	}
	// create a map of string -> int
	diff := make(map[string]int, len(x))
	for _, _x := range x {
		// 0 value for int is 0, so just increment a counter for the string
		diff[_x]++
	}
	for _, _y := range y {
		// If the string _y is not in diff bail out early
		if _, ok := diff[_y]; !ok {
			return false
		}
		diff[_y] -= 1
		if diff[_y] == 0 {
			delete(diff, _y)
		}
	}
	return len(diff) == 0
}

func RemoveRepByMap(slc []string) []string {
	var result []string
	tempMap := map[string]byte{} // 存放不重复主键
	for _, e := range slc {
		l := len(tempMap)
		tempMap[e] = 0
		if len(tempMap) != l { // 加入map后，map长度变化，则元素不重复
			result = append(result, e)
		}
	}
	return result
}

func Subset(first, second []string) bool {
	set := make(map[string]int)
	for _, value := range second {
		set[value] += 1
	}

	for _, value := range first {
		if count, found := set[value]; !found {
			return false
		} else if count < 1 {
			return false
		} else {
			set[value] = count - 1
		}
	}
	return true
}

func MD5(s string) (m string) {
	h := md5.New()
	io.WriteString(h, s)
	return fmt.Sprintf("%x", h.Sum(nil))
}

func MakeTaskHash(k string) string {
	hash := MD5(k)
	return hash
}

func CheckTaskHash(hash string) bool {
	_, ok := model.SuccessHash[hash]
	//log.Debugf("Success: %#v\n", model.SuccessHash)
	return ok
}

func SetTaskHash(hash string) {
	model.Mutex.Lock()
	model.SuccessHash[hash] = true
	model.Mutex.Unlock()
}

// ResultMap 当Mysql或者redis空密码的时候，任何密码都正确，会导致密码刷屏
var ResultMap = struct {
	sync.RWMutex
	m map[string]string
}{m: make(map[string]string)}

func SetResultMap(r model.CrackTaskResult) {
	ResultMap.Lock()
	ResultMap.m[fmt.Sprintf("%s==>%s:%s", r.CrackTask.CrackPlugin, r.CrackTask.Ip, r.CrackTask.Port)] = r.Result
	ResultMap.Unlock()
}

func ReadResultMap() {
	ResultMap.RLock()
	n := ResultMap.m
	ResultMap.RUnlock()
	for k, v := range n {
		log.Infof("[*]: %s %v", k, v)
	}
}

func getFinishTime(t1 time.Time) {

	fmt.Println(strings.Repeat(">", 50))
	End := time.Now().Format("2006-01-02 15:04:05")
	fmt.Printf("Finished:%s  Cost:%s", End, time.Since(t1))

}

var mu *sync.RWMutex

func WriteToFile(f *os.File, output string) error {
	mu.Lock()
	_, err := f.WriteString(fmt.Sprintf("%s\n", output))
	mu.Unlock()
	if err != nil {
		return fmt.Errorf("[!] Unable to write to file %w", err)
	}
	return nil
}
