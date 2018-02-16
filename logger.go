package async

import (
	"log"
	"os"
)

var logger = log.New(os.Stderr, "[asyncd] ", log.LstdFlags)
