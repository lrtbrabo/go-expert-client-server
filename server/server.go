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

type BidResponse struct {
  Bid string `json:"bid"`
}

func main() {
  db, err := NewDb()
  if err != nil {
    log.Panicln(err)
  }
  db.AutoMigrate(&Cotacao{})

	mux := http.NewServeMux()
	mux.HandleFunc("/cotacao", func(w http.ResponseWriter, r *http.Request){
    HandleGetCotacao(w, r, db)
  })
  log.Println("Servidor Inicializado")
	http.ListenAndServe(":8080", mux)
}

func HandleGetCotacao(w http.ResponseWriter, r *http.Request, db *gorm.DB) {
	if r.URL.Path != "/cotacao" { 
    w.WriteHeader(http.StatusNotFound)
		return
	}
  if r.Method != "GET" {
    w.WriteHeader(http.StatusMethodNotAllowed)
    return
  }
	w.Header().Set("Content-Type", "application/json")
  ctx := context.Background()

	cotacao, err := GetCotacao(ctx, db)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Panicln(err)
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(cotacao)
}

func GetCotacao(ctx context.Context, db *gorm.DB) (*BidResponse, error) {
  // Faz a chavamada para a API da economia.awesomeapi utilizando contexto 
  ctx, cancel := context.WithTimeout(ctx, time.Millisecond * 200)
  defer cancel()

	req, _ := http.NewRequestWithContext(ctx, "GET", "https://economia.awesomeapi.com.br/json/last/USD-BRL", nil)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}

  // Trata o retorno da API para remover o USDBRL e deixar em uma formatação mais limpa
	body, _ := io.ReadAll(resp.Body)
	var cotacaotemp map[string]Cotacao
	err = json.Unmarshal(body, &cotacaotemp)
	if err != nil {
		return nil, err
	}

	cotacao := cotacaotemp["USDBRL"]

  // Escreve a cotação no banco
	err = WriteToDB(&cotacao, ctx, db)
	if err != nil {
		return nil, err
	}

  // Limpa o resultado final apenas com o bid e retorna para o requisitante
  bidResponse := BidResponse{
    Bid: cotacao.Bid,
  }

	return &bidResponse, nil
}

func NewDb() (*gorm.DB, error) {
  // Cria referência para banco de dados sqlite
	db, err := gorm.Open(sqlite.Open("gorm.db"), &gorm.Config{})
	return db, err
}

func WriteToDB(c *Cotacao, ctx context.Context, db *gorm.DB) error {
  // Gera contexto para timeout de escrita no banco e escreve a referência de cotação
	ctx, cancel := context.WithTimeout(ctx, time.Millisecond*10)
	defer cancel()

	err := db.WithContext(ctx).Create(&c).Error
	return err
}
