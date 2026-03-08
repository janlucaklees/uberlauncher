package main

import (
	"context"
	"fmt"
	"os"

	"uberlauncher/internal/core"
	refreshjob "uberlauncher/internal/jobs/refresh_apps_cache"
)

func main() {
	if len(os.Args) >= 3 && os.Args[1] == "__internal" && os.Args[2] == "refresh-app-cache" {
		if err := refreshjob.Run(); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		return
	}

	if err := core.Run(context.Background()); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
