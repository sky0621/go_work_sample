package filedb_test

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/sky0621/go_work_sample/core/pkg/domain"
	"github.com/sky0621/go_work_sample/core/pkg/memory"
	"github.com/sky0621/go_work_sample/core/pkg/storage/filedb"
)

func TestRepositoryPersistence(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "state.json")

	repo, err := filedb.NewRepository(path, memory.SampleSeed())
	if err != nil {
		t.Fatalf("NewRepository failed: %v", err)
	}

	test := &domain.Test{
		ID:        domain.TestID("test-001"),
		TeacherID: domain.TeacherID("teacher-001"),
		Title:     "History Quiz",
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}
	questions := []domain.Question{{
		ID:        domain.QuestionID("question-001"),
		TestID:    test.ID,
		Sequence:  1,
		Prompt:    "Who?",
		Points:    10,
		CreatedAt: time.Now().UTC(),
	}}

	if err := repo.CreateTest(test, questions, []domain.StudentID{"student-001"}); err != nil {
		t.Fatalf("CreateTest failed: %v", err)
	}

	repo2, err := filedb.NewRepository(path, memory.SampleSeed())
	if err != nil {
		t.Fatalf("reloading repository failed: %v", err)
	}

	loaded, err := repo2.GetTest(test.ID)
	if err != nil {
		t.Fatalf("GetTest failed: %v", err)
	}
	if loaded == nil || loaded.ID != test.ID {
		t.Fatalf("expected test to persist, got %+v", loaded)
	}
}
