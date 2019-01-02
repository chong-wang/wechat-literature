// +build ignore

package main

import (
	"fmt"
	"time"
)

func main() {
	LoadProgress(DefaultProgressFile)
	start := time.Now()
	p := GenImage()
	end := time.Now()
	fmt.Println(len(p))
	fmt.Println(end.Sub(start))
}
