package domain

import "time"

// Identifier wrappers for stronger typing.
type (
	SchoolID   string
	GradeID    string
	ClassID    string
	TeacherID  string
	StudentID  string
	TestID     string
	QuestionID string
	AnswerID   string
	ResultID   string
)

// School groups grades, classes, teachers, and tests.
type School struct {
	ID        SchoolID
	Name      string
	CreatedAt time.Time
}

// Grade belongs to a school and groups classes.
type Grade struct {
	ID        GradeID
	SchoolID  SchoolID
	Name      string
	CreatedAt time.Time
}

// Class belongs to a grade and groups students.
type Class struct {
	ID        ClassID
	GradeID   GradeID
	Name      string
	CreatedAt time.Time
}

// Teacher teaches within a school.
type Teacher struct {
	ID        TeacherID
	SchoolID  SchoolID
	Name      string
	Email     string
	CreatedAt time.Time
}

// Student belongs to a class and takes tests.
type Student struct {
	ID        StudentID
	ClassID   ClassID
	Name      string
	Email     string
	CreatedAt time.Time
}

// Test authored by a teacher and assigned to students.
type Test struct {
	ID         TestID
	TeacherID  TeacherID
	Title      string
	Published  bool
	CreatedAt  time.Time
	UpdatedAt  time.Time
	AssignedTo []StudentID
}

// Question represents a test question.
type Question struct {
	ID        QuestionID
	TestID    TestID
	Sequence  int
	Prompt    string
	Points    int
	CreatedAt time.Time
}

// Answer submitted by a student for a question.
type Answer struct {
	ID         AnswerID
	TestID     TestID
	QuestionID QuestionID
	StudentID  StudentID
	Response   string
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

// Result represents grading feedback for an answer.
type Result struct {
	ID        ResultID
	AnswerID  AnswerID
	Score     int
	Feedback  string
	Completed bool
	CreatedAt time.Time
	UpdatedAt time.Time
}
