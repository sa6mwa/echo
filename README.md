# echo

A simple multi-tool TCP echo server for testing traffic patterns.

## Usage

```console
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
```

## Build

There is a `Makefile` to make life more files. `make` will run targets `clean`
and `build`. You will required Golang installed on your system. There is also a
Docker build using an intermediate build image (golang:1.19-alpine)...

```console
$ make
rm -f echo
CGO_ENABLED=0 go get -v -d ./...
CGO_ENABLED=0 go test -cover ./...
?       github.com/sa6mwa/echo/cmd/echo [no test files]
CGO_ENABLED=0 go build -v -ldflags '-s -X main.version=0' -o echo github.com/sa6mwa/echo/cmd/echo

# Docker build

$ make docker
docker build -t echo:latest .
Sending build context to Docker daemon  5.369MB
Step 1/13 : FROM golang:1.19-alpine AS builder
.
.
.
Step 13/13 : ENTRYPOINT [ "/app/echo" ]
 ---> Using cache
 ---> fe7bb4f7d65a
Successfully built fe7bb4f7d65a
Successfully tagged echo:latest
```

If you would like another tag, you can change repository and tag using the
Makefile variables `REPO` and `TAG`...

```console
$ make docker REPO=my-ecr-registry/echo TAG=0
.
.
.
Step 13/13 : ENTRYPOINT [ "/app/echo" ]
 ---> Using cache
 ---> fe7bb4f7d65a
Successfully built fe7bb4f7d65a
Successfully tagged my-ecr-registry/echo:0
```

From here you can push the image to your target repository...

```console
$ docker push my-ecr-registry/echo:latest
```

## Examples

Using `-httpcounter` with the docker container...

```
$ docker run -ti echo -httpcounter

   ____    __
  / __/___/ /  ___
 / _// __/ _ \/ _ \
/___/\__/_//_/\___/ v4.9.0
High performance, minimalist Go web framework
https://echo.labstack.com
____________________________________O/_______
                                    O\
⇨ http server started on 0.0.0.0:8080
{"time":"2022-11-09T21:50:36.565849978Z","id":"","remote_ip":"172.31.0.1","host":"172.31.0.2:8080","method":"GET","uri":"/ping","user_agent":"curl/7.68.0","status":200,"error":"","latency":52015,"latency_human":"52.015µs","bytes_in":0,"bytes_out":44
```

In another window...

```console
$ curl 172.31.0.2:8080/ping
{"remoteAddr":"172.31.0.1:50800","count":1}
$ 
```

To find out which IP the docker container has, you might do something like...
```console
$ docker inspect `docker ps | grep /app/echo | awk '{print $1}'` | jq -r .[].NetworkSettings.IPAddress
172.31.0.2
```

Subsequent requests increases the counter...

```console
$ while true ; do curl 172.31.0.2:8080/ping ; sleep 1 ; done
{"remoteAddr":"172.31.0.1:37948","count":2}
{"remoteAddr":"172.31.0.1:37952","count":3}
{"remoteAddr":"172.31.0.1:37958","count":4}
{"remoteAddr":"172.31.0.1:37966","count":5}
{"remoteAddr":"172.31.0.1:37982","count":6}
{"remoteAddr":"172.31.0.1:37990","count":7}
```

Another example using `socat` and `-counter`...

```console
$ docker run -ti echo -counter
2022/11/09 21:54:11 echo -counter is running on 0.0.0.0:8080
```

Client-side...

```console
$ while true ; do socat - tcp4:172.31.0.2:8080 ; sleep 0.5 ; done
1
2
3
4
```

Below is a testing scenario. This may appear when a single-instance container is
restarted and the `socat` loop continues (a daemon pod/task or e.g when AWS ECS
deploymentConfiguration says MinimumHealthPercent=0 and MaximumPercent=100). In
case of a replica container scheduler, two pods/tasks will run simultaneously
and when the new container is ready, traffic will switch from the old to the new
pod/task through the LB and you should only notice that the count restarts at 1
again (no downtime).

```console
$ while true ; do socat - tcp4:172.31.0.2:8080 ; sleep 0.5 ; done
6
7
8
9
10
2022/11/09 22:58:57 socat[1400263] E connect(5, AF=2 172.31.0.2:8080, 16): Connection refused
2022/11/09 22:59:00 socat[1400488] E connect(5, AF=2 172.31.0.2:8080, 16): Connection refused
1
2
3
4
5
```
