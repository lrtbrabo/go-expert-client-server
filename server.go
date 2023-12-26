package main

import (
	"context"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type Cotacao struct {
	Code       string `json:"code"`
	Codein     string `json:"codein"`
	Name       string `json:"name"`
	High       string `json:"high"`
	Low        string `json:"low"`
	VarBid     string `json:"varBid"`
	PctChange  string `json:"pctChange"`
	Bid        string `json:"bid"`
	Ask        string `json:"ask"`
	Timestamp  string `json:"timestamp"`
	CreateDate string `json:"create_date"`
	gorm.Model
}

func main() {
  db, err := NewDb()
  if err != nil {
    log.Panicln(err)
  }
  db.AutoMigrate(&Cotacao{})

	mux := http.NewServeMux()
	mux.HandleFunc("/cotacao", HandleGetCotacao)
  log.Println("Servidor Inicializado")
	http.ListenAndServe(":8080", mux)
}

func HandleGetCotacao(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/cotacao" {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json")

	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, time.Millisecond*200)
	defer cancel()

	cotacao, err := GetCotacao(ctx)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Panicln(err)
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(cotacao)
}

func GetCotacao(ctx context.Context) (*Cotacao, error) {
	req, _ := http.NewRequestWithContext(ctx, "GET", "https://economia.awesomeapi.com.br/json/last/USD-BRL", nil)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}

	body, _ := io.ReadAll(resp.Body)
	var cotacaotemp map[string]Cotacao
	err = json.Unmarshal(body, &cotacaotemp)
	if err != nil {
		return nil, err
	}

	cotacao := cotacaotemp["USDBRL"]

  db, err := NewDb()
  db.AutoMigrate(&Cotacao{})
	err = WriteToDB(&cotacao, db)
	if err != nil {
		return nil, err
	}

	return &cotacao, nil
}

func NewDb() (*gorm.DB, error) {
	db, err := gorm.Open(sqlite.Open("gorm.db"), &gorm.Config{})
	return db, err
}

func WriteToDB(c *Cotacao, db *gorm.DB) error {
	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, time.Millisecond*10)
	defer cancel()

	err := db.WithContext(ctx).Create(&c).Error
	return err
}
