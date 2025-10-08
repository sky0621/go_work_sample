package memory

import (
	"errors"
	"sort"
	"sync"
	"time"

	"github.com/sky0621/go_work_sample/core/pkg/domain"
	"github.com/sky0621/go_work_sample/core/pkg/repository"
)

// SeedData bootstraps the in-memory store with initial organization data.
type SeedData struct {
	Schools  []domain.School
	Grades   []domain.Grade
	Classes  []domain.Class
	Teachers []domain.Teacher
	Students []domain.Student
}

// Repository implements all repository interfaces in-memory.
type Repository struct {
	mu sync.RWMutex

	schools  map[domain.SchoolID]domain.School
	grades   map[domain.GradeID]domain.Grade
	classes  map[domain.ClassID]domain.Class
	teachers map[domain.TeacherID]domain.Teacher
	students map[domain.StudentID]domain.Student

	tests          map[domain.TestID]domain.Test
	questions      map[domain.QuestionID]domain.Question
	testQuestions  map[domain.TestID][]domain.QuestionID
	assignments    map[domain.TestID]map[domain.StudentID]struct{}
	studentTests   map[domain.StudentID]map[domain.TestID]struct{}
	answers        map[domain.AnswerID]domain.Answer
	answerIndex    map[string]domain.AnswerID
	answersByTest  map[domain.TestID]map[domain.AnswerID]struct{}
	results        map[domain.ResultID]domain.Result
	resultByAnswer map[domain.AnswerID]domain.ResultID
}

// NewRepository creates a repository loaded with the provided seed.
func NewRepository(seed SeedData) *Repository {
	repo := &Repository{
		schools:        make(map[domain.SchoolID]domain.School),
		grades:         make(map[domain.GradeID]domain.Grade),
		classes:        make(map[domain.ClassID]domain.Class),
		teachers:       make(map[domain.TeacherID]domain.Teacher),
		students:       make(map[domain.StudentID]domain.Student),
		tests:          make(map[domain.TestID]domain.Test),
		questions:      make(map[domain.QuestionID]domain.Question),
		testQuestions:  make(map[domain.TestID][]domain.QuestionID),
		assignments:    make(map[domain.TestID]map[domain.StudentID]struct{}),
		studentTests:   make(map[domain.StudentID]map[domain.TestID]struct{}),
		answers:        make(map[domain.AnswerID]domain.Answer),
		answerIndex:    make(map[string]domain.AnswerID),
		answersByTest:  make(map[domain.TestID]map[domain.AnswerID]struct{}),
		results:        make(map[domain.ResultID]domain.Result),
		resultByAnswer: make(map[domain.AnswerID]domain.ResultID),
	}

	for _, s := range seed.Schools {
		repo.schools[s.ID] = cloneSchool(s)
	}
	for _, g := range seed.Grades {
		repo.grades[g.ID] = cloneGrade(g)
	}
	for _, c := range seed.Classes {
		repo.classes[c.ID] = cloneClass(c)
	}
	for _, t := range seed.Teachers {
		repo.teachers[t.ID] = cloneTeacher(t)
	}
	for _, st := range seed.Students {
		repo.students[st.ID] = cloneStudent(st)
	}

	return repo
}

var _ repository.OrganizationRepository = (*Repository)(nil)
var _ repository.TestRepository = (*Repository)(nil)
var _ repository.AnswerRepository = (*Repository)(nil)
var _ repository.ResultRepository = (*Repository)(nil)

// OrganizationRepository implementation.

func (r *Repository) GetSchool(id domain.SchoolID) (*domain.School, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	school, ok := r.schools[id]
	if !ok {
		return nil, nil
	}
	s := cloneSchool(school)
	return &s, nil
}

func (r *Repository) ListSchools() ([]domain.School, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	schools := make([]domain.School, 0, len(r.schools))
	for _, school := range r.schools {
		schools = append(schools, cloneSchool(school))
	}

	sort.Slice(schools, func(i, j int) bool {
		return schools[i].CreatedAt.Before(schools[j].CreatedAt)
	})

	return schools, nil
}

