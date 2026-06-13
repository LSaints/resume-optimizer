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

	jobRepository := repositories.NewJobRepository(db)
	jobServices := services.NewJobServices(jobRepository)
	jobHandler := handlers.NewJobHandler(jobServices)

	optimizationRepository := repositories.NewOptimizationRepository(db)
	geminiClient := services.NewGeminiClient()
	optimizationServices := services.NewOptimizationServices(optimizationRepository, resumeRepository, jobRepository, geminiClient)
	optimizationHandler := handlers.NewOptimizationHandler(optimizationServices)

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

	mux.Handle(
		"POST /v1/jobs",
		authMiddleware.Middleware(
			http.HandlerFunc(jobHandler.Create),
		),
	)
	mux.Handle(
		"GET /v1/jobs",
		authMiddleware.Middleware(
			http.HandlerFunc(jobHandler.List),
		),
	)
	mux.Handle(
		"GET /v1/jobs/{id}",
		authMiddleware.Middleware(
			http.HandlerFunc(jobHandler.GetByID),
		),
	)
	mux.Handle(
		"PUT /v1/jobs/{id}",
		authMiddleware.Middleware(
			http.HandlerFunc(jobHandler.Update),
		),
	)
	mux.Handle(
		"DELETE /v1/jobs/{id}",
		authMiddleware.Middleware(
			http.HandlerFunc(jobHandler.Delete),
		),
	)

	mux.Handle(
		"POST /v1/resumes/{resumeID}/optimize",
		authMiddleware.Middleware(
			http.HandlerFunc(optimizationHandler.Optimize),
		),
	)
	mux.Handle(
		"GET /v1/resumes/{resumeID}/optimizations",
		authMiddleware.Middleware(
			http.HandlerFunc(optimizationHandler.ListByResume),
		),
	)
	mux.Handle(
		"GET /v1/resumes/{resumeID}/optimizations/{optimizationID}",
		authMiddleware.Middleware(
			http.HandlerFunc(optimizationHandler.GetByID),
		),
	)

	return mux
}
