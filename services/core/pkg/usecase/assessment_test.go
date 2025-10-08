package usecase_test

import (
	"context"
	"testing"

	"github.com/sky0621/go_work_sample/core/pkg/domain"
	"github.com/sky0621/go_work_sample/core/pkg/memory"
	"github.com/sky0621/go_work_sample/core/pkg/usecase"
)

func TestAssessmentService_Workflow(t *testing.T) {
	repo := memory.NewRepository(memory.SampleSeed())
	service := usecase.NewAssessmentService(repo, repo, repo, repo)

	teacherID := domain.TeacherID("teacher-001")
	studentIDs := []domain.StudentID{"student-001", "student-002"}

	test, questions, err := service.CreateTest(context.Background(), usecase.CreateTestInput{
		Title:      "Math Quiz",
		TeacherID:  teacherID,
		Questions:  []usecase.QuestionDraft{{Prompt: "1+1?", Points: 5}},
		StudentIDs: studentIDs,
	})
	if err != nil {
		t.Fatalf("CreateTest failed: %v", err)
	}
	if len(questions) != 1 {
		t.Fatalf("expected one question, got %d", len(questions))
	}

	tests, err := service.ListTestsByTeacher(context.Background(), teacherID)
	if err != nil {
		t.Fatalf("ListTestsByTeacher failed: %v", err)
	}
	if len(tests) != 1 {
		t.Fatalf("expected one test, got %d", len(tests))
	}

	studentTests, err := service.ListTestsForStudent(context.Background(), studentIDs[0])
	if err != nil {
		t.Fatalf("ListTestsForStudent failed: %v", err)
	}
	if len(studentTests) != 1 {
		t.Fatalf("expected student to have one test, got %d", len(studentTests))
	}

	answer := &domain.Answer{
		TestID:     test.ID,
		QuestionID: questions[0].ID,
		StudentID:  studentIDs[0],
		Response:   "2",
	}
	savedAnswer, err := service.SubmitAnswer(context.Background(), answer)
	if err != nil {
		t.Fatalf("SubmitAnswer failed: %v", err)
	}
	if savedAnswer.ID == "" {
		t.Fatal("expected answer to have ID")
	}

	result, err := service.GradeAnswer(context.Background(), usecase.GradeInput{
		TeacherID:  teacherID,
		TestID:     test.ID,
		QuestionID: questions[0].ID,
		StudentID:  studentIDs[0],
		Score:      5,
		Feedback:   "good",
		Completed:  true,
	})
	if err != nil {
		t.Fatalf("GradeAnswer failed: %v", err)
	}
	if result.Score != 5 {
		t.Fatalf("expected score 5, got %d", result.Score)
	}

	results, err := service.ListResultsForStudent(context.Background(), studentIDs[0], test.ID)
	if err != nil {
		t.Fatalf("ListResultsForStudent failed: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("expected one result, got %d", len(results))
	}
}
