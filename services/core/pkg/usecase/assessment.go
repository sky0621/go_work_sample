package usecase

import (
	"context"
	"sort"
	"time"

	"github.com/sky0621/go_work_sample/core/pkg/domain"
	"github.com/sky0621/go_work_sample/core/pkg/errs"
	"github.com/sky0621/go_work_sample/core/pkg/id"
	"github.com/sky0621/go_work_sample/core/pkg/repository"
)

// AssessmentService orchestrates teacher and student workflows around tests.
type AssessmentService struct {
	orgRepo    repository.OrganizationRepository
	testRepo   repository.TestRepository
	answerRepo repository.AnswerRepository
	resultRepo repository.ResultRepository
}

// NewAssessmentService constructs a service with shared repositories.
func NewAssessmentService(
	org repository.OrganizationRepository,
	test repository.TestRepository,
	answer repository.AnswerRepository,
	result repository.ResultRepository,
) *AssessmentService {
	return &AssessmentService{
		orgRepo:    org,
		testRepo:   test,
		answerRepo: answer,
		resultRepo: result,
	}
}

// CreateTestInput describes the data needed to author a test.
type CreateTestInput struct {
	Title      string
	TeacherID  domain.TeacherID
	Questions  []QuestionDraft
	StudentIDs []domain.StudentID
}

// QuestionDraft holds question details when creating a test.
type QuestionDraft struct {
	Prompt string
	Points int
}

// CreateTest registers a new test with questions and student assignments.
func (s *AssessmentService) CreateTest(ctx context.Context, input CreateTestInput) (*domain.Test, []domain.Question, error) {
	if input.Title == "" {
		return nil, nil, errs.ErrInvalidTest
	}
	if len(input.Questions) == 0 {
		return nil, nil, errs.ErrNoQuestions
	}

	teacher, err := s.orgRepo.GetTeacher(input.TeacherID)
	if err != nil {
		return nil, nil, err
	}
	if teacher == nil {
		return nil, nil, errs.ErrTeacherNotFound
	}

	for _, studentID := range input.StudentIDs {
		student, err := s.orgRepo.GetStudent(studentID)
		if err != nil {
			return nil, nil, err
		}
		if student == nil {
			return nil, nil, errs.ErrStudentNotFound
		}
	}

	now := time.Now().UTC()
	test := &domain.Test{
		ID:        domain.TestID(id.New()),
		TeacherID: input.TeacherID,
		Title:     input.Title,
		CreatedAt: now,
		UpdatedAt: now,
	}

	questions := make([]domain.Question, len(input.Questions))
	for i, q := range input.Questions {
		if q.Prompt == "" {
			return nil, nil, errs.ErrInvalidQuestion
		}
		questions[i] = domain.Question{
			ID:        domain.QuestionID(id.New()),
			TestID:    test.ID,
			Sequence:  i + 1,
			Prompt:    q.Prompt,
			Points:    q.Points,
			CreatedAt: now,
		}
	}

	if err := s.testRepo.CreateTest(test, questions, input.StudentIDs); err != nil {
		return nil, nil, err
	}

	test.AssignedTo = append([]domain.StudentID(nil), input.StudentIDs...)
	return test, questions, nil
}

// ListTestsByTeacher returns tests ordered by creation time.
func (s *AssessmentService) ListTestsByTeacher(ctx context.Context, teacherID domain.TeacherID) ([]domain.Test, error) {
	if err := s.ensureTeacherExists(teacherID); err != nil {
		return nil, err
	}

	tests, err := s.testRepo.ListTestsByTeacher(teacherID)
	if err != nil {
		return nil, err
	}

	sort.Slice(tests, func(i, j int) bool {
		return tests[i].CreatedAt.Before(tests[j].CreatedAt)
	})

	return tests, nil
}

