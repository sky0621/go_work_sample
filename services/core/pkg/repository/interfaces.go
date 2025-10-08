package repository

import "github.com/sky0621/go_work_sample/core/pkg/domain"

// OrganizationRepository exposes hierarchy data access.
type OrganizationRepository interface {
	ListSchools() ([]domain.School, error)
	GetSchool(id domain.SchoolID) (*domain.School, error)
	GetGrade(id domain.GradeID) (*domain.Grade, error)
	GetClass(id domain.ClassID) (*domain.Class, error)
	GetTeacher(id domain.TeacherID) (*domain.Teacher, error)
	GetStudent(id domain.StudentID) (*domain.Student, error)

	ListGrades(schoolID domain.SchoolID) ([]domain.Grade, error)
	ListClasses(gradeID domain.GradeID) ([]domain.Class, error)
	ListStudents(classID domain.ClassID) ([]domain.Student, error)
	ListTeachers(schoolID domain.SchoolID) ([]domain.Teacher, error)
}

// TestRepository manages tests and questions.
type TestRepository interface {
	CreateTest(test *domain.Test, questions []domain.Question, studentIDs []domain.StudentID) error
	UpdateTest(test *domain.Test) error
	GetTest(id domain.TestID) (*domain.Test, error)
	ListTestsByTeacher(teacherID domain.TeacherID) ([]domain.Test, error)
	ListTestsForStudent(studentID domain.StudentID) ([]domain.Test, error)
	ListQuestions(testID domain.TestID) ([]domain.Question, error)
	IsStudentAssigned(testID domain.TestID, studentID domain.StudentID) (bool, error)
}

// AnswerRepository persists student answers.
type AnswerRepository interface {
	UpsertAnswer(answer *domain.Answer) error
	GetAnswer(testID domain.TestID, questionID domain.QuestionID, studentID domain.StudentID) (*domain.Answer, error)
	ListAnswers(testID domain.TestID, studentID domain.StudentID) ([]domain.Answer, error)
	ListAnswersByTest(testID domain.TestID) ([]domain.Answer, error)
}

// ResultRepository persists grading results.
type ResultRepository interface {
	SaveResult(result *domain.Result) error
	GetResult(answerID domain.AnswerID) (*domain.Result, error)
	ListResultsByTest(testID domain.TestID) ([]domain.Result, error)
	ListResultsByStudent(testID domain.TestID, studentID domain.StudentID) ([]domain.Result, error)
}
