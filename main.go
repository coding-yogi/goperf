package main

import (
	"flag"
	"os"
	"runtime"
	"time"

	"github.com/coding-yogi/perftool/log"
	"github.com/coding-yogi/perftool/tests"
)

func main() {

	runtime.GOMAXPROCS(2)

	tc := flag.Int("threads", 10, "Thread Count")
	rut := flag.Int("rampup", 30, "Ramp up time in seconds")
	et := flag.Int("etime", 60, "Execution time in minutes")
	flag.Parse()

	//Check if execution time is more than ramp up time
	if *et*60 < *rut {
		log.Fatalln("Total execution time needs to be more than ramp up time")
	}

	//later need to move this somewhere else
	file, err := os.OpenFile("logs.txt", os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		log.Fatalln("Unable to create log file")
	}
	defer file.Close()
	log.SetOutput(file)

	waitTimeForRampUp := *rut / *tc

	log.Printf("Execution will happen with %d users with a ramp up time of %d seconds for %d minutes\n", *tc, *rut, *et)

	//goroutine to increment thread count and ramp up
	tt := *tc
	*tc = 0
	go func() {
		for ; *tc < tt; *tc++ {
			log.Printf("Thread Count %d", *tc)
			time.Sleep(time.Duration(waitTimeForRampUp) * time.Second)
		}
		log.Printf("Thread Count %d", *tc)
	}()

	//set timeout
	timeout := time.After(time.Duration(*et) * time.Second)
	ts := 1

	for {
		select {
		case <-timeout:
			log.Printf("Execution time completed")
			return
		default:
			if ts <= *tc {
				log.Printf("Thread No %d started", ts)
				ts++
				go func() {
					//Add tests to be perf tested
					tests.PerfTestParseDocument()
				}()
			}
		}
	}
}
