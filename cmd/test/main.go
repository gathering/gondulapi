package main

import (
	"github.com/gathering/gondulapi/receiver"
	_ "github.com/gathering/gondulapi/objects"
)

func main() {
	receiver.Start()
}
