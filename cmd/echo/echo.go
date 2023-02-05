/*
A very simple multi-tool TCP echo server for testing traffic patterns

Copyright (c) 2022 SA6MWA Michel

	$ ./echo -h
	Usage of ./echo:
	  -counter
	        Server prints how many connections it has had and closes the connection
	  -host string
	        Host to bind server to, empty means binding to all interfaces (default "0.0.0.0")
	  -httpcounter
	        Server becomes a Labstack Echo http server printing number of requests served as html on / and as json on /ping
	  -oneliner
	        Disconnect after receiving one line
	  -port int
	        Port to listen to (default 8080)

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
*/
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
	"runtime/debug"
	"strings"
	"sync"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

var mutex sync.Mutex
var counter int64

var versionOverride string

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

	bi, gotBuildInfo := debug.ReadBuildInfo()
	version := []string{}
	modulePath := "echo"
	if versionOverride != "" {
		version = append(version, versionOverride)
	} else {
		if gotBuildInfo {
			if bi.Main.Version != "" {
				version = append(version, bi.Main.Version)
			}
			for _, setting := range bi.Settings {
				switch setting.Key {
				case "vcs.revision":
					version = append(version, setting.Value)
				case "vcs.modified":
					if strings.EqualFold(setting.Value, "true") {
						version = append(version, "dirty")
					}
				}
			}
		}
	}
	if gotBuildInfo {
		if bi.Main.Path != "" {
			modulePath = bi.Main.Path
		}
	}

	log.Printf("%s version %s", modulePath, strings.Join(version, "-"))

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
