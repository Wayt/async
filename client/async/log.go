package async

import (
	"log"
	"os"
)

var logger = log.New(os.Stderr, "[async] ", log.LstdFlags)
