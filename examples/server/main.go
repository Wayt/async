package main

import (
	"context"
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/wayt/async"
	cli "github.com/wayt/async/client/async"
)

func init() {
	cli.Func("/v1/say-hello-world", func(ctx context.Context) error {

		fmt.Println("Hello World !")

		return nil
	})

	cli.Func("/v1/test-1", func(ctx context.Context) error {

		time.Sleep(5 * time.Second)
		fmt.Println("Test 1")
		return nil
	})

	cli.Func("/v1/test-2", func(ctx context.Context) error {

		fmt.Println("Test 1")
		return nil
	})

	cli.Func("/v1/test-fail", func(ctx context.Context) error {

		time.Sleep(15 * time.Second)
		fmt.Println("Test fail")
		return fmt.Errorf("Test fail")
	})

	// Run a test client
	go cli.Run()
}

func main() {
	jr := async.NewJobRepository()
	jm := async.NewJobManager(
		jr,
	)

	cr := async.NewCallbackRepository()
	cm := async.NewCallbackManager(cr)

	d := async.NewDispatcher(jm, cm, async.NewFunctionExecutor(fmt.Sprintf("http://127.0.0.1:%d", cli.Port)))
	p := async.NewPoller(jr)
	go p.Poll(d)

	rescheduler := async.NewExpiredRescheduler(jm, cm)
	go rescheduler.Run()

	e := gin.Default()

	async.NewHttpJobHandler(e, jm)
	async.NewHttpCallbackHandler(e, cm, jm)

	e.Run(":8080")
}
