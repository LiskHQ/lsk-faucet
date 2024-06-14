package main

import (
	"github.com/LiskHQ/lsk-faucet/cmd"
)

//go:generate npm run build
func main() {
	cmd.Execute()
}
