package tests

import (
	"time"

	"github.com/coding-yogi/perftool/services/fileparser"
)

// PerfTestParseDocument ..
func PerfTestParseDocument() {
	r := fileparser.Reply{}
	r.CreateQueue()
	for {
		r.Parse()
		time.Sleep(100 * time.Millisecond)
	}
}
