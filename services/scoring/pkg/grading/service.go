package grading

import (
	"context"

	"github.com/sky0621/go_work_sample/core/pkg/domain"
	"github.com/sky0621/go_work_sample/core/pkg/usecase"
)

// Service wraps assessment use cases to expose grading specific APIs.
type Service struct {
	assessments *usecase.AssessmentService
}

// NewService creates a grading service instance.
func NewService(assessments *usecase.AssessmentService) *Service {
	return &Service{assessments: assessments}
}

// GradeAnswer delegates to the underlying assessment logic.
func (s *Service) GradeAnswer(ctx context.Context, teacherID domain.TeacherID, payload usecase.GradeInput) (*domain.Result, error) {
	payload.TeacherID = teacherID
	return s.assessments.GradeAnswer(ctx, payload)
}
