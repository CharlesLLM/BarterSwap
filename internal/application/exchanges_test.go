package application

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"testing"

	"github.com/CharlesLLM/BarterSwap/internal/domain"
)

type exchangeRepositoryStub struct {
	service     domain.Service
	balance     int
	exchange    domain.Exchange
	wantFrom    string
	wantTo      string
	wantChanges []domain.CreditChange
}

func (repository exchangeRepositoryStub) FindService(context.Context, int) (domain.Service, error) {
	return repository.service, nil
}

func (repository exchangeRepositoryStub) FindUser(context.Context, int) (domain.User, error) {
	return domain.User{ID: 1}, nil
}

func (repository exchangeRepositoryStub) CreditBalance(context.Context, int) (int, error) {
	return repository.balance, nil
}

func (repository exchangeRepositoryStub) CreateExchange(context.Context, int, domain.Service) (domain.Exchange, error) {
	return repository.exchange, nil
}

func (repository exchangeRepositoryStub) ListExchanges(context.Context, int, domain.ExchangeFilter) ([]domain.Exchange, error) {
	return []domain.Exchange{repository.exchange}, nil
}

func (repository exchangeRepositoryStub) FindExchange(context.Context, int) (domain.Exchange, error) {
	return repository.exchange, nil
}

func (repository exchangeRepositoryStub) UpdateExchangeStatus(
	_ context.Context,
	_ int,
	from string,
	to string,
	changes []domain.CreditChange,
) (domain.Exchange, error) {
	if from != repository.wantFrom || to != repository.wantTo {
		return domain.Exchange{}, fmt.Errorf("transition reçue %s -> %s", from, to)
	}
	if !reflect.DeepEqual(changes, repository.wantChanges) {
		return domain.Exchange{}, fmt.Errorf("mouvements de crédit reçus : %+v", changes)
	}
	exchange := repository.exchange
	exchange.Status = to
	return exchange, nil
}

func TestExchangeCreate(testContext *testing.T) {
	tests := []struct {
		name        string
		requesterID int
		providerID  int
		balance     int
		wantErr     error
	}{
		{name: "demande valide", requesterID: 1, providerID: 2, balance: 10},
		{name: "propre service", requesterID: 2, providerID: 2, balance: 10, wantErr: domain.ErrExchangeSelfService},
		{name: "crédits insuffisants", requesterID: 1, providerID: 2, balance: 3, wantErr: domain.ErrExchangeInsufficientFund},
	}

	for _, test := range tests {
		testContext.Run(test.name, func(testCaseContext *testing.T) {
			repository := exchangeRepositoryStub{
				service:  domain.Service{ID: 5, ProviderID: test.providerID, Credits: 4},
				balance:  test.balance,
				exchange: domain.Exchange{ID: 8, Status: domain.ExchangeStatusPending},
			}
			service := NewExchangeService(repository)
			_, err := service.Create(context.Background(), test.requesterID, domain.CreateExchangeInput{ServiceID: 5})
			if !errors.Is(err, test.wantErr) {
				testCaseContext.Fatalf("Create() error = %v, want %v", err, test.wantErr)
			}
		})
	}
}

func TestExchangeTransitions(testContext *testing.T) {
	tests := []struct {
		name        string
		action      string
		status      string
		userID      int
		wantStatus  string
		wantErr     error
		wantChanges []domain.CreditChange
	}{
		{
			name: "acceptation", action: "accept", status: domain.ExchangeStatusPending, userID: 2,
			wantStatus:  domain.ExchangeStatusAccepted,
			wantChanges: []domain.CreditChange{{UserID: 1, Montant: -4, Type: domain.CreditTypeSpend}},
		},
		{name: "refus", action: "reject", status: domain.ExchangeStatusPending, userID: 2, wantStatus: domain.ExchangeStatusRejected},
		{
			name: "complétion", action: "complete", status: domain.ExchangeStatusAccepted, userID: 1,
			wantStatus:  domain.ExchangeStatusCompleted,
			wantChanges: []domain.CreditChange{{UserID: 2, Montant: 4, Type: domain.CreditTypeEarn}},
		},
		{
			name: "annulation avec remboursement", action: "cancel", status: domain.ExchangeStatusAccepted, userID: 2,
			wantStatus:  domain.ExchangeStatusCancelled,
			wantChanges: []domain.CreditChange{{UserID: 1, Montant: 4, Type: domain.CreditTypeRefund}},
		},
		{name: "annulation en attente", action: "cancel", status: domain.ExchangeStatusPending, userID: 1, wantStatus: domain.ExchangeStatusCancelled},
		{name: "acceptation interdite", action: "accept", status: domain.ExchangeStatusPending, userID: 1, wantErr: domain.ErrExchangeForbidden},
		{name: "complétion impossible", action: "complete", status: domain.ExchangeStatusPending, userID: 1, wantErr: domain.ErrExchangeTransition},
	}

	for _, test := range tests {
		testContext.Run(test.name, func(testCaseContext *testing.T) {
			repository := exchangeRepositoryStub{
				exchange: domain.Exchange{ID: 8, RequesterID: 1, OwnerID: 2, Status: test.status, Credits: 4},
				wantFrom: test.status, wantTo: test.wantStatus, wantChanges: test.wantChanges,
			}
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
		})
	}
}

func TestExchangeListRejectsInvalidStatus(testContext *testing.T) {
	service := NewExchangeService(exchangeRepositoryStub{})
	_, err := service.List(context.Background(), 1, domain.ExchangeFilter{Status: "inconnu"})
	if !errors.Is(err, domain.ErrExchangeStatusInvalid) {
		testContext.Fatalf("List() error = %v, want %v", err, domain.ErrExchangeStatusInvalid)
	}
}

func TestExchangeListAndGet(testContext *testing.T) {
	repository := exchangeRepositoryStub{exchange: domain.Exchange{
		ID: 8, RequesterID: 1, OwnerID: 2, Status: domain.ExchangeStatusPending,
	}}
	service := NewExchangeService(repository)

	exchanges, err := service.List(context.Background(), 1, domain.ExchangeFilter{Status: " PENDING "})
	if err != nil || len(exchanges) != 1 {
		testContext.Fatalf("List() = %+v, %v", exchanges, err)
	}
	if exchange, err := service.Get(context.Background(), 1, 8); err != nil || exchange.ID != 8 {
		testContext.Fatalf("Get() = %+v, %v", exchange, err)
	}
	if _, err := service.Get(context.Background(), 3, 8); !errors.Is(err, domain.ErrExchangeForbidden) {
		testContext.Fatalf("Get() error = %v, want %v", err, domain.ErrExchangeForbidden)
	}
}
