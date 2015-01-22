package main

import (
	"crypto/sha1"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
)

type (
	Config struct {
		Urls []string `json:"urls"`
	}
)

var (
	outdir = flag.String("outdir", ".", "directory to output files to")
	help   = flag.Bool("help", false, "show usage")
)

func main() {

	flag.Parse()

	if *help {
		fmt.Println("usage: go-getter [-flags]")
		flag.PrintDefaults()
		os.Exit(0)
	}

	file, err := os.Open("config.json")
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	var config Config
	if err := json.NewDecoder(file).Decode(&config); err != nil {
		log.Fatal(err)
	}

	if *outdir != "." {
		os.MkdirAll(*outdir, 0777)
	}

	// fmt.Printf("%#v", config)

	for _, url := range config.Urls {

		// get the file
		resp, err := http.Get(url)
		if err != nil {
			log.Fatal(err)
		}
		if resp.StatusCode != http.StatusOK {
			log.Fatalf("Expected 200, got %d %s", resp.StatusCode, resp.Status)
		}

		// create outfile
		h := sha1.New()
		h.Write([]byte(url))
		filename := fmt.Sprintf("%x", h.Sum(nil))

		out, err := os.OpenFile(filepath.Join(*outdir, filename), os.O_CREATE|os.O_WRONLY, 0666)
		if err != nil {
			log.Fatal(err)
		}

		// copy the contents
		if _, err := io.Copy(out, resp.Body); err != nil {
			log.Fatal(err)
		}

		out.Close()
		resp.Body.Close()

		// get url
		fmt.Println(filename)
	}

	log.Println("done!")
}
