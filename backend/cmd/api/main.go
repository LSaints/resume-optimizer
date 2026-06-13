package main

import (
	"backend/internal/apresentation/routes"
	"log"
	"net/http"
)

func main() {
	router := routes.RegisterRoutes()

	log.Println("Servidor iniciado na porta 8080")

	err := http.ListenAndServe(":8080", router)
	if err != nil {
		log.Fatalf("%s", "Não foi possivel inciar o servidor: "+err.Error())
	}
}
