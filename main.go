package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/crowdmob/goamz/aws"
	"github.com/crowdmob/goamz/s3"
)

type (
	Config struct {
		RequestList []Request `json:"requestList"`
	}
	Request struct {
		URL      string `json:"url"`
		Interval int64  `json:"requestInterval"`
		FileName string `json:"fileName"`
	}
)

var (
	outdir = flag.String("outdir", ".", "directory to output files to")
	help   = flag.Bool("help", false, "show usage")

	workers = make(chan struct{}, 10)

	auth = aws.Auth{
		AccessKey: os.Getenv("AWS_ACCESS_KEY_ID"),
		SecretKey: os.Getenv("AWS_SECRET_ACCESS_KEY"),
	}

	connection = s3.New(auth, aws.USWest2)
	mybucket   = connection.Bucket("apiassets")
)

func main() {

	flag.Parse()
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	if *help {
		fmt.Println("usage: go-getter [-flags]")
		flag.PrintDefaults()
		os.Exit(0)
	}

	data, err := mybucket.Get("config.json")
	if err != nil {
		log.Fatal(err)
	}

	// open config.json
	var config Config
	if err := json.Unmarshal(data, &config); err != nil {
		log.Fatal(err)
	}

	// fmt.Printf("%s", data)

	for _, req := range config.RequestList {
		workers <- struct{}{}
		go refreshJson(req)
	}
	// to ensure workers are finished
	for i := 0; i < cap(workers); i++ {
		workers <- struct{}{}
	}

	log.Println("done!")
}

func refreshJson(req Request) error {
	defer func() { <-workers }()
	//if interval has passed go get json from url
	resp, err := mybucket.Head("apiresponse/"+req.FileName, nil)
	if err == nil && resp.StatusCode == http.StatusOK {
		t, err := time.Parse(time.RFC1123, resp.Header.Get("Last-Modified"))
		if err == nil && time.Now().Add(-time.Duration(req.Interval)*time.Second).After(t) {
			return nil
		}
	}

	// get the file
	resp, err = http.Get(req.URL)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Expected 200, got %d %s", resp.StatusCode, resp.Status)
	}
	data, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		return err
	}

	err = mybucket.Put("apiresponse/"+req.FileName, data, "application/json", s3.BucketOwnerFull, s3.Options{})
	if err != nil {
		return err
	}
	return nil
}