func (r *Repository) GetGrade(id domain.GradeID) (*domain.Grade, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	grade, ok := r.grades[id]
	if !ok {
		return nil, nil
	}
	g := cloneGrade(grade)
	return &g, nil
}

func (r *Repository) GetClass(id domain.ClassID) (*domain.Class, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	class, ok := r.classes[id]
	if !ok {
		return nil, nil
	}
	c := cloneClass(class)
	return &c, nil
}

func (r *Repository) GetTeacher(id domain.TeacherID) (*domain.Teacher, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	teacher, ok := r.teachers[id]
	if !ok {
		return nil, nil
	}
	t := cloneTeacher(teacher)
	return &t, nil
}

func (r *Repository) GetStudent(id domain.StudentID) (*domain.Student, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	student, ok := r.students[id]
	if !ok {
		return nil, nil
	}
	s := cloneStudent(student)
	return &s, nil
}

func (r *Repository) ListGrades(schoolID domain.SchoolID) ([]domain.Grade, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	grades := make([]domain.Grade, 0)
	for _, g := range r.grades {
		if g.SchoolID == schoolID {
			grades = append(grades, cloneGrade(g))
		}
	}

	sort.Slice(grades, func(i, j int) bool {
		return grades[i].CreatedAt.Before(grades[j].CreatedAt)
	})

	return grades, nil
}

func (r *Repository) ListClasses(gradeID domain.GradeID) ([]domain.Class, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	classes := make([]domain.Class, 0)
	for _, c := range r.classes {
		if c.GradeID == gradeID {
			classes = append(classes, cloneClass(c))
		}
	}

	sort.Slice(classes, func(i, j int) bool {
		return classes[i].CreatedAt.Before(classes[j].CreatedAt)
	})

	return classes, nil
}

func (r *Repository) ListStudents(classID domain.ClassID) ([]domain.Student, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	students := make([]domain.Student, 0)
	for _, st := range r.students {
		if st.ClassID == classID {
			students = append(students, cloneStudent(st))
		}
	}

	sort.Slice(students, func(i, j int) bool {
		return students[i].CreatedAt.Before(students[j].CreatedAt)
	})

	return students, nil
}

func (r *Repository) ListTeachers(schoolID domain.SchoolID) ([]domain.Teacher, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	teachers := make([]domain.Teacher, 0)
	for _, t := range r.teachers {
		if t.SchoolID == schoolID {
			teachers = append(teachers, cloneTeacher(t))
		}
	}

	sort.Slice(teachers, func(i, j int) bool {
		return teachers[i].CreatedAt.Before(teachers[j].CreatedAt)
	})

	return teachers, nil
}

// TestRepository implementation.

func (r *Repository) CreateTest(test *domain.Test, questions []domain.Question, studentIDs []domain.StudentID) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.tests[test.ID]; exists {
		return errors.New("test already exists")
	}

	if _, ok := r.teachers[test.TeacherID]; !ok {
		return errors.New("teacher not found")
	}

	for _, studentID := range studentIDs {
		if _, ok := r.students[studentID]; !ok {
			return errors.New("student not found")
		}
	}

	clone := cloneTest(*test)
	clone.AssignedTo = append([]domain.StudentID(nil), studentIDs...)
	r.tests[test.ID] = clone

	questionIDs := make([]domain.QuestionID, len(questions))
	for i, q := range questions {
		questionIDs[i] = q.ID
		r.questions[q.ID] = cloneQuestion(q)
	}
	r.testQuestions[test.ID] = questionIDs

	if _, ok := r.assignments[test.ID]; !ok {
		r.assignments[test.ID] = make(map[domain.StudentID]struct{})
	}
	for _, studentID := range studentIDs {
		r.assignments[test.ID][studentID] = struct{}{}
		if _, ok := r.studentTests[studentID]; !ok {
			r.studentTests[studentID] = make(map[domain.TestID]struct{})
		}
		r.studentTests[studentID][test.ID] = struct{}{}
	}

	return nil
}

func (r *Repository) UpdateTest(test *domain.Test) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.tests[test.ID]; !ok {
		return errors.New("test not found")
	}
	clone := cloneTest(*test)
	r.tests[test.ID] = clone
	return nil
}

