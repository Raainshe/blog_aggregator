package main

import (
	"fmt"

	"github.com/raainshe/blog_aggregator/internal/config"
)

func main() {

	CFG, err := config.Read()
	if err != nil {
		return
	}
	CFG.SetUser("ryan")
	CFG, err = config.Read()
	if err != nil {
		return
	}
	fmt.Printf("%v", CFG)
}
