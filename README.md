![](https://gowok.github.io/docs/gowok-logo-docs.png)
[![made-with-Go](https://img.shields.io/badge/Made%20with-Go-1f425f.svg)](http://golang.org)
[![Go Version](https://img.shields.io/github/go-mod/go-version/gowok/gowok.svg)](https://github.com/gowok/gowok)
[![GoDoc Reference](https://img.shields.io/badge/godoc-reference-blue.svg)](https://pkg.go.dev/github.com/gowok/gowok)
[![GoReportCard](https://goreportcard.com/badge/github.com/gowok/gowok)](https://goreportcard.com/report/github.com/gowok/gowok)
[![Coverage Status](https://codecov.io/gh/gowok/gowok/branch/master/graph/badge.svg?token=7ORUPOWS3I)](https://codecov.io/gh/gowok/gowok)


Gowok is a library that contains a lot of functions that help you to build [Go](https://go.dev) project.

It has some utilities like:
* config loader,
* project bootstrapper,
* HTTP response builder,
* nil safety,
* password hash, and so on.

# Getting Started
## Installation
Run this command inside your project.
```bash
go get github.com/gowok/gowok
```

## Import in Project
In your `main.go`, write code like following example:
```go
package main

import "github.com/gowok/gowok"

func main() {
	gowok.Get().Run()
}
```

## Create Configuration
1. Create a YAML file named `config.yaml`.
2. Then write this.
```yaml
web:
  enabled: true
  host: :8080
```

## Running the Project
Run this command inside your project.
```bash
go run main.go
```

Or if config file not on root dir, use flag `--config`
``` bash
go run main.go --config=folder/gowok.yaml
```

It will show output like this:
```
2025/01/13 10:43:09 INFO starting web
```

Your project now ready to use 🔥

Let's try to send a request using `curl`!

```bash
curl localhost:8080
```

It will show output like this:
```
404 page not found
```

It means that your project already run.
It can receive actual request and give response to it.

# How to Contribute
Feel free to raise an issue or lovely pull request 😊
