package server

/*
 * Parsing and formatting functions
 */

import (
	"time"
)

func formatDate() string {
	return time.Now().Format(time.RFC822)
}