func (r *Repository) GetTest(id domain.TestID) (*domain.Test, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	test, ok := r.tests[id]
	if !ok {
		return nil, nil
	}
	t := cloneTest(test)
	return &t, nil
}

func (r *Repository) ListTestsByTeacher(teacherID domain.TeacherID) ([]domain.Test, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	tests := make([]domain.Test, 0)
	for _, test := range r.tests {
		if test.TeacherID == teacherID {
			tests = append(tests, cloneTest(test))
		}
	}

	sort.Slice(tests, func(i, j int) bool {
		return tests[i].CreatedAt.Before(tests[j].CreatedAt)
	})

	return tests, nil
}

func (r *Repository) ListTestsForStudent(studentID domain.StudentID) ([]domain.Test, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	testRefs, ok := r.studentTests[studentID]
	if !ok {
		return []domain.Test{}, nil
	}

	tests := make([]domain.Test, 0, len(testRefs))
	for testID := range testRefs {
		if test, ok := r.tests[testID]; ok {
			tests = append(tests, cloneTest(test))
		}
	}

	sort.Slice(tests, func(i, j int) bool {
		return tests[i].CreatedAt.Before(tests[j].CreatedAt)
	})

	return tests, nil
}

func (r *Repository) ListQuestions(testID domain.TestID) ([]domain.Question, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	ids, ok := r.testQuestions[testID]
	if !ok {
		return []domain.Question{}, nil
	}

	questions := make([]domain.Question, 0, len(ids))
	for _, id := range ids {
		if q, ok := r.questions[id]; ok {
			questions = append(questions, cloneQuestion(q))
		}
	}

	sort.Slice(questions, func(i, j int) bool {
		return questions[i].Sequence < questions[j].Sequence
	})

	return questions, nil
}

func (r *Repository) IsStudentAssigned(testID domain.TestID, studentID domain.StudentID) (bool, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	students, ok := r.assignments[testID]
	if !ok {
		return false, nil
	}

	_, assigned := students[studentID]
	return assigned, nil
}

// AnswerRepository implementation.

func (r *Repository) UpsertAnswer(answer *domain.Answer) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	key := answerKey(answer.TestID, answer.QuestionID, answer.StudentID)
	r.answers[answer.ID] = cloneAnswer(*answer)
	r.answerIndex[key] = answer.ID

	if _, ok := r.answersByTest[answer.TestID]; !ok {
		r.answersByTest[answer.TestID] = make(map[domain.AnswerID]struct{})
	}
	r.answersByTest[answer.TestID][answer.ID] = struct{}{}

	return nil
}

func (r *Repository) GetAnswer(testID domain.TestID, questionID domain.QuestionID, studentID domain.StudentID) (*domain.Answer, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	key := answerKey(testID, questionID, studentID)
	ansID, ok := r.answerIndex[key]
	if !ok {
		return nil, nil
	}

	ans, ok := r.answers[ansID]
	if !ok {
		return nil, nil
	}
	cloned := cloneAnswer(ans)
	return &cloned, nil
}

func (r *Repository) ListAnswers(testID domain.TestID, studentID domain.StudentID) ([]domain.Answer, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	ids, ok := r.answersByTest[testID]
	if !ok {
		return []domain.Answer{}, nil
	}

	answers := make([]domain.Answer, 0)
	for id := range ids {
		if ans, ok := r.answers[id]; ok && ans.StudentID == studentID {
			answers = append(answers, cloneAnswer(ans))
		}
	}

	sort.Slice(answers, func(i, j int) bool {
		return answers[i].CreatedAt.Before(answers[j].CreatedAt)
	})

	return answers, nil
}

func (r *Repository) ListAnswersByTest(testID domain.TestID) ([]domain.Answer, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	ids, ok := r.answersByTest[testID]
	if !ok {
		return []domain.Answer{}, nil
	}

	answers := make([]domain.Answer, 0, len(ids))
	for id := range ids {
		if ans, ok := r.answers[id]; ok {
			answers = append(answers, cloneAnswer(ans))
		}
	}

	sort.Slice(answers, func(i, j int) bool {
		return answers[i].CreatedAt.Before(answers[j].CreatedAt)
	})

	return answers, nil
}

// ResultRepository implementation.

