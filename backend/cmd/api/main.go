package main

import (
	routes "backend/internal/http"
	"log"
	"net/http"

	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("Arquivo .env não encontrado, usando variáveis de ambiente do sistema")
	}

	router := routes.RegisterRoutes()

	log.Println("Servidor iniciado na porta 8080")

	err := http.ListenAndServe(":8080", router)
	if err != nil {
		log.Fatalf("%s", "Não foi possivel inciar o servidor: "+err.Error())
	}
}
