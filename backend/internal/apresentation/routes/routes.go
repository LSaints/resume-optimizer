package routes

import (
	"net/http"

	"backend/internal/application/services"
	"backend/internal/apresentation/handlers"
	"backend/internal/apresentation/middleware"
	"backend/internal/infrastructure/data"
	"backend/internal/infrastructure/repositories"
)

func RegisterRoutes() http.Handler {
	mux := http.NewServeMux()

	db := data.NewConnection()

	userRepository := repositories.NewUserRepository(db)
	userServices := services.NewUserServices(userRepository)
	userHandler := handlers.NewUserHandler(userServices)

	authServices := services.NewAuthServices(userRepository)
	authHandler := handlers.NewAuthHandler(authServices)

	authMiddleware := middleware.NewAuthMiddleware(authServices)

	resumeRepository := repositories.NewResumeRepository(db)
	textExtractor := services.NewTextExtractor()
	resumeServices := services.NewResumeServices(resumeRepository, textExtractor)
	resumeHandler := handlers.NewResumeHandler(resumeServices)

	mux.HandleFunc("POST /v1/auth/login", authHandler.Login)

	mux.Handle(
		"GET /v1/users",
		authMiddleware.Middleware(
			http.HandlerFunc(userHandler.GetUsers),
		),
	)
	mux.Handle(
		"GET /v1/users/{id}",
		authMiddleware.Middleware(
			http.HandlerFunc(userHandler.GetUsersById),
		),
	)
	mux.Handle("POST /v1/users", http.HandlerFunc(userHandler.CreateUser))

	mux.Handle(
		"POST /v1/resumes",
		authMiddleware.Middleware(
			http.HandlerFunc(resumeHandler.Create),
		),
	)
	mux.Handle(
		"GET /v1/resumes",
		authMiddleware.Middleware(
			http.HandlerFunc(resumeHandler.List),
		),
	)
	mux.Handle(
		"GET /v1/resumes/{id}",
		authMiddleware.Middleware(
			http.HandlerFunc(resumeHandler.GetByID),
		),
	)
	mux.Handle(
		"PUT /v1/resumes/{id}",
		authMiddleware.Middleware(
			http.HandlerFunc(resumeHandler.Update),
		),
	)
	mux.Handle(
		"DELETE /v1/resumes/{id}",
		authMiddleware.Middleware(
			http.HandlerFunc(resumeHandler.Delete),
		),
	)

	return mux
}
