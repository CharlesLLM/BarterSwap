package application

import (
	"context"
	"strings"

	"github.com/CharlesLLM/BarterSwap/internal/domain"
)

type ReviewRepository interface {
	FindExchange(context.Context, int) (domain.Exchange, error)
	FindUser(context.Context, int) (domain.User, error)
	FindService(context.Context, int) (domain.Service, error)
	CreateReview(context.Context, int, int, int, domain.CreateReviewInput) (domain.Review, error)
	ListUserReviews(context.Context, int) ([]domain.Review, error)
	ListServiceReviews(context.Context, int) ([]domain.Review, error)
}

type ReviewService struct {
	repository ReviewRepository
}

func NewReviewService(repository ReviewRepository) ReviewService {
	return ReviewService{repository: repository}
}

func (service ReviewService) Create(ctx context.Context, exchangeID, authorID int, input domain.CreateReviewInput) (domain.Review, error) {
	if input.Note < 1 || input.Note > 5 {
		return domain.Review{}, domain.ErrReviewNoteInvalid
	}
	input.Commentaire = strings.TrimSpace(input.Commentaire)

	exchange, err := service.repository.FindExchange(ctx, exchangeID)
	if err != nil {
		return domain.Review{}, err
	}
	if exchange.Status != domain.ExchangeStatusCompleted {
		return domain.Review{}, domain.ErrReviewExchangeIncomplete
	}

	targetID := exchange.RequesterID
	if authorID == exchange.RequesterID {
		targetID = exchange.OwnerID
	} else if authorID != exchange.OwnerID {
		return domain.Review{}, domain.ErrReviewForbidden
	}

	return service.repository.CreateReview(ctx, exchangeID, authorID, targetID, input)
}

func (service ReviewService) ListForUser(ctx context.Context, userID int) ([]domain.Review, error) {
	if _, err := service.repository.FindUser(ctx, userID); err != nil {
		return nil, err
	}
	return service.repository.ListUserReviews(ctx, userID)
}

func (service ReviewService) ListForService(ctx context.Context, serviceID int) ([]domain.Review, error) {
	if _, err := service.repository.FindService(ctx, serviceID); err != nil {
		return nil, err
	}
	return service.repository.ListServiceReviews(ctx, serviceID)
}
