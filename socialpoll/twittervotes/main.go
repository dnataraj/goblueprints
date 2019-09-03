package main

import (
	"context"
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

var client *mongo.Client

type poll struct {
	Options []string
}

func dialdb() (context.Context, error) {
	var err error
	log.Println("dialing mongodb: localhost")
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	client, err = mongo.Connect(ctx, options.Client().ApplyURI("mongodb://localhost:27017"))
	if err != nil {
		log.Println("dialing mongodb: failed to create session")
		return nil, err
	}
	return ctx, nil
}

func closedb(ctx context.Context) {
	err := client.Disconnect(ctx)
	if err != nil {
		log.Println("dialing mongodb: failed to disconnect client")
	}
	return
}

func loadOptions() ([]string, error) {
	var options []string
	var err error
	ctx, _ := context.WithTimeout(context.Background(), 15*time.Second)
	cur, err := client.Database("ballots").Collection("polls").Find(ctx, bson.D{})
	if err != nil {
		log.Println("loading options failed: ", err)
		return nil, err
	}
	defer cur.Close(ctx)
	var p poll
	for cur.Next(ctx) {
		err = cur.Decode(&p)
		options = append(options, p.Options...)
	}
	return options, err
}

func publishVotes(votes <-chan string) <-chan struct{} {
	stopchan := make(chan struct{}, 1)
	pub, _ := nsq.NewProducer("localhost:4150", nsq.NewConfig())
	go func() {
		for vote := range votes {
			pub.Publish("votes", []byte(vote)) // publish vote
		}
		log.Println("publisher: stopping")
		pub.Stop()
		log.Println("publisher: stopped")
		stopchan <- struct{}{}
	}()
	return stopchan
}

func main() {
	var stoplock sync.Mutex // protects stop
	stop := false
	stopChan := make(chan struct{}, 1)
	signalChan := make(chan os.Signal, 1)
	go func() {
		<-signalChan
		stoplock.Lock()
		stop = true
		stoplock.Unlock()
		log.Println("stopping...")
		stopChan <- struct{}{}
		closeConn()
	}()
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)

	var ctx context.Context
	var err error
	if ctx, err = dialdb(); err != nil {
		log.Fatalln("failed to dial mongodb: ", err)
	}
	defer closedb(ctx)

	// start things
	votes := make(chan string)
	publisherStoppedChan := publishVotes(votes)
	twitterStoppedChan := startTwitterStream(stopChan, votes)
	go func() {
		for {
			time.Sleep(1 * time.Minute)
			closeConn()
			stoplock.Lock()
			if stop {
				stoplock.Unlock()
				return
			}
			stoplock.Unlock()
		}
	}()
	<-twitterStoppedChan
	close(votes)
	<-publisherStoppedChan

}
