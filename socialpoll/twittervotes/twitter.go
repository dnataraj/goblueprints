package main

import (
	"bufio"
	"encoding/json"
	"github.com/gomodule/oauth1/oauth"
	"io"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

var conn net.Conn
var reader io.ReadCloser
var (
	authClient *oauth.Client
	creds      *oauth.Credentials
)
var (
	authSetupOnce sync.Once
	httpClient    *http.Client
)

type tweet struct {
	Text string
}

func dial(netw, addr string) (net.Conn, error) {
	if conn != nil {
		conn.Close()
		conn = nil
	}
	netc, err := net.DialTimeout(netw, addr, 5*time.Second)
	if err != nil {
		return nil, err
	}
	conn = netc
	return netc, nil
}

func closeConn() {
	if conn != nil {
		conn.Close()
	}
	if reader != nil {
		reader.Close()
	}
}

func setupTwitterAuth() {
	creds = &oauth.Credentials{
		Token:  os.Getenv("SP_TWITTER_ACCESSTOKEN"),
		Secret: os.Getenv("SP_TWITTER_ACCESSSECRET"),
	}
	authClient = &oauth.Client{
		Credentials: oauth.Credentials{
			Token:  os.Getenv("SP_TWITTER_KEY"),
			Secret: os.Getenv("SP_TWITTER_SECRET"),
		},
	}
}

func makeRequest(req *http.Request, params url.Values) (*http.Response, error) {
	authSetupOnce.Do(func() {
		setupTwitterAuth()
		httpClient = &http.Client{
			Transport: &http.Transport{
				Dial: dial,
			},
		}
	})
	formEnc := params.Encode()
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Content-Length", strconv.Itoa(len(formEnc)))
	authClient.SetAuthorizationHeader(req.Header, creds, "POST", req.URL, params)
	return httpClient.Do(req)
}

func readFromTwitter(votes chan<- string) {
	options, err := loadOptions()
	if err != nil {
		log.Println("failed to load options: ", err)
		return
	}
	log.Println("read in options for processing : ", options)
	u, err := url.Parse("https://stream.twitter.com/1.1/statuses/filter.json")
	if err != nil {
		log.Println("creating filter request failed: ", err)
		return
	}
	hashtags := make([]string, len(options))
	for i := range options {
		hashtags[i] = "#" + strings.ToLower(options[i])
	}
	query := make(url.Values)
	query.Set("track", strings.Join(hashtags, ","))
	req, err := http.NewRequest("POST", u.String(), strings.NewReader(query.Encode()))
	if err != nil {
		log.Println("creating filter request failed: ", err)
		return
	}
	resp, err := makeRequest(req, query)
	if err != nil {
		log.Println("making request failed: ", err)
		return
	}
	if resp.StatusCode != http.StatusOK {
		s := bufio.NewScanner(resp.Body)
		s.Scan()
		log.Println(s.Text())
		return
	}
	reader := resp.Body
	decoder := json.NewDecoder(reader)
	for {
		var t tweet
		if err := decoder.Decode(&t); err != nil {
			log.Println("error reading tweet : ", err)
			break
		}
		for _, option := range options {
			if strings.Contains(strings.ToLower(t.Text), strings.ToLower(option)) {
				log.Printf("tweet : %s, vote: %s\n", t.Text, option)
				votes <- option
			}
		}
	}
}

func startTwitterStream(stopchan <-chan struct{}, votes chan<- string) <-chan struct{} {
	stoppedchan := make(chan struct{}, 1)
	go func() {
		defer func() {
			stoppedchan <- struct{}{}
		}()
		for {
			select {
			case <-stopchan:
				log.Println("stopping twitter...")
				return
			default:
				log.Println("querying twitter...")
				readFromTwitter(votes)
				log.Println(" (waiting)")
				time.Sleep(10 * time.Second) // wait before reconnecting
			}
		}

	}()
	return stoppedchan
}
