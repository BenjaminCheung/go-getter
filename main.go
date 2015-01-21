package main

import (
  "fmt"
  "net/http"
  "io/ioutil"
  "os"
)

func main() {
  url := "http://api.themoviedb.org/3/discover/movie?api_key=8441d2cd141a4f1530356e8634f3af99&sort_by=popularity.desc&vote_count.gte=50&year=2000"
  response, err := http.Get(url)
  if err != nil {
      fmt.Printf("%s", err)
      os.Exit(1)
  } else {
      defer response.Body.Close()
      contents, err := ioutil.ReadAll(response.Body)
      if err != nil {
          fmt.Printf("%s", err)
          os.Exit(1)
      }
      fmt.Printf("%s\n", string(contents))
  }
}