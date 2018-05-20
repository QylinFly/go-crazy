package main

import (
    "flag"
    "fmt"
    "log"
    "net/http"
    "runtime"
	"time"
	Gin "github.com/gin-gonic/gin"
	
)

var (
    reqs int
	max  int
	
	
)

func init() {
    flag.IntVar(&reqs, "reqs", 30, "Total requests")
    flag.IntVar(&max, "concurrent", 3, "Maximum concurrent requests")
}

type ResponsePool struct {
	resp *http.Response
	cont *Gin.Context
    err error
}

type RequestPool struct {
	req *http.Request
	cont *Gin.Context
    err error
}


// Dispatcher
func dispatcher(reqChan chan RequestPool) {
	fmt.Printf("dispatcher in\n")
	
    // defer close(reqChan)
    for i := 0; i < reqs; i++ {
        req, err := http.NewRequest("GET", "http://10.99.2.116:8087/invoke", nil)
        if err != nil {
            log.Println(err)
		}
		r := RequestPool{req,nil ,err}
		reqChan <- r
		fmt.Printf("dispatcher reqChan <- req\n")
		// time.Sleep(time.Second)
    }
}

// Worker Pool
func workerPool(reqChan chan RequestPool, respChan chan ResponsePool) {
	t := &http.Transport{}
	fmt.Printf("workerPool in\n")
	
    for i := 0; i < max; i++ {
        go worker(t, reqChan, respChan)
    }
}
// Worker
func worker(t *http.Transport, reqChan chan RequestPool, respChan chan ResponsePool) {
	fmt.Printf("worker in\n")
	
    for req := range reqChan {
        resp, err := t.RoundTrip(req.req)
        r := ResponsePool{resp,nil, err}
		respChan <- r
		fmt.Printf("worker respChan <- r\n")
		
    }
}

// Consumer
func consumer(respChan chan ResponsePool) (int64, int64) {
    var (
        conns int64
        size  int64
    )
    for conns < int64(reqs) {
        select {
        case r,ok := <-respChan:
            if ok {
			fmt.Printf("consumer\n")
				
                if r.err != nil {
                    log.Println(r.err)
                } else {
                    size += r.resp.ContentLength
                    if err := r.resp.Body.Close(); err != nil {
                        log.Println(r.err)
                    }
                }
                conns++
            }
        }
    }
    return conns, size
}

func main22e() {
    flag.Parse()
	runtime.GOMAXPROCS(runtime.NumCPU())
	runtime.NumGoroutine()
    reqChan := make(chan RequestPool)
    respChan := make(chan ResponsePool)
    start := time.Now()
    go dispatcher(reqChan)
	go workerPool(reqChan, respChan)

	conns, size := consumer(respChan)
	

    took := time.Since(start)
    ns := took.Nanoseconds()
    av := ns / conns
    average, err := time.ParseDuration(fmt.Sprintf("%d", av) + "ns")
    if err != nil {
        log.Println(err)
    }
    fmt.Printf("Connections:\t%d\nConcurrent:\t%d\nTotal size:\t%d bytes\nTotal time:\t%s\nAverage time:\t%s\n", conns, max, size, took, average)
}