func (r *Repository) SaveResult(result *domain.Result) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.results[result.ID] = cloneResult(*result)
	r.resultByAnswer[result.AnswerID] = result.ID

	return nil
}

func (r *Repository) GetResult(answerID domain.AnswerID) (*domain.Result, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	resultID, ok := r.resultByAnswer[answerID]
	if !ok {
		return nil, nil
	}

	res, ok := r.results[resultID]
	if !ok {
		return nil, nil
	}
	cloned := cloneResult(res)
	return &cloned, nil
}

func (r *Repository) ListResultsByTest(testID domain.TestID) ([]domain.Result, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	answerIDs, ok := r.answersByTest[testID]
	if !ok {
		return []domain.Result{}, nil
	}

	results := make([]domain.Result, 0)
	for answerID := range answerIDs {
		if resultID, ok := r.resultByAnswer[answerID]; ok {
			if res, ok := r.results[resultID]; ok {
				results = append(results, cloneResult(res))
			}
		}
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].CreatedAt.Before(results[j].CreatedAt)
	})

	return results, nil
}

func (r *Repository) ListResultsByStudent(testID domain.TestID, studentID domain.StudentID) ([]domain.Result, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	answerIDs, ok := r.answersByTest[testID]
	if !ok {
		return []domain.Result{}, nil
	}

	results := make([]domain.Result, 0)
	for answerID := range answerIDs {
		ans, ok := r.answers[answerID]
		if !ok || ans.StudentID != studentID {
			continue
		}
		if resultID, ok := r.resultByAnswer[answerID]; ok {
			if res, ok := r.results[resultID]; ok {
				results = append(results, cloneResult(res))
			}
		}
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].CreatedAt.Before(results[j].CreatedAt)
	})

	return results, nil
}

// Helpers.

func answerKey(testID domain.TestID, questionID domain.QuestionID, studentID domain.StudentID) string {
	return string(testID) + "|" + string(questionID) + "|" + string(studentID)
}

func cloneSchool(in domain.School) domain.School    { return in }
func cloneGrade(in domain.Grade) domain.Grade       { return in }
func cloneClass(in domain.Class) domain.Class       { return in }
func cloneTeacher(in domain.Teacher) domain.Teacher { return in }
func cloneStudent(in domain.Student) domain.Student { return in }

func cloneTest(in domain.Test) domain.Test {
	clone := in
	clone.AssignedTo = append([]domain.StudentID(nil), in.AssignedTo...)
	return clone
}

func cloneQuestion(in domain.Question) domain.Question { return in }
func cloneAnswer(in domain.Answer) domain.Answer       { return in }
func cloneResult(in domain.Result) domain.Result       { return in }

// SampleSeed provides deterministic data for demos.
func SampleSeed() SeedData {
	now := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)

	schoolID := domain.SchoolID("school-001")
	gradeID := domain.GradeID("grade-001")
	classA := domain.ClassID("class-1A")
	classB := domain.ClassID("class-1B")
	teacherID := domain.TeacherID("teacher-001")

	return SeedData{
		Schools: []domain.School{{ID: schoolID, Name: "Example High School", CreatedAt: now}},
		Grades:  []domain.Grade{{ID: gradeID, SchoolID: schoolID, Name: "1st Grade", CreatedAt: now}},
		Classes: []domain.Class{
			{ID: classA, GradeID: gradeID, Name: "Class A", CreatedAt: now},
			{ID: classB, GradeID: gradeID, Name: "Class B", CreatedAt: now},
		},
		Teachers: []domain.Teacher{{ID: teacherID, SchoolID: schoolID, Name: "Mrs. Smith", Email: "smith@example.com", CreatedAt: now}},
		Students: []domain.Student{
			{ID: domain.StudentID("student-001"), ClassID: classA, Name: "Alice", Email: "alice@example.com", CreatedAt: now},
			{ID: domain.StudentID("student-002"), ClassID: classA, Name: "Bob", Email: "bob@example.com", CreatedAt: now.Add(time.Minute)},
			{ID: domain.StudentID("student-003"), ClassID: classB, Name: "Charlie", Email: "charlie@example.com", CreatedAt: now.Add(2 * time.Minute)},
		},
	}
}
