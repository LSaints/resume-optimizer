package main

import (
	routes "backend/internal/http"
	"log/slog"
	"net/http"

	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		slog.Warn("Arquivo .env não encontrado, usando variáveis de ambiente do sistema")
	}

	router := routes.RegisterRoutes()

	slog.Info("Servidor iniciado na porta 8080")

	err := http.ListenAndServe(":8080", router)
	if err != nil {
		slog.Error("%s", "Não foi possivel inciar o servidor: ", err.Error())
	}
}
