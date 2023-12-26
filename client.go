package main

import (
	"context"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"time"
  "os"
)

type Bid struct {
  Bid string `json:"bid"`
}


func main() {
  bid, err := GetBid()
  if err != nil {
    log.Panicln(err)
  }

  err = WriteToFile(bid)
  if err != nil {
    log.Panicln(err)
  }
}


func GetBid() (*Bid, error) {
  // Chama o server para recuperar o bid com timeout de 300 segundos
  ctx := context.Background()
  ctx, cancel := context.WithTimeout(ctx, time.Millisecond * 300)
  defer cancel()

  req, _ := http.NewRequestWithContext(ctx, "GET", "http://localhost:8080/cotacao", nil)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}

	body, _ := io.ReadAll(resp.Body)
	var bid Bid
	err = json.Unmarshal(body, &bid)
	if err != nil {
		return nil, err
	}
  return &bid, nil
}


func WriteToFile(bid *Bid) error {
  // Escreve a cotação do dólar no arquivo final
  write := []byte("Dólar: " + bid.Bid)
  err := os.WriteFile("arquivo.txt", write, 0644)
  if err != nil {
    log.Panicln(err)
    return err
  }
  return nil
}
