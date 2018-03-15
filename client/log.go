package client

import (
	"log"
	"os"
)

var logger = log.New(os.Stderr, "[client] ", log.LstdFlags)
