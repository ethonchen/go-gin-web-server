//package main
//
//import (
//	"fmt"
//	"github.com/gin-gonic/gin"
//	"log"
//	"os"
//	"runtime"
//)
//
//func main() {
//	ConfigRuntime()
//	StartWorkers()
//	StartGin()
//}
//
//// ConfigRuntime sets the number of operating system threads.
//func ConfigRuntime() {
//	nuCPU := runtime.NumCPU()
//	runtime.GOMAXPROCS(nuCPU)
//	fmt.Printf("Running with %d CPUs\n", nuCPU)
//}
//
//// StartWorkers start starsWorker by goroutine.
//func StartWorkers() {
//	go statsWorker()
//}
//
//// StartGin starts gin web server with setting router.
//func StartGin() {
//	gin.SetMode(gin.ReleaseMode)
//
//	router := gin.New()
//	router.Use(rateLimit, gin.Recovery())
//	router.LoadHTMLGlob("resources/*.templ.html")
//	router.Static("/static", "resources/static")
//	router.GET("/", index)
//	router.GET("/room/:roomid", roomGET)
//	router.POST("/room-post/:roomid", roomPOST)
//	router.GET("/stream/:roomid", streamRoom)
//
//	port := os.Getenv("PORT")
//	if port == "" {
//		port = "8080"
//	}
//	if err := router.Run(":" + port); err != nil {
//		log.Panicf("error: %s", err)
//	}
//}

package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"
)

var mm map[string]string = make(map[string]string)
var msges []string = make([]string, 0)

// GO_ENABLED=0 GOOS=linux GOARCH=amd64 go build .
//var lru = NewLruList(50)

var lastGetTime int64 = 0
var lastCode string = ""

// *handlers* 是 `net/http` 服务器里面的一个基本概念。
// handler 对象实现了 `http.Handler` 接口。
// 编写 handler 的常见方法是，在具有适当签名的函数上使用 `http.HandlerFunc` 适配器。
func save(w http.ResponseWriter, req *http.Request) {

	// handler 函数有两个参数，`http.ResponseWriter` 和 `http.Request`。
	// response writer 被用于写入 HTTP 响应数据，这里我们简单的返回 "hello\n"。
	values := req.URL.Query()
	//skey := values.Get("session")
	//svalue := values.Get("values")
	//mm[skey] = svalue
	jsonStr, err := json.Marshal(values)
	if err != nil {
		fmt.Errorf("err parse json")
	} else {
		fmt.Printf("#%s \r\n", jsonStr)
		if len(msges) > 100 {
			msges = msges[1:]
		}
		msges = append(msges, string(jsonStr))
	}

	//values.(map[string]string)
	for _, v := range values {
		v1, err := url.QueryUnescape(v[0])
		if err != nil {
			fmt.Errorf("%s -> %v \r\n", v, err)
		}
		//fmt.Println(k, v1)
		//compileRegex := regexp.MustCompile("登录(.*?)，哈哈哈")
		compileRegex := regexp.MustCompile("登录(.*?)为(.*?)，验证码") // 正则表达式的分组，以括号()表示，每一对括号就是我们匹配到的一个文本，可以把他们提取出来。
		matchArr := compileRegex.FindStringSubmatch(v1)
		if len(matchArr) > 0 {
			lastGetTime = time.Now().Unix()
			lastCode = strings.TrimSpace(matchArr[2])
			fmt.Println("login code ", lastCode)
		}

		task := regexp.MustCompile("下载的任务名为(.*?)的压缩包，解压密码为(.*?)，下载时间")
		arr1 := task.FindStringSubmatch(v1)
		if len(arr1) > 0 {
			mm[strings.TrimSpace(arr1[1])] = strings.TrimSpace(arr1[2])
			fmt.Printf("download pwd '%s' = '%s'", strings.TrimSpace(arr1[1]), strings.TrimSpace(arr1[2]))
		}

	}

	fmt.Fprintf(w, "ok\n")
}

func get(w http.ResponseWriter, req *http.Request) {

	values := req.URL.Query()
	skey := values.Get("session")
	if skey == "login" {
		tnow := time.Now().Unix()
		if tnow-lastGetTime < 300 {
			fmt.Fprintf(w, lastCode)
		}
	} else {
		if mm[skey] != "" {
			fmt.Fprintf(w, mm[skey])
		}
	}

}

func message(w http.ResponseWriter, req *http.Request) {
	values := req.URL.Query()
	skey := values.Get("number")
	vnum := 10
	if skey != "" {
		i, err := strconv.Atoi(skey)
		if err == nil {
			vnum = i
		}
	}
	vms := msges
	if len(msges) > vnum {
		vms = msges[len(msges)-vnum:]
	}
	jsonStr, err := json.Marshal(vms)
	if err == nil {
		fmt.Fprintf(w, string(jsonStr))
	}
}

func main() {

	// 使用 `http.HandleFunc` 函数，可以方便的将我们的 handler 注册到服务器路由。
	// 它是 `net/http` 包中的默认路由，接受一个函数作为参数。
	http.HandleFunc("/save", save)
	http.HandleFunc("/get", get)
	http.HandleFunc("/message", message)

	// 最后，我们调用 `ListenAndServe` 并带上端口和 handler。
	// nil 表示使用我们刚刚设置的默认路由器。
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	http.ListenAndServe(":"+port, nil)
}
