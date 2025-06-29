package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"time"

	_ "modernc.org/sqlite"
)

type Cotacao struct {
	Bid string `json:"bid"`
}

type Response struct {
	USDBRL Cotacao `json:"USDBRL"`
}

func main() {
	db, err := sql.Open("sqlite", "./cotacoes.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS cotacoes (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		valor TEXT,
		criado_em DATETIME DEFAULT CURRENT_TIMESTAMP
	)`)
	if err != nil {
		log.Fatal(err)
	}

	http.HandleFunc("/cotacao", func(w http.ResponseWriter, r *http.Request) {
		ctxAPI, cancelAPI := context.WithTimeout(r.Context(), 200*time.Millisecond)
		defer cancelAPI()

		req, err := http.NewRequestWithContext(ctxAPI, "GET", "https://economia.awesomeapi.com.br/json/last/USD-BRL", nil)
		if err != nil {
			log.Println("erro criando request para API:", err)
			http.Error(w, "erro interno", http.StatusInternalServerError)
			return
		}

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			log.Println("erro na chamada da API:", err)
			http.Error(w, "erro ao buscar cotacao", http.StatusGatewayTimeout)
			return
		}
		defer resp.Body.Close()

		var resultado Response
		if err := json.NewDecoder(resp.Body).Decode(&resultado); err != nil {
			log.Println("erro ao decodificar resposta:", err)
			http.Error(w, "erro ao ler cotacao", http.StatusInternalServerError)
			return
		}

		ctxDB, cancelDB := context.WithTimeout(context.Background(), 10*time.Millisecond)
		defer cancelDB()

		stmt, err := db.PrepareContext(ctxDB, "INSERT INTO cotacoes(valor) VALUES(?)")
		if err != nil {
			log.Println("erro preparando stmt:", err)
		} else {
			_, err = stmt.ExecContext(ctxDB, resultado.USDBRL.Bid)
			if err != nil {
				log.Println("erro salvando no banco:", err)
			}
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"bid": resultado.USDBRL.Bid})
	})

	log.Println("Servidor rodando na porta 8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
