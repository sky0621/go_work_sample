package http

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/sky0621/go_work_sample/core/pkg/domain"
	"github.com/sky0621/go_work_sample/core/pkg/errs"
	"github.com/sky0621/go_work_sample/core/pkg/repository"
)

// Handler exposes read-only organization endpoints.
type Handler struct {
	org repository.OrganizationRepository
}

// NewHandler creates a handler instance.
func NewHandler(org repository.OrganizationRepository) *Handler {
	return &Handler{org: org}
}

// Register wires endpoints onto the mux.
func (h *Handler) Register(mux *http.ServeMux) {
	mux.Handle("/api/schools", h.handleSchools())
	mux.Handle("/api/grades/", http.HandlerFunc(h.handleGradeScoped))
	mux.Handle("/api/classes/", http.HandlerFunc(h.handleClassScoped))
	mux.Handle("/api/teachers/", http.HandlerFunc(h.handleTeacherScoped))
	mux.Handle("/api/students/", http.HandlerFunc(h.handleStudentScoped))
}

func (h *Handler) handleSchools() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/schools" && !strings.HasPrefix(r.URL.Path, "/api/schools/") {
			writeError(w, http.StatusNotFound, "not found")
			return
		}

		switch r.Method {
		case http.MethodGet:
			if r.URL.Path == "/api/schools" {
				h.listSchools(w, r)
				return
			}
			h.handleSchoolScoped(w, r)
		default:
			writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		}
	})
}

func (h *Handler) handleSchoolScoped(w http.ResponseWriter, r *http.Request) {
	parts := splitPath(strings.TrimPrefix(r.URL.Path, "/api/schools/"))
	if len(parts) == 0 {
		writeError(w, http.StatusNotFound, "not found")
		return
	}

	schoolID := domain.SchoolID(parts[0])

	if len(parts) == 1 {
		school, err := h.org.GetSchool(schoolID)
		if err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		if school == nil {
			writeError(w, http.StatusNotFound, errs.ErrSchoolNotFound.Error())
			return
		}
		writeJSON(w, http.StatusOK, school)
		return
	}

	if len(parts) == 2 {
		switch parts[1] {
		case "grades":
			grades, err := h.org.ListGrades(schoolID)
			if err != nil {
				writeError(w, http.StatusInternalServerError, err.Error())
				return
			}
			writeJSON(w, http.StatusOK, map[string]any{"grades": grades})
			return
		case "teachers":
			teachers, err := h.org.ListTeachers(schoolID)
			if err != nil {
				writeError(w, http.StatusInternalServerError, err.Error())
				return
			}
			writeJSON(w, http.StatusOK, map[string]any{"teachers": teachers})
			return
		}
	}

	writeError(w, http.StatusNotFound, "not found")
}

func (h *Handler) handleGradeScoped(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	parts := splitPath(strings.TrimPrefix(r.URL.Path, "/api/grades/"))
	if len(parts) == 0 {
		writeError(w, http.StatusNotFound, "not found")
		return
	}

	gradeID := domain.GradeID(parts[0])

	if len(parts) == 1 {
		grade, err := h.org.GetGrade(gradeID)
		if err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		if grade == nil {
			writeError(w, http.StatusNotFound, errs.ErrGradeNotFound.Error())
			return
		}
		writeJSON(w, http.StatusOK, grade)
		return
	}

	if len(parts) == 2 && parts[1] == "classes" {
		classes, err := h.org.ListClasses(gradeID)
		if err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"classes": classes})
		return
	}

	writeError(w, http.StatusNotFound, "not found")
}

func (h *Handler) handleClassScoped(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	parts := splitPath(strings.TrimPrefix(r.URL.Path, "/api/classes/"))
	if len(parts) == 0 {
		writeError(w, http.StatusNotFound, "not found")
		return
	}
	classID := domain.ClassID(parts[0])

	if len(parts) == 1 {
		class, err := h.org.GetClass(classID)
		if err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		if class == nil {
			writeError(w, http.StatusNotFound, errs.ErrClassNotFound.Error())
			return
		}
		writeJSON(w, http.StatusOK, class)
		return
	}

	if len(parts) == 2 && parts[1] == "students" {
		students, err := h.org.ListStudents(classID)
		if err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"students": students})
		return
	}

	writeError(w, http.StatusNotFound, "not found")
}

func (h *Handler) handleTeacherScoped(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	parts := splitPath(strings.TrimPrefix(r.URL.Path, "/api/teachers/"))
	if len(parts) != 1 {
		writeError(w, http.StatusNotFound, "not found")
		return
	}

	teacher, err := h.org.GetTeacher(domain.TeacherID(parts[0]))
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if teacher == nil {
		writeError(w, http.StatusNotFound, errs.ErrTeacherNotFound.Error())
		return
	}

	writeJSON(w, http.StatusOK, teacher)
}

func (h *Handler) handleStudentScoped(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	parts := splitPath(strings.TrimPrefix(r.URL.Path, "/api/students/"))
	if len(parts) != 1 {
		writeError(w, http.StatusNotFound, "not found")
		return
	}

	student, err := h.org.GetStudent(domain.StudentID(parts[0]))
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if student == nil {
		writeError(w, http.StatusNotFound, errs.ErrStudentNotFound.Error())
		return
	}

	writeJSON(w, http.StatusOK, student)
}

func (h *Handler) listSchools(w http.ResponseWriter, r *http.Request) {
	schools, err := h.org.ListSchools()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{"schools": schools})
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
