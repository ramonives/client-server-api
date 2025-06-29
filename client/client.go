package main

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"time"
)

type CotacaoResponse struct {
	Bid string `json:"bid"`
}

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 300*time.Millisecond)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", "http://localhost:8080/cotacao", nil)
	if err != nil {
		log.Fatal("erro criando request:", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Fatal("erro ao fazer request:", err)
	}
	defer resp.Body.Close()

	var cotacao CotacaoResponse
	if err := json.NewDecoder(resp.Body).Decode(&cotacao); err != nil {
		log.Fatal("erro ao decodificar JSON:", err)
	}

	content := "Dólar: " + cotacao.Bid
	err = ioutil.WriteFile("cotacao.txt", []byte(content), 0644)
	if err != nil {
		log.Fatal("erro ao escrever arquivo:", err)
	}

	log.Println("Cotação salva com sucesso:", content)
}
