package main

import (
	"log"
	"os"

	"github.com/noypi/election2016/comelec"
)

func main() {
	if 2 > len(os.Args) {
		log.Fatal("invalid params")
	}
	fpath := os.Args[1]

	vm := comelec.NewComelecVm()
	if 2 < len(os.Args) {
		vm.VM().Set("$args", os.Args[2:])
	}

	vm.Include(fpath)

}
