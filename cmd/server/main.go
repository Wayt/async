package main

import (
	"flag"
	"fmt"

	cli "github.com/wayt/async/client/async"
)

var (
	executorURL = flag.String("-e", fmt.Sprintf("http://127.0.0.1:%d", cli.Port), "Function executor API url")
	bind        = flag.String("-b", ":8080", "Server bind address")
)

func main() {
	flag.Parse()

}
