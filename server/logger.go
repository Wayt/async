package server

import (
	"log"
	"os"
)

var logger = log.New(os.Stderr, "[server] ", log.LstdFlags)
