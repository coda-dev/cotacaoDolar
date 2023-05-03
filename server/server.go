package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"time"

	"github.com/google/uuid"
	_ "github.com/mattn/go-sqlite3"
)

type CotDolarReal struct {
	CotDolarReal Cotacao `json:"USDBRL"`
}

type Cotacao struct {
	ID         string `gorm:"primaryKey"`  // uuid.New().String()
	Code       string `json:"code"`        // code "USD"
	Codein     string `json:"codein"`      // codein	"BRL"
	Name       string `json:"name"`        // name	"Dólar Americano/Real Brasileiro"
	High       string `json:"high"`        // high	"4.9644"
	Low        string `json:"low"`         // low	"4.8931"
	VarBid     string `json:"varBid"`      // varBid	"-0.0193"
	PctChange  string `json:"pctChange"`   // pctChange	"-0.39"
	Bid        string `json:"bid"`         // bid	"4.9065"
	Ask        string `json:"ask"`         // ask	"4.9092"
	Timestamp  string `json:"timestamp"`   // timestamp	"1681505996"
	CreateDate string `json:"create_date"` // create_date	"2023-04-14 17:59:56"
}

func main() {
	fmt.Printf("server: inicio \n")

	http.HandleFunc("/cotacao", BuscaCotacaoHandler)
	http.ListenAndServe(":8080", nil)
	fmt.Printf("server: fim \n")
}

func BuscaCotacaoHandler(w http.ResponseWriter, r *http.Request) {

	if r.URL.Path != "/cotacao" {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	cotacao, error := BuscaCotacao()
	if error != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Printf("API %s \n", error.Error())
		return
	}

	error = insereCotacao(cotacao)
	if error != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Printf("Banco %s \n", error.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	json.NewEncoder(w).Encode(cotacao)

}

func BuscaCotacao() (*Cotacao, error) {

	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	req, error := http.NewRequestWithContext(ctx, "GET", "https://economia.awesomeapi.com.br/json/last/USD-BRL", nil)
	if error != nil {
		return nil, error
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		panic(err)
	}

	body, error := ioutil.ReadAll(resp.Body)
	if error != nil {
		return nil, error
	}
	defer resp.Body.Close()

	io.Copy(os.Stdout, resp.Body)

	var c CotDolarReal
	error = json.Unmarshal(body, &c)
	if error != nil {
		return nil, error
	}

	return &c.CotDolarReal, nil

}

func insereCotacao(cotacao *Cotacao) error {

	db, err := criaBaseDadosCotacao()
	if err != nil {
		panic(err)
	}
	defer db.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	stmt, err := db.PrepareContext(ctx, "insert into cotacao(id,code,codein,name,high,low,varBid,pctChange,bid,ask,timestamp,create_date) values(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)")
	if err != nil {
		return err
	}
	defer stmt.Close()
	_, err = stmt.ExecContext(ctx, uuid.New().String(), cotacao.Code, cotacao.Codein, cotacao.Name, cotacao.High, cotacao.Low, cotacao.VarBid, cotacao.PctChange, cotacao.Bid, cotacao.Ask, cotacao.Timestamp, cotacao.CreateDate)
	if err != nil {
		return err
	}

	stmt, err = db.Prepare("select create_date from cotacao where create_date = ?")
	if err != nil {
		return err
	}

	var create_date string
	err = stmt.QueryRow(cotacao.CreateDate).Scan(&create_date)
	if err != nil {
		return err
	}

	fmt.Println("bid cad. " + create_date)

	return nil

}

func criaBaseDadosCotacao() (*sql.DB, error) {

	db, err := sql.Open("sqlite3", "cotacao.db")

	if err != nil {
		return nil, err
	}

	stmt := `
			DROP TABLE IF EXISTS cotacao;
			CREATE TABLE IF NOT EXISTS cotacao(id varchar(255), code varchar(15), codein varchar(15), name varchar(800), high varchar(15), low varchar(15), varBid varchar(15), pctChange varchar(15), bid varchar(15), ask varchar(15), timestamp varchar(15), create_date varchar(30), primary key(id));
			`
	_, err = db.Exec(stmt)

	if err != nil {
		return nil, err
	}

	fmt.Println("tabela cotacao criada")

	return db, nil

}
