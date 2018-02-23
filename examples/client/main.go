package main

import (
	"context"
	"fmt"
	"time"

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
}

func main() {
	// Run a test client
	cli.Run()
}
