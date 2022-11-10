package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

var mutex sync.Mutex
var counter int64

func main() {
	var host string
	var port int
	var onelinerMode, connectionCounterMode, httpCounterMode bool

	flag.StringVar(&host, "host", "0.0.0.0", "Host to bind server to, empty means binding to all interfaces")
	flag.IntVar(&port, "port", 8080, "Port to listen to")
	flag.BoolVar(&onelinerMode, "oneliner", false, "Disconnect after receiving one line")
	flag.BoolVar(&connectionCounterMode, "counter", false, "Server prints how many connections it has had and closes the connection")
	flag.BoolVar(&httpCounterMode, "httpcounter", false, "Server becomes a Labstack Echo http server printing number of requests served as html on / and as json on /ping")
	flag.Parse()
	addr := host + ":" + fmt.Sprint(port)

	if httpCounterMode {
		e := echo.New()
		e.Use(middleware.Logger())
		e.Use(middleware.Recover())
		e.GET("/", func(c echo.Context) error {
			mutex.Lock()
			counter++
			mutex.Unlock()
			return c.HTML(http.StatusOK, fmt.Sprintf("%d from %s\n", counter, c.Request().RemoteAddr))
		})
		e.GET("/ping", func(c echo.Context) error {
			mutex.Lock()
			counter++
			mutex.Unlock()
			return c.JSON(http.StatusOK, struct {
				RemoteAddr string `json:"remoteAddr"`
				Count      int64  `json:"count"`
			}{
				RemoteAddr: c.Request().RemoteAddr,
				Count:      counter,
			})
		})
		e.Logger.Fatal(e.Start(addr))
	} else {
		server, err := net.Listen("tcp", addr)
		if err != nil {
			log.Fatalln(err)
		}
		defer server.Close()

		if len(os.Args) > 1 {
			log.Printf("%s %s is running on %s", filepath.Base(os.Args[0]), strings.Join(os.Args[1:], " "), addr)
		} else {
			log.Printf("%s is running on %s (%s -h for syntax)", filepath.Base(os.Args[0]), addr, filepath.Base(os.Args[0]))
		}

		for {
			conn, err := server.Accept()
			if err != nil {
				log.Println("Failed to accept connection:", err)
				continue
			}
			log.Printf("Connect from %s", conn.RemoteAddr())
			go func(conn net.Conn) {
				defer func() {
					conn.Close()
				}()
				if connectionCounterMode {
					mutex.Lock()
					counter++
					mutex.Unlock()
					fmt.Fprintf(conn, "%d\n", counter)
				} else if onelinerMode {
					scanner := bufio.NewScanner(conn)
					scanner.Scan()
					io.WriteString(conn, scanner.Text()+"\n")
				} else {
					io.Copy(conn, conn)
				}
			}(conn)
		}
	}
}
