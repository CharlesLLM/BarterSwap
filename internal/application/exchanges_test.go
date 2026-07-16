package application

import (
	"context"
	"errors"
	"testing"

	"github.com/CharlesLLM/BarterSwap/internal/domain"
)

type fakeExchangeRepository struct {
	offeredService domain.Service
	user           domain.User
	balance        int
	exchange       domain.Exchange
	createErr      error
	updateErr      error
	updatedFrom    string
	updatedTo      string
	creditChanges  []domain.CreditChange
}

func (repository *fakeExchangeRepository) FindService(context.Context, int) (domain.Service, error) {
	return repository.offeredService, nil
}

func (repository *fakeExchangeRepository) FindUser(context.Context, int) (domain.User, error) {
	return repository.user, nil
}

func (repository *fakeExchangeRepository) CreditBalance(context.Context, int) (int, error) {
	return repository.balance, nil
}

func (repository *fakeExchangeRepository) CreateExchange(context.Context, int, domain.Service) (domain.Exchange, error) {
	return repository.exchange, repository.createErr
}

func (repository *fakeExchangeRepository) ListExchanges(context.Context, int, domain.ExchangeFilter) ([]domain.Exchange, error) {
	return []domain.Exchange{repository.exchange}, nil
}

func (repository *fakeExchangeRepository) FindExchange(context.Context, int) (domain.Exchange, error) {
	return repository.exchange, nil
}

func (repository *fakeExchangeRepository) UpdateExchangeStatus(
	_ context.Context,
	_ int,
	expectedStatus string,
	newStatus string,
	creditChanges []domain.CreditChange,
) (domain.Exchange, error) {
	repository.updatedFrom = expectedStatus
	repository.updatedTo = newStatus
	repository.creditChanges = creditChanges
	updated := repository.exchange
	updated.Status = newStatus
	return updated, repository.updateErr
}

func TestExchangeServiceCreate(testContext *testing.T) {
	tests := []struct {
		name        string
		requesterID int
		service     domain.Service
		balance     int
		wantErr     error
	}{
		{
			name:        "demande valide",
			requesterID: 1,
			service:     domain.Service{ID: 5, ProviderID: 2, Credits: 4},
			balance:     10,
		},
		{
			name:        "propre service",
			requesterID: 2,
			service:     domain.Service{ID: 5, ProviderID: 2, Credits: 4},
			balance:     10,
			wantErr:     domain.ErrExchangeSelfService,
		},
		{
			name:        "crédits insuffisants",
			requesterID: 1,
			service:     domain.Service{ID: 5, ProviderID: 2, Credits: 4},
			balance:     3,
			wantErr:     domain.ErrExchangeInsufficientFund,
		},
	}

	for _, test := range tests {
		testContext.Run(test.name, func(testCaseContext *testing.T) {
			repository := &fakeExchangeRepository{
				offeredService: test.service,
				balance:        test.balance,
				exchange:       domain.Exchange{ID: 8, Status: domain.ExchangeStatusPending},
			}
			service := NewExchangeService(repository)
			_, err := service.Create(context.Background(), test.requesterID, domain.CreateExchangeInput{ServiceID: test.service.ID})
			if !errors.Is(err, test.wantErr) {
				testCaseContext.Fatalf("Create() error = %v, want %v", err, test.wantErr)
			}
		})
	}
}

func TestExchangeServiceTransitions(testContext *testing.T) {
	tests := []struct {
		name       string
		action     string
		status     string
		userID     int
		wantStatus string
		wantErr    error
		wantChange *domain.CreditChange
	}{
		{
			name:       "acceptation par le propriétaire",
			action:     "accept",
			status:     domain.ExchangeStatusPending,
			userID:     2,
			wantStatus: domain.ExchangeStatusAccepted,
			wantChange: &domain.CreditChange{UserID: 1, Montant: -4, Type: domain.CreditTypeSpend},
		},
		{
			name:       "refus par le propriétaire",
			action:     "reject",
			status:     domain.ExchangeStatusPending,
			userID:     2,
			wantStatus: domain.ExchangeStatusRejected,
		},
		{
			name:       "complétion par un participant",
			action:     "complete",
			status:     domain.ExchangeStatusAccepted,
			userID:     1,
			wantStatus: domain.ExchangeStatusCompleted,
			wantChange: &domain.CreditChange{UserID: 2, Montant: 4, Type: domain.CreditTypeEarn},
		},
		{
			name:       "annulation acceptée et remboursement",
			action:     "cancel",
			status:     domain.ExchangeStatusAccepted,
			userID:     2,
			wantStatus: domain.ExchangeStatusCancelled,
			wantChange: &domain.CreditChange{UserID: 1, Montant: 4, Type: domain.CreditTypeRefund},
		},
		{
			name:       "annulation en attente sans mouvement de crédit",
			action:     "cancel",
			status:     domain.ExchangeStatusPending,
			userID:     1,
			wantStatus: domain.ExchangeStatusCancelled,
		},
		{
			name:    "acceptation par le demandeur interdite",
			action:  "accept",
			status:  domain.ExchangeStatusPending,
			userID:  1,
			wantErr: domain.ErrExchangeForbidden,
		},
		{
			name:    "complétion avant acceptation impossible",
			action:  "complete",
			status:  domain.ExchangeStatusPending,
			userID:  1,
			wantErr: domain.ErrExchangeTransition,
		},
	}

	for _, test := range tests {
		testContext.Run(test.name, func(testCaseContext *testing.T) {
			repository := &fakeExchangeRepository{exchange: domain.Exchange{
				ID: 8, ServiceID: 5, RequesterID: 1, OwnerID: 2, Status: test.status, Credits: 4,
			}}
			service := NewExchangeService(repository)

			var err error
			switch test.action {
			case "accept":
				_, err = service.Accept(context.Background(), test.userID, 8)
			case "reject":
				_, err = service.Reject(context.Background(), test.userID, 8)
			case "complete":
				_, err = service.Complete(context.Background(), test.userID, 8)
			case "cancel":
				_, err = service.Cancel(context.Background(), test.userID, 8)
			}

			if !errors.Is(err, test.wantErr) {
				testCaseContext.Fatalf("%s() error = %v, want %v", test.action, err, test.wantErr)
			}
			if test.wantErr != nil {
				return
			}
			if repository.updatedFrom != test.status || repository.updatedTo != test.wantStatus {
				testCaseContext.Fatalf("transition = %s -> %s, want %s -> %s", repository.updatedFrom, repository.updatedTo, test.status, test.wantStatus)
			}
			if test.wantChange == nil {
				if len(repository.creditChanges) != 0 {
					testCaseContext.Fatalf("creditChanges = %+v, want aucun mouvement", repository.creditChanges)
				}
				return
			}
			if len(repository.creditChanges) != 1 || repository.creditChanges[0] != *test.wantChange {
				testCaseContext.Fatalf("creditChanges = %+v, want %+v", repository.creditChanges, *test.wantChange)
			}
		})
	}
}

func TestExchangeServiceListRejectsInvalidStatus(testContext *testing.T) {
	service := NewExchangeService(&fakeExchangeRepository{})
	_, err := service.List(context.Background(), 1, domain.ExchangeFilter{Status: "inconnu"})
	if !errors.Is(err, domain.ErrExchangeStatusInvalid) {
		testContext.Fatalf("List() error = %v, want %v", err, domain.ErrExchangeStatusInvalid)
	}
}
