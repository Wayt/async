package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/wayt/async/async"
)

func init() {

	async.Func("/v1/say-hello-world", func(ctx context.Context) error {

		fmt.Println("Hello World !")

		return nil
	})

	async.Func("/v1/test-1", func(ctx context.Context) error {

		time.Sleep(1 * time.Second)
		fmt.Println("Test 1")
		return nil
	})

	async.Func("/v1/test-2", func(ctx context.Context) error {

		fmt.Println("Test 2")
		return nil
	})

	async.Func("/v1/test-fail", func(ctx context.Context) error {

		time.Sleep(1 * time.Second)
		fmt.Println("Test fail")
		return fmt.Errorf("Test fail")
	})
}

func main() {
	// Run a test async
	if err := async.Run(); err != nil {
		log.Fatal(err)
	}
}