// ListAnswersByTest returns answers for a test ensuring teacher ownership.
func (s *AssessmentService) ListAnswersByTest(ctx context.Context, teacherID domain.TeacherID, testID domain.TestID) ([]domain.Answer, error) {
	if err := s.ensureTeacherOwnsTest(teacherID, testID); err != nil {
		return nil, err
	}

	answers, err := s.answerRepo.ListAnswersByTest(testID)
	if err != nil {
		return nil, err
	}

	sort.Slice(answers, func(i, j int) bool {
		return answers[i].CreatedAt.Before(answers[j].CreatedAt)
	})

	return answers, nil
}

// ListResultsByTest returns grading results for a test ensuring teacher ownership.
func (s *AssessmentService) ListResultsByTest(ctx context.Context, teacherID domain.TeacherID, testID domain.TestID) ([]domain.Result, error) {
	if err := s.ensureTeacherOwnsTest(teacherID, testID); err != nil {
		return nil, err
	}

	results, err := s.resultRepo.ListResultsByTest(testID)
	if err != nil {
		return nil, err
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].CreatedAt.Before(results[j].CreatedAt)
	})

	return results, nil
}

// ListTestsForStudent returns assigned tests for a student.
func (s *AssessmentService) ListTestsForStudent(ctx context.Context, studentID domain.StudentID) ([]domain.Test, error) {
	if err := s.ensureStudentExists(studentID); err != nil {
		return nil, err
	}

	tests, err := s.testRepo.ListTestsForStudent(studentID)
	if err != nil {
		return nil, err
	}

	sort.Slice(tests, func(i, j int) bool {
		return tests[i].CreatedAt.Before(tests[j].CreatedAt)
	})

	return tests, nil
}

// GetQuestionsForTeacher returns questions ensuring teacher access.
func (s *AssessmentService) GetQuestionsForTeacher(ctx context.Context, teacherID domain.TeacherID, testID domain.TestID) ([]domain.Question, error) {
	if err := s.ensureTeacherOwnsTest(teacherID, testID); err != nil {
		return nil, err
	}
	return s.listQuestions(testID)
}

// GetQuestionsForStudent returns questions ensuring assignment.
func (s *AssessmentService) GetQuestionsForStudent(ctx context.Context, studentID domain.StudentID, testID domain.TestID) ([]domain.Question, error) {
	if err := s.ensureStudentExists(studentID); err != nil {
		return nil, err
	}

	assigned, err := s.testRepo.IsStudentAssigned(testID, studentID)
	if err != nil {
		return nil, err
	}
	if !assigned {
		return nil, errs.ErrStudentNotAssigned
	}

	return s.listQuestions(testID)
}

// SubmitAnswer stores or updates a student's answer.
func (s *AssessmentService) SubmitAnswer(ctx context.Context, answer *domain.Answer) (*domain.Answer, error) {
	if answer == nil {
		return nil, errs.ErrInvalidAnswer
	}
	if err := s.ensureStudentExists(answer.StudentID); err != nil {
		return nil, err
	}

	assigned, err := s.testRepo.IsStudentAssigned(answer.TestID, answer.StudentID)
	if err != nil {
		return nil, err
	}
	if !assigned {
		return nil, errs.ErrStudentNotAssigned
	}

	questions, err := s.testRepo.ListQuestions(answer.TestID)
	if err != nil {
		return nil, err
	}

	var found bool
	for _, q := range questions {
		if q.ID == answer.QuestionID {
			found = true
			break
		}
	}
	if !found {
		return nil, errs.ErrQuestionNotFound
	}

	now := time.Now().UTC()
	existing, err := s.answerRepo.GetAnswer(answer.TestID, answer.QuestionID, answer.StudentID)
	if err != nil {
		return nil, err
	}

	if existing != nil {
		answer.ID = existing.ID
		answer.CreatedAt = existing.CreatedAt
		answer.UpdatedAt = now
	} else {
		answer.ID = domain.AnswerID(id.New())
		answer.CreatedAt = now
		answer.UpdatedAt = now
	}

	if err := s.answerRepo.UpsertAnswer(answer); err != nil {
		return nil, err
	}

	return answer, nil
}

