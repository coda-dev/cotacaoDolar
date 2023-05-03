package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"time"
)

type Dolar struct {
	ValorEmReal string `json:"bid"`
}

func main() {

	fmt.Printf("pegando o dolar \n")
	dolar, error := GetCotacaoDolarApi()
	if error != nil {
		panic(error)
	}

	fmt.Printf("grava no arquivo \n")
	GravaCotacaoNoArquivo(*dolar)
}

func GetCotacaoDolarApi() (*Dolar, error) {

	// trabalhando com contexto junto resquest, response e hhtp client
	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, 300*time.Millisecond)
	defer cancel()

	fmt.Printf("acesso a cotacao \n")
	req, err := http.NewRequestWithContext(ctx, "GET", "http://localhost:8080/cotacao", nil)
	if err != nil {
		return nil, err
	}

	fmt.Printf("resposta cotacao \n")
	resp, err := http.DefaultClient.Do(req) // DefaultClient é a mesma coisa que http.client{}
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	fmt.Printf("pegando a cotacao \n")
	var d Dolar
	err = json.Unmarshal(body, &d)
	if err != nil {
		return nil, err
	}

	fmt.Printf("retornando a cotacao \n")
	return &d, nil

}

func GravaCotacaoNoArquivo(dolar Dolar) {

	// criando arquivo
	file, err := os.Create("cotacao.txt")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Erro ao criar arquivo: %v\n", err)
	}
	defer file.Close()
	_, err = file.WriteString(fmt.Sprintf("Dólar: %s", dolar.ValorEmReal))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Erro ao gravar arquivo: %v\n", err)
	}
	fmt.Println("Arquivo criado com sucesso!")

}
