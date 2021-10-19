package main

import (
	"context"
	"fmt"
	"sync"

	dockerapi "github.com/docker/docker/client"
)

func main() {
	client, err := dockerapi.NewClientWithOpts(dockerapi.FromEnv)
	client.NegotiateAPIVersion(context.Background())

	if err != nil {
		fmt.Printf("error %v", err)
	}

	for i := 0; i < 40; i++ {
		info, err := client.Info(context.TODO())
		if err != nil {
			fmt.Printf("error: %v", err)
		}
		fmt.Printf("info: %s", info.Architecture)
	}

	var wg sync.WaitGroup
	for i := 0; i < 40; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			info, err := client.Info(context.TODO())
			if err != nil {
				fmt.Printf("error: %v", err)
			}

			fmt.Printf("info: %s", info.Architecture)
		}()

	}

	wg.Wait()
}