// ListResultsForStudent lists grading results for a student's test.
func (s *AssessmentService) ListResultsForStudent(ctx context.Context, studentID domain.StudentID, testID domain.TestID) ([]domain.Result, error) {
	if err := s.ensureStudentExists(studentID); err != nil {
		return nil, err
	}

	assigned, err := s.testRepo.IsStudentAssigned(testID, studentID)
	if err != nil {
		return nil, err
	}
	if !assigned {
		return nil, errs.ErrStudentNotAssigned
	}

	results, err := s.resultRepo.ListResultsByStudent(testID, studentID)
	if err != nil {
		return nil, err
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].CreatedAt.Before(results[j].CreatedAt)
	})

	return results, nil
}

// GradeInput describes grading instructions.
type GradeInput struct {
	TeacherID  domain.TeacherID
	TestID     domain.TestID
	QuestionID domain.QuestionID
	StudentID  domain.StudentID
	Score      int
	Feedback   string
	Completed  bool
}

// GradeAnswer upserts a grading result. Teacher ownership is validated.
func (s *AssessmentService) GradeAnswer(ctx context.Context, input GradeInput) (*domain.Result, error) {
	if err := s.ensureTeacherOwnsTest(input.TeacherID, input.TestID); err != nil {
		return nil, err
	}

	assigned, err := s.testRepo.IsStudentAssigned(input.TestID, input.StudentID)
	if err != nil {
		return nil, err
	}
	if !assigned {
		return nil, errs.ErrStudentNotAssigned
	}

	answer, err := s.answerRepo.GetAnswer(input.TestID, input.QuestionID, input.StudentID)
	if err != nil {
		return nil, err
	}
	if answer == nil {
		return nil, errs.ErrAnswerNotFound
	}

	now := time.Now().UTC()
	existing, err := s.resultRepo.GetResult(answer.ID)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		existing.Score = input.Score
		existing.Feedback = input.Feedback
		existing.Completed = input.Completed
		existing.UpdatedAt = now
		if err := s.resultRepo.SaveResult(existing); err != nil {
			return nil, err
		}
		return existing, nil
	}

	result := &domain.Result{
		ID:        domain.ResultID(id.New()),
		AnswerID:  answer.ID,
		Score:     input.Score,
		Feedback:  input.Feedback,
		Completed: input.Completed,
		CreatedAt: now,
		UpdatedAt: now,
	}

	if err := s.resultRepo.SaveResult(result); err != nil {
		return nil, err
	}

	return result, nil
}

// Helpers.

func (s *AssessmentService) ensureTeacherExists(teacherID domain.TeacherID) error {
	teacher, err := s.orgRepo.GetTeacher(teacherID)
	if err != nil {
		return err
	}
	if teacher == nil {
		return errs.ErrTeacherNotFound
	}
	return nil
}

func (s *AssessmentService) ensureStudentExists(studentID domain.StudentID) error {
	student, err := s.orgRepo.GetStudent(studentID)
	if err != nil {
		return err
	}
	if student == nil {
		return errs.ErrStudentNotFound
	}
	return nil
}

func (s *AssessmentService) ensureTeacherOwnsTest(teacherID domain.TeacherID, testID domain.TestID) error {
	test, err := s.testRepo.GetTest(testID)
	if err != nil {
		return err
	}
	if test == nil {
		return errs.ErrTestNotFound
	}
	if test.TeacherID != teacherID {
		return errs.ErrForbiddenTeacher
	}
	return nil
}

func (s *AssessmentService) listQuestions(testID domain.TestID) ([]domain.Question, error) {
	questions, err := s.testRepo.ListQuestions(testID)
	if err != nil {
		return nil, err
	}

	sort.Slice(questions, func(i, j int) bool {
		return questions[i].Sequence < questions[j].Sequence
	})

	return questions, nil
}
