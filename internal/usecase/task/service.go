package task

import (
	"context"
	"fmt"
	"strings"
	"time"

	taskdomain "example.com/taskservice/internal/domain/task"
)

type Service struct {
	repo Repository
	now  func() time.Time
}

func NewService(repo Repository) *Service {
	return &Service{
		repo: repo,
		now:  func() time.Time { return time.Now().UTC() },
	}
}

func (s *Service) Create(ctx context.Context, input CreateInput) (*taskdomain.Task, error) {
	normalized, err := validateCreateInput(input)
	if err != nil {
		return nil, err
	}

	model := &taskdomain.Task{
		Title:       normalized.Title,
		Description: normalized.Description,
		Status:      normalized.Status,
		Type:        normalized.Type,
		Interval:    normalized.Interval,
		ScheduleAt:  normalized.ScheduleAt,
	}
	now := s.now()
	model.CreatedAt = now
	model.UpdatedAt = now

	created, err := s.repo.Create(ctx, model)
	if err != nil {
		return nil, err
	}

	return created, nil
}

func (s *Service) GetByID(ctx context.Context, id int64) (*taskdomain.Task, error) {
	if id <= 0 {
		return nil, fmt.Errorf("%w: id must be positive", ErrInvalidInput)
	}

	return s.repo.GetByID(ctx, id)
}

func (s *Service) Update(ctx context.Context, id int64, input UpdateInput) (*taskdomain.Task, error) {
	if id <= 0 {
		return nil, fmt.Errorf("%w: id must be positive", ErrInvalidInput)
	}

	normalized, err := validateUpdateInput(input)
	if err != nil {
		return nil, err
	}

	model := &taskdomain.Task{
		ID:          id,
		Title:       normalized.Title,
		Description: normalized.Description,
		Status:      normalized.Status,
		Type:        normalized.Type,
		Interval:    normalized.Interval,
		ScheduleAt:  normalized.ScheduleAt,
		UpdatedAt:   s.now(),
	}

	updated, err := s.repo.Update(ctx, model)
	if err != nil {
		return nil, err
	}

	return updated, nil
}

func (s *Service) Delete(ctx context.Context, id int64) error {
	if id <= 0 {
		return fmt.Errorf("%w: id must be positive", ErrInvalidInput)
	}

	return s.repo.Delete(ctx, id)
}

func (s *Service) List(ctx context.Context) ([]taskdomain.Task, error) {
	return s.repo.List(ctx)
}

func (s *Service) GetByType(ctx context.Context, t taskdomain.Type) ([]taskdomain.Task, error) {
	if !t.Valid() {
		return []taskdomain.Task{}, fmt.Errorf("%w: invalid status", ErrInvalidInput)
	}
	return s.repo.GetByType(ctx, t)
}

func validateCreateInput(input CreateInput) (CreateInput, error) {
	input.Title = strings.TrimSpace(input.Title)
	input.Description = strings.TrimSpace(input.Description)

	if input.Title == "" {
		return CreateInput{}, fmt.Errorf("%w: title is required", ErrInvalidInput)
	}

	if input.Status == "" {
		input.Status = taskdomain.StatusNew
	}

	if !input.Status.Valid() {
		return CreateInput{}, fmt.Errorf("%w: invalid status", ErrInvalidInput)
	}

	if !input.Type.Valid() {
		return CreateInput{}, fmt.Errorf("%w: invalid interval type", ErrInvalidInput)
	}

	if !validateTaskInterval(input.Interval, input.Type) {
		return CreateInput{}, fmt.Errorf("%w: invalid task interval", ErrInvalidInput)
	}

	if !validateScheduleAt(input.ScheduleAt) {
		return CreateInput{}, fmt.Errorf("%w: invalid schedule at", ErrInvalidInput)
	}

	return input, nil
}

func validateUpdateInput(input UpdateInput) (UpdateInput, error) {
	input.Title = strings.TrimSpace(input.Title)
	input.Description = strings.TrimSpace(input.Description)

	if input.Title == "" {
		return UpdateInput{}, fmt.Errorf("%w: title is required", ErrInvalidInput)
	}

	if !input.Status.Valid() {
		return UpdateInput{}, fmt.Errorf("%w: invalid status", ErrInvalidInput)
	}

	if !input.Type.Valid() {
		return UpdateInput{}, fmt.Errorf("%w: invalid interval type", ErrInvalidInput)
	}

	if !validateTaskInterval(input.Interval, input.Type) {
		return UpdateInput{}, fmt.Errorf("%w: invalid task interval", ErrInvalidInput)
	}

	if !validateScheduleAt(input.ScheduleAt) {
		return UpdateInput{}, fmt.Errorf("%w: invalid schedule at", ErrInvalidInput)
	}

	return input, nil
}

func validateTaskInterval(i int64, t taskdomain.Type) bool {

	switch t {
	case taskdomain.Daily:
		if i > 0 {
			return true
		}
		return false
	case taskdomain.Monthly:
		if i > 0 && i < 31 {
			return true
		}
		return false
	case taskdomain.SpecialDate:
		return true
	case taskdomain.EvenOdd:
		if i >= 1 && i <= 2 {
			return true
		}
		return false
	default:
		return false

	}
}
func validateScheduleAt(s time.Time) bool {
	if s.IsZero() {
		return false
	}
	nowTime := time.Now()
	if s.Before(nowTime) {
		return false
	}
	return true
}
