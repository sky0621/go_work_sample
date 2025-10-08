package http

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/sky0621/go_work_sample/core/pkg/domain"
	"github.com/sky0621/go_work_sample/core/pkg/errs"
	"github.com/sky0621/go_work_sample/core/pkg/usecase"
	"github.com/sky0621/go_work_sample/scoring/pkg/grading"
)

// Handler exposes grading endpoints.
type Handler struct {
	grading *grading.Service
}

// NewHandler creates a handler instance.
func NewHandler(grading *grading.Service) *Handler {
	return &Handler{grading: grading}
}

// Register wires endpoints onto mux.
func (h *Handler) Register(mux *http.ServeMux) {
	mux.Handle("/api/teachers/", http.HandlerFunc(h.route))
}

func (h *Handler) route(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	parts := splitPath(strings.TrimPrefix(r.URL.Path, "/api/teachers/"))
	if len(parts) != 4 || parts[1] != "tests" || parts[3] != "grade" {
		writeError(w, http.StatusNotFound, "not found")
		return
	}

	teacherID := domain.TeacherID(parts[0])
	testID := domain.TestID(parts[2])

	var req struct {
		QuestionID string `json:"question_id"`
		StudentID  string `json:"student_id"`
		Score      int    `json:"score"`
		Feedback   string `json:"feedback"`
		Completed  bool   `json:"completed"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json payload")
		return
	}

	payload := usecase.GradeInput{
		TestID:     testID,
		QuestionID: domain.QuestionID(strings.TrimSpace(req.QuestionID)),
		StudentID:  domain.StudentID(strings.TrimSpace(req.StudentID)),
		Score:      req.Score,
		Feedback:   strings.TrimSpace(req.Feedback),
		Completed:  req.Completed,
	}

	result, err := h.grading.GradeAnswer(r.Context(), teacherID, payload)
	if err != nil {
		switch err {
		case errs.ErrTeacherNotFound, errs.ErrTestNotFound:
			writeError(w, http.StatusNotFound, err.Error())
			return
		case errs.ErrStudentNotFound, errs.ErrStudentNotAssigned, errs.ErrAnswerNotFound:
			writeError(w, http.StatusBadRequest, err.Error())
			return
		case errs.ErrForbiddenTeacher:
			writeError(w, http.StatusForbidden, err.Error())
			return
		default:
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"result_id":  string(result.ID),
		"answer_id":  string(result.AnswerID),
		"score":      result.Score,
		"feedback":   result.Feedback,
		"completed":  result.Completed,
		"created_at": result.CreatedAt,
		"updated_at": result.UpdatedAt,
	})
}

func splitPath(path string) []string {
	if path == "" {
		return nil
	}
	parts := strings.Split(path, "/")
	out := parts[:0]
	for _, p := range parts {
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}
