package main

import (
	"bytes"
	"fmt"
	"os/exec"
	"sync"
)

func main() {

	bufs := make([]bytes.Buffer, 3)
	wg := new(sync.WaitGroup)
	for i := 0; i < 3; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			cmd := exec.Command("cloudquery", "sync", "cq-config.yaml", "--log-console", "--log-level=error", "--telemetry-level=none")
			cmd.Stdout, cmd.Stderr = &bufs[i], &bufs[i]
			if err := cmd.Run(); err != nil {
				fmt.Println(err)
			}
		}()
	}

	wg.Wait()

	for _, buf := range bufs {
		fmt.Println(buf.String())
	}

}
