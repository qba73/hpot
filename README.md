[![Go Reference](https://pkg.go.dev/badge/github.com/qba73/hpot.svg)](https://pkg.go.dev/github.com/qba73/hpot)
[![Go Report Card](https://goreportcard.com/badge/github.com/qba73/hpot)](https://goreportcard.com/report/github.com/qba73/hpot)
[![Tests](https://github.com/qba73/hpot/actions/workflows/go.yml/badge.svg)](https://github.com/qba73/hpot/actions/workflows/go.yml)
![GitHub go.mod Go version](https://img.shields.io/github/go-mod/go-version/qba73/hpot)
![GitHub](https://img.shields.io/github/license/qba73/hpot)


# hpot

`hpot` is a command line tool for running a honeypot server.

Here's how to install it:

```sh
go install github.com/qba73/hpot/cmd/hpot@latest
```

To run it:

Start `HPot`
```sh
hpot 8091,8092,8093
```

Send a request
```sh
curl http://localhost:8091
curl: (56) Recv failure: Connection reset by peer
```

## Verbose mode

To see more information about incoming connections, use the `-v` flag:

```sh
hpot -v 8091,8092,8093
```
```
Starting listener on: [::]:8091
Starting listener on: [::]:8092
Starting listener on: [::]:8093
```

Send a request
```sh
curl http://localhost:8091
```
```sh
Incomming connection from:  127.0.0.1:53771
```
