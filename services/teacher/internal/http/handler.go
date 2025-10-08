package http

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/sky0621/go_work_sample/core/pkg/domain"
	"github.com/sky0621/go_work_sample/core/pkg/errs"
	"github.com/sky0621/go_work_sample/core/pkg/usecase"
	"github.com/sky0621/go_work_sample/scoring/pkg/grading"
)

// Handler exposes teacher-facing endpoints.
type Handler struct {
	assessments *usecase.AssessmentService
	grading     *grading.Service
}

// NewHandler builds a handler with required services.
func NewHandler(assessments *usecase.AssessmentService, grading *grading.Service) *Handler {
	return &Handler{assessments: assessments, grading: grading}
}

// Register wires HTTP endpoints.
func (h *Handler) Register(mux *http.ServeMux) {
	mux.Handle("/api/teachers/", http.HandlerFunc(h.route))
}

func (h *Handler) route(w http.ResponseWriter, r *http.Request) {
	parts := splitPath(strings.TrimPrefix(r.URL.Path, "/api/teachers/"))
	if len(parts) == 0 {
		writeError(w, http.StatusNotFound, "not found")
		return
	}

	teacherID := domain.TeacherID(parts[0])

	if len(parts) == 2 && parts[1] == "tests" {
		switch r.Method {
		case http.MethodPost:
			h.createTest(w, r, teacherID)
			return
		case http.MethodGet:
			h.listTests(w, r, teacherID)
			return
		}
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
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
			h.getQuestions(w, r, teacherID, testID)
			return
		case "answers":
			if r.Method != http.MethodGet {
				writeError(w, http.StatusMethodNotAllowed, "method not allowed")
				return
			}
			h.listAnswers(w, r, teacherID, testID)
			return
		case "results":
			if r.Method != http.MethodGet {
				writeError(w, http.StatusMethodNotAllowed, "method not allowed")
				return
			}
			h.listResults(w, r, teacherID, testID)
			return
		case "grade":
			if r.Method != http.MethodPost {
				writeError(w, http.StatusMethodNotAllowed, "method not allowed")
				return
			}
			h.gradeAnswer(w, r, teacherID, testID)
			return
		}
	}

	writeError(w, http.StatusNotFound, "not found")
}

type createTestRequest struct {
	Title     string `json:"title"`
	Questions []struct {
		Prompt string `json:"prompt"`
		Points int    `json:"points"`
	} `json:"questions"`
	StudentIDs []string `json:"student_ids"`
}

type testResponse struct {
	TestID     string             `json:"test_id"`
	Title      string             `json:"title"`
	CreatedAt  time.Time          `json:"created_at"`
	UpdatedAt  time.Time          `json:"updated_at"`
	StudentIDs []string           `json:"student_ids"`
	Questions  []questionResponse `json:"questions"`
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

func (h *Handler) createTest(w http.ResponseWriter, r *http.Request, teacherID domain.TeacherID) {
	var req createTestRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json payload")
		return
	}

	input := usecase.CreateTestInput{
		Title:     strings.TrimSpace(req.Title),
		TeacherID: teacherID,
	}

	for _, q := range req.Questions {
		input.Questions = append(input.Questions, usecase.QuestionDraft{
			Prompt: strings.TrimSpace(q.Prompt),
			Points: q.Points,
		})
	}

	for _, sid := range req.StudentIDs {
		sid = strings.TrimSpace(sid)
		if sid == "" {
			continue
		}
		input.StudentIDs = append(input.StudentIDs, domain.StudentID(sid))
	}

	test, questions, err := h.assessments.CreateTest(r.Context(), input)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, toTestResponse(*test, questions))
}

