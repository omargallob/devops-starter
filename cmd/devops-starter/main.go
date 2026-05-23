// Command devops-starter is an opinionated cross-platform DevOps tool installer.
package main

import (
	"os"

	"github.com/omargallob/devops-starter/internal/cli"
)

func main() {
	cmd := cli.NewRootCmd()
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
