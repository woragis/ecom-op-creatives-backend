package service

import (
	"context"
	"strings"

	"github.com/google/uuid"
	"github.com/woragis/ecom-op-creatives-backend/server/internal/apperrors"
	"github.com/woragis/ecom-op-creatives-backend/server/internal/models"
	productrepo "github.com/woragis/ecom-op-creatives-backend/server/internal/product/repository"
)

type Service struct {
	repo *productrepo.Repository
}

func New(repo *productrepo.Repository) *Service {
	return &Service{repo: repo}
}

type CreateInput struct {
	Name  string
	URL   *string
	Niche *string
}

func (s *Service) Create(ctx context.Context, in CreateInput) (*models.Product, error) {
	name := strings.TrimSpace(in.Name)
	if name == "" {
		return nil, apperrors.Invalid(apperrors.CodeProductCreateInvalidBody, apperrors.MsgProductCreateInvalidBody)
	}
	product := &models.Product{
		Name:  name,
		URL:   in.URL,
		Niche: in.Niche,
	}
	if err := s.repo.Create(ctx, product); err != nil {
		return nil, apperrors.Wrapf(err, "product create")
	}
	return product, nil
}

func (s *Service) List(ctx context.Context) ([]models.Product, error) {
	items, err := s.repo.List(ctx)
	if err != nil {
		return nil, apperrors.Wrapf(err, "product list")
	}
	return items, nil
}

func (s *Service) GetByID(ctx context.Context, id uuid.UUID) (*models.Product, error) {
	return s.repo.GetByID(ctx, id)
}
