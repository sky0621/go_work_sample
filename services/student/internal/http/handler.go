package http

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/sky0621/go_work_sample/core/pkg/domain"
	"github.com/sky0621/go_work_sample/core/pkg/errs"
	"github.com/sky0621/go_work_sample/core/pkg/usecase"
)

// Handler exposes student-facing endpoints.
type Handler struct {
	assessments *usecase.AssessmentService
}

// NewHandler builds a handler.
func NewHandler(assessments *usecase.AssessmentService) *Handler {
	return &Handler{assessments: assessments}
}

// Register wires endpoints.
func (h *Handler) Register(mux *http.ServeMux) {
	mux.Handle("/api/students/", http.HandlerFunc(h.route))
}

func (h *Handler) route(w http.ResponseWriter, r *http.Request) {
	parts := splitPath(strings.TrimPrefix(r.URL.Path, "/api/students/"))
	if len(parts) == 0 {
		writeError(w, http.StatusNotFound, "not found")
		return
	}

	studentID := domain.StudentID(parts[0])

	if len(parts) == 2 && parts[1] == "tests" {
		if r.Method != http.MethodGet {
			writeError(w, http.StatusMethodNotAllowed, "method not allowed")
			return
		}
		h.listTests(w, r, studentID)
		return
	}

	if len(parts) >= 4 && parts[1] == "tests" {
		testID := domain.TestID(parts[2])
		switch parts[3] {
		case "questions":
			if r.Method != http.MethodGet {
				writeError(w, http.StatusMethodNotAllowed, "method not allowed")
				return
			}
			h.getQuestions(w, r, studentID, testID)
			return
		case "answers":
			if r.Method != http.MethodPost {
				writeError(w, http.StatusMethodNotAllowed, "method not allowed")
				return
			}
			h.submitAnswer(w, r, studentID, testID)
			return
		case "results":
			if r.Method != http.MethodGet {
				writeError(w, http.StatusMethodNotAllowed, "method not allowed")
				return
			}
			h.listResults(w, r, studentID, testID)
			return
		}
	}

	writeError(w, http.StatusNotFound, "not found")
}

type testSummary struct {
	TestID    string    `json:"test_id"`
	Title     string    `json:"title"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type questionResponse struct {
	QuestionID string    `json:"question_id"`
	Sequence   int       `json:"sequence"`
	Prompt     string    `json:"prompt"`
	Points     int       `json:"points"`
	CreatedAt  time.Time `json:"created_at"`
}

type answerResponse struct {
	AnswerID   string    `json:"answer_id"`
	TestID     string    `json:"test_id"`
	QuestionID string    `json:"question_id"`
	StudentID  string    `json:"student_id"`
	Response   string    `json:"response"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

type resultResponse struct {
	ResultID  string    `json:"result_id"`
	AnswerID  string    `json:"answer_id"`
	Score     int       `json:"score"`
	Feedback  string    `json:"feedback"`
	Completed bool      `json:"completed"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (h *Handler) listTests(w http.ResponseWriter, r *http.Request, studentID domain.StudentID) {
	tests, err := h.assessments.ListTestsForStudent(r.Context(), studentID)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	payload := make([]testSummary, len(tests))
	for i, test := range tests {
		payload[i] = testSummary{
			TestID:    string(test.ID),
			Title:     test.Title,
			CreatedAt: test.CreatedAt,
			UpdatedAt: test.UpdatedAt,
		}
	}

	writeJSON(w, http.StatusOK, map[string]any{"tests": payload})
}

func (h *Handler) getQuestions(w http.ResponseWriter, r *http.Request, studentID domain.StudentID, testID domain.TestID) {
	questions, err := h.assessments.GetQuestionsForStudent(r.Context(), studentID, testID)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	payload := make([]questionResponse, len(questions))
	for i, q := range questions {
		payload[i] = questionResponse{
			QuestionID: string(q.ID),
			Sequence:   q.Sequence,
			Prompt:     q.Prompt,
			Points:     q.Points,
			CreatedAt:  q.CreatedAt,
		}
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"test_id":   string(testID),
		"questions": payload,
	})
}

func (h *Handler) submitAnswer(w http.ResponseWriter, r *http.Request, studentID domain.StudentID, testID domain.TestID) {
	var req struct {
		QuestionID string `json:"question_id"`
		Response   string `json:"response"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json payload")
		return
	}

	answer := &domain.Answer{
		TestID:     testID,
		QuestionID: domain.QuestionID(strings.TrimSpace(req.QuestionID)),
		StudentID:  studentID,
		Response:   strings.TrimSpace(req.Response),
	}

	saved, err := h.assessments.SubmitAnswer(r.Context(), answer)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusAccepted, answerResponse{
		AnswerID:   string(saved.ID),
		TestID:     string(saved.TestID),
		QuestionID: string(saved.QuestionID),
		StudentID:  string(saved.StudentID),
		Response:   saved.Response,
		CreatedAt:  saved.CreatedAt,
		UpdatedAt:  saved.UpdatedAt,
	})
}

func (h *Handler) listResults(w http.ResponseWriter, r *http.Request, studentID domain.StudentID, testID domain.TestID) {
	results, err := h.assessments.ListResultsForStudent(r.Context(), studentID, testID)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	payload := make([]resultResponse, len(results))
	for i, res := range results {
		payload[i] = resultResponse{
			ResultID:  string(res.ID),
			AnswerID:  string(res.AnswerID),
			Score:     res.Score,
			Feedback:  res.Feedback,
			Completed: res.Completed,
			CreatedAt: res.CreatedAt,
			UpdatedAt: res.UpdatedAt,
		}
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"test_id": string(testID),
		"results": payload,
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

func handleServiceError(w http.ResponseWriter, err error) {
	switch err {
	case errs.ErrStudentNotFound, errs.ErrTestNotFound:
		writeError(w, http.StatusNotFound, err.Error())
	case errs.ErrStudentNotAssigned, errs.ErrInvalidAnswer, errs.ErrQuestionNotFound:
		writeError(w, http.StatusBadRequest, err.Error())
	default:
		writeError(w, http.StatusInternalServerError, err.Error())
	}
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}
