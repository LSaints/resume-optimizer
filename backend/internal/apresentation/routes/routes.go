package routes

import (
	"net/http"
	"os"

	"backend/internal/application/services"
	"backend/internal/apresentation/handlers"
	"backend/internal/apresentation/middleware"
	"backend/internal/infrastructure/data"
	"backend/internal/infrastructure/repositories"

	"github.com/rs/cors"
)

func RegisterRoutes() http.Handler {
	mux := http.NewServeMux()

	db := data.NewConnection()

	frontendUrl := os.Getenv("FRONTEND_URL")
	if frontendUrl == "" {
		frontendUrl = "http://localhost:5173"
	}

	c := cors.New(cors.Options{
		AllowedOrigins:   []string{frontendUrl},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		AllowCredentials: true,
		Debug:            true,
	})

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

	typstRenderService := services.NewTypstRenderService()
	renderHandler := handlers.NewRenderHandler(optimizationServices, typstRenderService)

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
	mux.Handle(
		"DELETE /v1/resumes/{resumeID}/optimizations/{optimizationID}",
		authMiddleware.Middleware(
			http.HandlerFunc(optimizationHandler.Delete),
		),
	)

	mux.Handle(
		"GET /v1/optimizations/{optimizationID}/render",
		authMiddleware.Middleware(
			http.HandlerFunc(renderHandler.RenderSVG),
		),
	)
	mux.HandleFunc("GET /v1/optimizations/{optimizationID}/render/pdf", renderHandler.RenderPDF)

	return c.Handler(mux)
}
