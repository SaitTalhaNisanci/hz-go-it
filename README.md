# hz-go-it

# Running on Local Environment
## Prerequisites

* Please make sure that Docker is installed, after installation please verify `docker version` has same or higher versions 
```bash
$ docker version
Client:
 Version:      1.13.1
 API version:  1.26
 Go version:   go1.7.5
 Git commit:   092cba3
 Built:        Wed Feb  8 08:47:51 2017
 OS/Arch:      darwin/amd64

Server:
 Version:      1.13.1
 API version:  1.26 (minimum version 1.12)
 Go version:   go1.7.5
 Git commit:   092cba3
 Built:        Wed Feb  8 08:47:51 2017
 OS/Arch:      linux/amd64
 Experimental: false

```
* Please make sure that `docker compose` installed on your machine, you may follow [this guide](https://docs.docker.com/compose/install/)
After installation please verify docker compose version.
```bash
$ docker-compose version
docker-compose version 1.11.1, build 7c5d5e4
docker-py version: 2.0.2
CPython version: 2.7.12
OpenSSL version: OpenSSL 1.0.2j  26 Sep 2016
```

* Please make sure Go Development Environment already installed, you may follow [this guide](https://golang.org/doc/install)
After installation please verify go version as below or higher
 ```bash
$ go version
go version go1.9.2 darwin/amd64
```

## Building and Running Tests

* Traverse to `./acceptance` directory and run below commands

```bash
$ go get -u all
Please note that you may need to download all dependencies one by one if `go get -u all` does not work
$ go build

Build docker image to run IT tests in, therefore please traverse to root directory
$ docker build . -t go-it:1
Create a network first
$ docker network create hz-go-it
And run tests as below
$ docker run --network=hz-go-it --name=go-it -v $GOPATH:/go -v /var/run/docker.sock:/var/run/docker.sock -v <your-path-to-project>/acceptance:/local/source go-it:1
```
 Please refer to `Dockerfile` and `client.sh` for tests to run and details.
