//+build load

package main

import (
	"flag"

	"github.com/lukkor/chessex/chessex"
)

func main() {
	cfgPath := flag.String("cfg", "", "config file path")
	printCfg := flag.Bool("print-cfg", false, "used to print current configuration")
	flag.Parse()

	s := chessex.NewService(*cfgPath, *printCfg, true)
	s.Run()
}
