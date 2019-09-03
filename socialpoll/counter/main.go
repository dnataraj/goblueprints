package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/nsqio/go-nsq"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

var fatalErr error

const updateDuration = 1 * time.Second

func fatal(e error) {
	fmt.Println(e)
	flag.PrintDefaults()
	fatalErr = e
}

func doCount(ctx context.Context, countsLock *sync.Mutex, counts *map[string]int, polldata *mongo.Collection) {
	countsLock.Lock()
	defer countsLock.Unlock()
	if len(*counts) == 0 {
		log.Println("no new votes, skipping database update")
		return
	}
	log.Println("updating database...")
	log.Println(*counts)
	ok := true
	for option, count := range *counts {
		sel := bson.M{"options": bson.M{"$in": []string{option}}}
		up := bson.M{"$inc": bson.M{"results." + option: count}}
		if _, err := polldata.UpdateMany(ctx, sel, up); err != nil {
			log.Println("failed to update: ", err)
			ok = false
		}
	}
	if ok {
		log.Println("finished updating database...")
		*counts = nil
	}
}

func main() {
	defer func() {
		if fatalErr != nil {
			os.Exit(1)
		}
	}()

	log.Println("dialing mongodb: localhost")
	ctx, cancelFunc := context.WithCancel(context.Background())
	client, err := mongo.Connect(ctx, options.Client().ApplyURI("mongodb://localhost:27017"))
	if err != nil {
		fatal(err)
		return
	}
	defer func() {
		log.Println("closing db connection...")
		err := client.Disconnect(ctx)
		if err != nil {
			log.Println("dialing mongodb: failed to disconnect client")
		}
		return
	}()

	pollData := client.Database("ballots").Collection("polls")

	var counts map[string]int
	var countsLock sync.Mutex

	log.Println("connecting to nsq...")
	q, err := nsq.NewConsumer("votes", "counter", nsq.NewConfig())
	if err != nil {
		fatal(err)
		return
	}
	q.AddHandler(nsq.HandlerFunc(func(m *nsq.Message) error {
		countsLock.Lock()
		defer countsLock.Unlock()
		if counts == nil {
			counts = make(map[string]int)
		}
		vote := string(m.Body)
		counts[vote]++
		return nil
	}))
	if err := q.ConnectToNSQLookupd("localhost:4161"); err != nil {
		fatal(err)
		return
	}

	ticker := time.NewTicker(updateDuration)
	termChan := make(chan os.Signal, 1)
	signal.Notify(termChan, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)
	for {
		select {
		case <-ticker.C:
			doCount(ctx, &countsLock, &counts, pollData)
		case <-termChan:
			ticker.Stop()
			q.Stop()
			cancelFunc()
		case <-q.StopChan:
			return
		}
	}

}
