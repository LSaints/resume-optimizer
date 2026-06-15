package http

import (
	"backend/internal/ats"
	"backend/internal/auth"
	"backend/internal/job"
	"backend/internal/render"
	"backend/internal/resume"
	resumeoptimized "backend/internal/resume_optimized"
	"backend/internal/user"
	"backend/pkg/ai"
	"backend/pkg/data"
	textextractor "backend/pkg/text_extractor"
	"net/http"
	"os"

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

	userRepository := user.NewUserRepository(db)
	userServices := user.NewUserServices(userRepository)
	userHandler := user.NewUserHandler(userServices)

	authServices := auth.NewAuthServices(userRepository)
	authHandler := auth.NewAuthHandler(authServices)
	authMiddleware := auth.NewAuthMiddleware(authServices)

	resumeRepository := resume.NewResumeRepository(db)
	textExtractor := textextractor.NewTextExtractor()
	resumeServices := resume.NewResumeServices(resumeRepository, textExtractor)
	resumeHandler := resume.NewResumeHandler(resumeServices)

	jobRepository := job.NewJobRepository(db)
	jobServices := job.NewJobServices(jobRepository)
	jobHandler := job.NewJobHandler(jobServices)

	optimizationRepository := resumeoptimized.NewOptimizationRepository(db)
	geminiClient := ai.NewGeminiClient()
	optimizationServices := resumeoptimized.NewOptimizationServices(optimizationRepository, resumeRepository, jobRepository, geminiClient)
	optimizationHandler := resumeoptimized.NewOptimizationHandler(optimizationServices)

	atsEvaluationRepository := ats.NewAtsEvaluationRepository(db)
	atsScoringServices := ats.NewAtsScoringServices(atsEvaluationRepository, resumeRepository, jobRepository, geminiClient)
	atsScoringHandler := ats.NewAtsScoringHandler(atsScoringServices)

	typstRenderService := render.NewTypstRenderService()
	renderHandler := render.NewRenderHandler(optimizationServices, typstRenderService)

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

	mux.Handle(
		"POST /v1/resumes/{resumeID}/evaluate",
		authMiddleware.Middleware(
			http.HandlerFunc(atsScoringHandler.Evaluate),
		),
	)
	mux.Handle(
		"GET /v1/resumes/{resumeID}/evaluations",
		authMiddleware.Middleware(
			http.HandlerFunc(atsScoringHandler.ListByResume),
		),
	)
	mux.Handle(
		"GET /v1/resumes/{resumeID}/evaluations/{evaluationID}",
		authMiddleware.Middleware(
			http.HandlerFunc(atsScoringHandler.GetByID),
		),
	)

	return c.Handler(mux)
}
