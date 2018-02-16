package main

import (
	"context"
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/wayt/async"
	"github.com/wayt/async/api"
	cli "github.com/wayt/async/client/async"
)

func init() {
	cli.Func("/v1/say-hello-world", func(ctx context.Context) error {

		fmt.Println("Hello World !")

		return nil
	})

	cli.Func("/v1/test-1", func(ctx context.Context) error {

		fmt.Println("Test 1")
		return nil
	})

	cli.Func("/v1/test-2", func(ctx context.Context) error {

		fmt.Println("Test 1")
		return nil
	})

	cli.Func("/v1/test-fail", func(ctx context.Context) error {

		fmt.Println("Test fail")
		return fmt.Errorf("Test fail")
	})

	// Run a test client
	go cli.Run()
}

func main() {
	jm := async.NewJobManager(
		async.NewJobRepository(),
		async.NewFunctionExecutor(fmt.Sprintf("http://127.0.0.1:%d", cli.Port)),
	)

	go jm.BackgroundProcess()

	// exec := executor.NewExecutor(jobs, workflows, functions,
	// 	map[string]executor.FunctionExecutor{
	// 		"default": executor.NewDefaultFunctionExecutor(
	// 	})
	// for i := 0; i < runtime.NumCPU(); i++ {
	// 	w := worker.NewWorker(jobs, exec)
	// 	go w.Work()
	// }

	e := gin.Default()

	api.NewHttpJobHandler(e, jm)

	e.Run(":8080")
}