func (h *Handler) listTests(w http.ResponseWriter, r *http.Request, teacherID domain.TeacherID) {
	tests, err := h.assessments.ListTestsByTeacher(r.Context(), teacherID)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	payload := make([]testResponse, 0, len(tests))
	for _, test := range tests {
		questions, qErr := h.assessments.GetQuestionsForTeacher(r.Context(), teacherID, test.ID)
		if qErr != nil {
			handleServiceError(w, qErr)
			return
		}
		payload = append(payload, toTestResponse(test, questions))
	}

	writeJSON(w, http.StatusOK, map[string]any{"tests": payload})
}

func (h *Handler) getQuestions(w http.ResponseWriter, r *http.Request, teacherID domain.TeacherID, testID domain.TestID) {
	questions, err := h.assessments.GetQuestionsForTeacher(r.Context(), teacherID, testID)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	resp := make([]questionResponse, len(questions))
	for i, q := range questions {
		resp[i] = questionResponse{
			QuestionID: string(q.ID),
			Sequence:   q.Sequence,
			Prompt:     q.Prompt,
			Points:     q.Points,
			CreatedAt:  q.CreatedAt,
		}
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"test_id":   string(testID),
		"questions": resp,
	})
}

func (h *Handler) listAnswers(w http.ResponseWriter, r *http.Request, teacherID domain.TeacherID, testID domain.TestID) {
	answers, err := h.assessments.ListAnswersByTest(r.Context(), teacherID, testID)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	resp := make([]answerResponse, len(answers))
	for i, ans := range answers {
		resp[i] = answerResponse{
			AnswerID:   string(ans.ID),
			QuestionID: string(ans.QuestionID),
			StudentID:  string(ans.StudentID),
			Response:   ans.Response,
			CreatedAt:  ans.CreatedAt,
			UpdatedAt:  ans.UpdatedAt,
		}
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"test_id": string(testID),
		"answers": resp,
	})
}

func (h *Handler) listResults(w http.ResponseWriter, r *http.Request, teacherID domain.TeacherID, testID domain.TestID) {
	results, err := h.assessments.ListResultsByTest(r.Context(), teacherID, testID)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	resp := make([]resultResponse, len(results))
	for i, res := range results {
		resp[i] = resultResponse{
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
		"results": resp,
	})
}

func (h *Handler) gradeAnswer(w http.ResponseWriter, r *http.Request, teacherID domain.TeacherID, testID domain.TestID) {
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
		handleServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, resultResponse{
		ResultID:  string(result.ID),
		AnswerID:  string(result.AnswerID),
		Score:     result.Score,
		Feedback:  result.Feedback,
		Completed: result.Completed,
		CreatedAt: result.CreatedAt,
		UpdatedAt: result.UpdatedAt,
	})
}

func toTestResponse(test domain.Test, questions []domain.Question) testResponse {
	resp := testResponse{
		TestID:     string(test.ID),
		Title:      test.Title,
		CreatedAt:  test.CreatedAt,
		UpdatedAt:  test.UpdatedAt,
		StudentIDs: make([]string, len(test.AssignedTo)),
		Questions:  make([]questionResponse, len(questions)),
	}

	for i, sid := range test.AssignedTo {
		resp.StudentIDs[i] = string(sid)
	}

	for i, q := range questions {
		resp.Questions[i] = questionResponse{
			QuestionID: string(q.ID),
			Sequence:   q.Sequence,
			Prompt:     q.Prompt,
			Points:     q.Points,
			CreatedAt:  q.CreatedAt,
		}
	}

	return resp
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
	case errs.ErrTeacherNotFound, errs.ErrTestNotFound:
		writeError(w, http.StatusNotFound, err.Error())
	case errs.ErrStudentNotFound, errs.ErrStudentNotAssigned, errs.ErrInvalidTest, errs.ErrInvalidQuestion, errs.ErrInvalidAnswer, errs.ErrAnswerNotFound:
		writeError(w, http.StatusBadRequest, err.Error())
	case errs.ErrForbiddenTeacher:
		writeError(w, http.StatusForbidden, err.Error())
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
