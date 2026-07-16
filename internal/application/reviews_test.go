package application

import (
	"context"
	"errors"
	"testing"

	"github.com/CharlesLLM/BarterSwap/internal/domain"
)

type reviewRepositoryStub struct {
	exchange  domain.Exchange
	createErr error
}

func (repository reviewRepositoryStub) FindExchange(context.Context, int) (domain.Exchange, error) {
	return repository.exchange, nil
}

func (reviewRepositoryStub) FindUser(context.Context, int) (domain.User, error) {
	return domain.User{ID: 1}, nil
}

func (reviewRepositoryStub) FindService(context.Context, int) (domain.Service, error) {
	return domain.Service{ID: 1}, nil
}

func (repository reviewRepositoryStub) CreateReview(
	_ context.Context,
	exchangeID int,
	authorID int,
	targetID int,
	input domain.CreateReviewInput,
) (domain.Review, error) {
	return domain.Review{
		ID: 1, ExchangeID: exchangeID, AuthorID: authorID,
		TargetID: targetID, Note: input.Note, Commentaire: input.Commentaire,
	}, repository.createErr
}

func (reviewRepositoryStub) ListUserReviews(context.Context, int) ([]domain.Review, error) {
	return []domain.Review{}, nil
}

func (reviewRepositoryStub) ListServiceReviews(context.Context, int) ([]domain.Review, error) {
	return []domain.Review{}, nil
}

func TestReviewCreate(testContext *testing.T) {
	tests := []struct {
		name       string
		status     string
		authorID   int
		note       int
		wantTarget int
		wantErr    error
	}{
		{name: "demandeur évalue l'offreur", status: domain.ExchangeStatusCompleted, authorID: 1, note: 5, wantTarget: 2},
		{name: "offreur évalue le demandeur", status: domain.ExchangeStatusCompleted, authorID: 2, note: 4, wantTarget: 1},
		{name: "note trop basse", status: domain.ExchangeStatusCompleted, authorID: 1, note: 0, wantErr: domain.ErrReviewNoteInvalid},
		{name: "note trop haute", status: domain.ExchangeStatusCompleted, authorID: 1, note: 6, wantErr: domain.ErrReviewNoteInvalid},
		{name: "échange non terminé", status: domain.ExchangeStatusAccepted, authorID: 1, note: 5, wantErr: domain.ErrReviewExchangeIncomplete},
		{name: "utilisateur extérieur", status: domain.ExchangeStatusCompleted, authorID: 3, note: 5, wantErr: domain.ErrReviewForbidden},
	}

	for _, test := range tests {
		testContext.Run(test.name, func(testCaseContext *testing.T) {
			repository := reviewRepositoryStub{exchange: domain.Exchange{
				ID: 8, RequesterID: 1, OwnerID: 2, Status: test.status,
			}}
			service := NewReviewService(repository)
			review, err := service.Create(context.Background(), 8, test.authorID, domain.CreateReviewInput{
				Note: test.note, Commentaire: "  Très bien  ",
			})

			if !errors.Is(err, test.wantErr) {
				testCaseContext.Fatalf("Create() error = %v, want %v", err, test.wantErr)
			}
			if err == nil && review.TargetID != test.wantTarget {
				testCaseContext.Fatalf("target_id = %d, want %d", review.TargetID, test.wantTarget)
			}
			if err == nil && review.Commentaire != "Très bien" {
				testCaseContext.Fatalf("commentaire = %q, want %q", review.Commentaire, "Très bien")
			}
		})
	}
}

func TestReviewCannotBeCreatedTwice(testContext *testing.T) {
	repository := reviewRepositoryStub{
		exchange:  domain.Exchange{ID: 8, RequesterID: 1, OwnerID: 2, Status: domain.ExchangeStatusCompleted},
		createErr: domain.ErrReviewAlreadyExists,
	}
	service := NewReviewService(repository)
	_, err := service.Create(context.Background(), 8, 1, domain.CreateReviewInput{Note: 5})
	if !errors.Is(err, domain.ErrReviewAlreadyExists) {
		testContext.Fatalf("Create() error = %v, want %v", err, domain.ErrReviewAlreadyExists)
	}
}
