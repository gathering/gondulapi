package main

import (
	_ "github.com/gathering/gondulapi/db"
	_ "github.com/gathering/gondulapi/objects"
	"github.com/gathering/gondulapi/receiver"
)

func main() {
	receiver.Start()
}
