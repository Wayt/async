package main

import (
	"context"
	"fmt"
	"time"

	"github.com/wayt/async/client"
)

func init() {

	client.Func("/v1/say-hello-world", func(ctx context.Context) error {

		fmt.Println("Hello World !")

		return nil
	})

	client.Func("/v1/test-1", func(ctx context.Context) error {

		time.Sleep(1 * time.Second)
		fmt.Println("Test 1")
		return nil
	})

	client.Func("/v1/test-2", func(ctx context.Context) error {

		fmt.Println("Test 2")
		return nil
	})

	client.Func("/v1/test-fail", func(ctx context.Context) error {

		time.Sleep(1 * time.Second)
		fmt.Println("Test fail")
		return fmt.Errorf("Test fail")
	})
}

func main() {
	// Run a test client
	client.Run()
}
