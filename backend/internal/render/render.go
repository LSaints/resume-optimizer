package render

import (
	"encoding/json"
	"net/http"

	"backend/internal/auth"
	resumeoptimized "backend/internal/resume_optimized"
)

type RenderHandler struct {
	OptService    *resumeoptimized.OptimizationServices
	RenderService *TypstRenderService
}

func NewRenderHandler(optService *resumeoptimized.OptimizationServices, renderService *TypstRenderService) *RenderHandler {
	return &RenderHandler{
		OptService:    optService,
		RenderService: renderService,
	}
}

func (h *RenderHandler) RenderSVG(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	userID, ok := r.Context().Value(auth.UserIDKey).(string)
	if !ok {
		http.Error(w, "não autorizado", http.StatusUnauthorized)
		return
	}
	optimizationID := r.PathValue("optimizationID")

	opt, err := h.OptService.GetByID(userID, optimizationID)
	if err != nil {
		status := http.StatusInternalServerError
		if err.Error() == "otimização não encontrada" {
			status = http.StatusNotFound
		}
		http.Error(w, err.Error(), status)
		return
	}

	svg, err := h.RenderService.RenderToSVG(opt.TypstContent)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadGateway)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(RenderResponse{SvgContent: svg})
}

func (h *RenderHandler) RenderPDF(w http.ResponseWriter, r *http.Request) {
	optimizationID := r.PathValue("optimizationID")

	opt, err := h.OptService.GetByIDPublic(optimizationID)
	if err != nil {
		status := http.StatusInternalServerError
		if err.Error() == "otimização não encontrada" {
			status = http.StatusNotFound
		}
		w.Header().Set("Content-Type", "application/json")
		http.Error(w, err.Error(), status)
		return
	}

	pdf, err := h.RenderService.RenderToPDF(opt.TypstContent)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		http.Error(w, err.Error(), http.StatusBadGateway)
		return
	}

	w.Header().Set("Content-Type", "application/pdf")
	w.Header().Set("Content-Disposition", `attachment; filename="curriculo-otimizado.pdf"`)
	w.WriteHeader(http.StatusOK)
	w.Write(pdf)
}
