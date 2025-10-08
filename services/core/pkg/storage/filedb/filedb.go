package filedb

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"sync"

	"github.com/sky0621/go_work_sample/core/pkg/domain"
	"github.com/sky0621/go_work_sample/core/pkg/memory"
	"github.com/sky0621/go_work_sample/core/pkg/repository"
)

// Repository provides a JSON file backed implementation of repository interfaces.
type Repository struct {
	mu       sync.Mutex
	path     string
	delegate *memory.Repository
}

// Ensure interface compliance.
var (
	_ repository.OrganizationRepository = (*Repository)(nil)
	_ repository.TestRepository         = (*Repository)(nil)
	_ repository.AnswerRepository       = (*Repository)(nil)
	_ repository.ResultRepository       = (*Repository)(nil)
)

// NewRepository loads state from the provided path or seeds a new one.
func NewRepository(path string, seed memory.SeedData) (*Repository, error) {
	if path == "" {
		return nil, errors.New("filedb: path must be provided")
	}

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return nil, err
	}

	var delegate *memory.Repository
	if _, err := os.Stat(path); err == nil {
		state, loadErr := loadState(path)
		if loadErr != nil {
			return nil, loadErr
		}
		delegate = memory.NewRepositoryFromState(state)
	} else {
		delegate = memory.NewRepository(seed)
	}

	repo := &Repository{path: path, delegate: delegate}

	if _, err := os.Stat(path); errors.Is(err, os.ErrNotExist) {
		if err := repo.persist(); err != nil {
			return nil, err
		}
	}

	return repo, nil
}

// OrganizationRepository delegation.

func (r *Repository) ListSchools() ([]domain.School, error) {
	return r.delegate.ListSchools()
}

func (r *Repository) GetSchool(id domain.SchoolID) (*domain.School, error) {
	return r.delegate.GetSchool(id)
}

func (r *Repository) GetGrade(id domain.GradeID) (*domain.Grade, error) {
	return r.delegate.GetGrade(id)
}

func (r *Repository) GetClass(id domain.ClassID) (*domain.Class, error) {
	return r.delegate.GetClass(id)
}

func (r *Repository) GetTeacher(id domain.TeacherID) (*domain.Teacher, error) {
	return r.delegate.GetTeacher(id)
}

func (r *Repository) GetStudent(id domain.StudentID) (*domain.Student, error) {
	return r.delegate.GetStudent(id)
}

func (r *Repository) ListGrades(schoolID domain.SchoolID) ([]domain.Grade, error) {
	return r.delegate.ListGrades(schoolID)
}

func (r *Repository) ListClasses(gradeID domain.GradeID) ([]domain.Class, error) {
	return r.delegate.ListClasses(gradeID)
}

func (r *Repository) ListStudents(classID domain.ClassID) ([]domain.Student, error) {
	return r.delegate.ListStudents(classID)
}

func (r *Repository) ListTeachers(schoolID domain.SchoolID) ([]domain.Teacher, error) {
	return r.delegate.ListTeachers(schoolID)
}

// TestRepository delegation with persistence on mutations.

func (r *Repository) CreateTest(test *domain.Test, questions []domain.Question, studentIDs []domain.StudentID) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if err := r.delegate.CreateTest(test, questions, studentIDs); err != nil {
		return err
	}
	return r.persist()
}

func (r *Repository) UpdateTest(test *domain.Test) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if err := r.delegate.UpdateTest(test); err != nil {
		return err
	}
	return r.persist()
}

func (r *Repository) GetTest(id domain.TestID) (*domain.Test, error) {
	return r.delegate.GetTest(id)
}

func (r *Repository) ListTestsByTeacher(teacherID domain.TeacherID) ([]domain.Test, error) {
	return r.delegate.ListTestsByTeacher(teacherID)
}

func (r *Repository) ListTestsForStudent(studentID domain.StudentID) ([]domain.Test, error) {
	return r.delegate.ListTestsForStudent(studentID)
}

func (r *Repository) ListQuestions(testID domain.TestID) ([]domain.Question, error) {
	return r.delegate.ListQuestions(testID)
}

func (r *Repository) IsStudentAssigned(testID domain.TestID, studentID domain.StudentID) (bool, error) {
	return r.delegate.IsStudentAssigned(testID, studentID)
}

// AnswerRepository delegation with persistence.

func (r *Repository) UpsertAnswer(answer *domain.Answer) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if err := r.delegate.UpsertAnswer(answer); err != nil {
		return err
	}
	return r.persist()
}

func (r *Repository) GetAnswer(testID domain.TestID, questionID domain.QuestionID, studentID domain.StudentID) (*domain.Answer, error) {
	return r.delegate.GetAnswer(testID, questionID, studentID)
}

func (r *Repository) ListAnswers(testID domain.TestID, studentID domain.StudentID) ([]domain.Answer, error) {
	return r.delegate.ListAnswers(testID, studentID)
}

func (r *Repository) ListAnswersByTest(testID domain.TestID) ([]domain.Answer, error) {
	return r.delegate.ListAnswersByTest(testID)
}

// ResultRepository delegation with persistence.

func (r *Repository) SaveResult(result *domain.Result) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if err := r.delegate.SaveResult(result); err != nil {
		return err
	}
	return r.persist()
}

func (r *Repository) GetResult(answerID domain.AnswerID) (*domain.Result, error) {
	return r.delegate.GetResult(answerID)
}

func (r *Repository) ListResultsByTest(testID domain.TestID) ([]domain.Result, error) {
	return r.delegate.ListResultsByTest(testID)
}

func (r *Repository) ListResultsByStudent(testID domain.TestID, studentID domain.StudentID) ([]domain.Result, error) {
	return r.delegate.ListResultsByStudent(testID, studentID)
}

// Helpers.

func (r *Repository) persist() error {
	state := r.delegate.ExportState()
	tmp := r.path + ".tmp"

	file, err := os.Create(tmp)
	if err != nil {
		return err
	}
	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(state); err != nil {
		file.Close()
		return err
	}
	if err := file.Close(); err != nil {
		return err
	}
	return os.Rename(tmp, r.path)
}

func loadState(path string) (memory.State, error) {
	file, err := os.Open(path)
	if err != nil {
		return memory.State{}, err
	}
	defer file.Close()

	var state memory.State
	if err := json.NewDecoder(file).Decode(&state); err != nil {
		return memory.State{}, err
	}
	return state, nil
}

// Delegate exposes the underlying memory repository for testing purposes.
func (r *Repository) Delegate() *memory.Repository {
	return r.delegate
}
