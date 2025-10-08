package errs

import "errors"

var (
	ErrTeacherNotFound    = errors.New("teacher not found")
	ErrStudentNotFound    = errors.New("student not found")
	ErrSchoolNotFound     = errors.New("school not found")
	ErrGradeNotFound      = errors.New("grade not found")
	ErrClassNotFound      = errors.New("class not found")
	ErrTestNotFound       = errors.New("test not found")
	ErrQuestionNotFound   = errors.New("question not found")
	ErrAnswerNotFound     = errors.New("answer not found")
	ErrResultNotFound     = errors.New("result not found")
	ErrStudentNotAssigned = errors.New("student not assigned to test")
	ErrForbiddenTeacher   = errors.New("teacher cannot access this resource")
	ErrInvalidTest        = errors.New("invalid test payload")
	ErrInvalidQuestion    = errors.New("invalid question payload")
	ErrInvalidAnswer      = errors.New("invalid answer payload")
	ErrNoQuestions        = errors.New("no questions provided")
)
