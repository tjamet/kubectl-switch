package main

import (
	"github.com/tjamet/kubectl-switch/pkg/kswitch"
)

func main() {
	// backward compatibility, should deprecate at some point
	kswitch.Main()
}
