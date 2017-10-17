package main

import (
	"flag"
	"log"
	"runtime"
	"time"

	"github.com/coding-yogi/perftool/services/fileparser"

	"github.com/coding-yogi/perftool/rmq"
)

func main() {

	runtime.GOMAXPROCS(2)

	tc := flag.Int("threads", 5, "Thread Count")
	rut := flag.Int("rampup", 30, "Ramp up time in seconds")
	et := flag.Int("etime", 60, "Execution time in minutes")
	flag.Parse()

	//Check if execution time is more than ramp up time
	if *et*60 < *rut {
		log.Fatalln("Total execution time needs to be more than ramp up time")
	}

	waitTimeForRampUp := *rut / *tc

	log.Printf("Execution will happen with %d users with a ramp up time of %d seconds for %d minutes\n", *tc, *rut, *et)

	amqpClient := rmq.Client{
		Host:     "localhost",
		Port:     5672,
		Username: "guest",
		Password: "guest",
	}

	amqpClient.NewConnection()
	defer amqpClient.CloseConnection()

	//Setup
	fileparser.Setup(&amqpClient)

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
	running := true
	ts := 1

	for running == true {
		select {
		case <-timeout:
			log.Printf("Execution time completed")
			running = false
		default:
			if ts <= *tc {
				log.Printf("Thread No %d started", ts)
				ts++
				go func() {
					//create reply queue for every thread
					r := fileparser.Reply{}
					r.CreateQueue()
					for {
						r.Parse()
						time.Sleep(200 * time.Millisecond)
					}
				}()
			}
		}
	}
}
