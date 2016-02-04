package main

import (
	"github.com/d-bf/client/config"
	"github.com/d-bf/client/term"
)

func init() {
	config.Check()
}

func main() {
	term.Clear()
}
