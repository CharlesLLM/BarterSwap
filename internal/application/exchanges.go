package application

import (
	"context"
	"strings"

	"github.com/CharlesLLM/BarterSwap/internal/domain"
)

type ExchangeRepository interface {
	FindService(context.Context, int) (domain.Service, error)
	FindUser(context.Context, int) (domain.User, error)
	CreditBalance(context.Context, int) (int, error)
	CreateExchange(context.Context, int, domain.Service) (domain.Exchange, error)
	ListExchanges(context.Context, int, domain.ExchangeFilter) ([]domain.Exchange, error)
	FindExchange(context.Context, int) (domain.Exchange, error)
	UpdateExchangeStatus(context.Context, int, string, string, []domain.CreditChange) (domain.Exchange, error)
}

type ExchangeService struct {
	repository ExchangeRepository
}

func NewExchangeService(repository ExchangeRepository) *ExchangeService {
	return &ExchangeService{repository: repository}
}

func (service *ExchangeService) Create(ctx context.Context, requesterID int, input domain.CreateExchangeInput) (domain.Exchange, error) {
	if input.ServiceID <= 0 {
		return domain.Exchange{}, domain.ErrExchangeServiceRequired
	}

	offeredService, err := service.repository.FindService(ctx, input.ServiceID)
	if err != nil {
		return domain.Exchange{}, err
	}
	if offeredService.ProviderID == requesterID {
		return domain.Exchange{}, domain.ErrExchangeSelfService
	}

	balance, err := service.repository.CreditBalance(ctx, requesterID)
	if err != nil {
		return domain.Exchange{}, err
	}
	if balance < offeredService.Credits {
		return domain.Exchange{}, domain.ErrExchangeInsufficientFund
	}

	return service.repository.CreateExchange(ctx, requesterID, offeredService)
}

func (service *ExchangeService) List(ctx context.Context, userID int, filter domain.ExchangeFilter) ([]domain.Exchange, error) {
	filter.Status = strings.ToLower(strings.TrimSpace(filter.Status))
	if filter.Status != "" && !domain.IsValidExchangeStatus(filter.Status) {
		return nil, domain.ErrExchangeStatusInvalid
	}
	if _, err := service.repository.FindUser(ctx, userID); err != nil {
		return nil, err
	}
	return service.repository.ListExchanges(ctx, userID, filter)
}

func (service *ExchangeService) Get(ctx context.Context, userID, exchangeID int) (domain.Exchange, error) {
	exchange, err := service.repository.FindExchange(ctx, exchangeID)
	if err != nil {
		return domain.Exchange{}, err
	}
	if !isExchangeParticipant(exchange, userID) {
		return domain.Exchange{}, domain.ErrExchangeForbidden
	}
	return exchange, nil
}

func (service *ExchangeService) Accept(ctx context.Context, userID, exchangeID int) (domain.Exchange, error) {
	exchange, err := service.pendingExchangeForOwner(ctx, userID, exchangeID)
	if err != nil {
		return domain.Exchange{}, err
	}

	return service.repository.UpdateExchangeStatus(ctx, exchange.ID, domain.ExchangeStatusPending, domain.ExchangeStatusAccepted, []domain.CreditChange{
		{UserID: exchange.RequesterID, Montant: -exchange.Credits, Type: domain.CreditTypeSpend},
	})
}

func (service *ExchangeService) Reject(ctx context.Context, userID, exchangeID int) (domain.Exchange, error) {
	exchange, err := service.pendingExchangeForOwner(ctx, userID, exchangeID)
	if err != nil {
		return domain.Exchange{}, err
	}
	return service.repository.UpdateExchangeStatus(ctx, exchange.ID, domain.ExchangeStatusPending, domain.ExchangeStatusRejected, nil)
}

func (service *ExchangeService) Complete(ctx context.Context, userID, exchangeID int) (domain.Exchange, error) {
	exchange, err := service.repository.FindExchange(ctx, exchangeID)
	if err != nil {
		return domain.Exchange{}, err
	}
	if !isExchangeParticipant(exchange, userID) {
		return domain.Exchange{}, domain.ErrExchangeForbidden
	}
	if exchange.Status != domain.ExchangeStatusAccepted {
		return domain.Exchange{}, domain.ErrExchangeTransition
	}

	return service.repository.UpdateExchangeStatus(ctx, exchange.ID, domain.ExchangeStatusAccepted, domain.ExchangeStatusCompleted, []domain.CreditChange{
		{UserID: exchange.OwnerID, Montant: exchange.Credits, Type: domain.CreditTypeEarn},
	})
}

func (service *ExchangeService) Cancel(ctx context.Context, userID, exchangeID int) (domain.Exchange, error) {
	exchange, err := service.repository.FindExchange(ctx, exchangeID)
	if err != nil {
		return domain.Exchange{}, err
	}
	if !isExchangeParticipant(exchange, userID) {
		return domain.Exchange{}, domain.ErrExchangeForbidden
	}

	var creditChanges []domain.CreditChange
	switch exchange.Status {
	case domain.ExchangeStatusPending:
	case domain.ExchangeStatusAccepted:
		creditChanges = []domain.CreditChange{
			{UserID: exchange.RequesterID, Montant: exchange.Credits, Type: domain.CreditTypeRefund},
		}
	default:
		return domain.Exchange{}, domain.ErrExchangeTransition
	}

	return service.repository.UpdateExchangeStatus(ctx, exchange.ID, exchange.Status, domain.ExchangeStatusCancelled, creditChanges)
}

func (service *ExchangeService) pendingExchangeForOwner(ctx context.Context, userID, exchangeID int) (domain.Exchange, error) {
	exchange, err := service.repository.FindExchange(ctx, exchangeID)
	if err != nil {
		return domain.Exchange{}, err
	}
	if exchange.OwnerID != userID {
		return domain.Exchange{}, domain.ErrExchangeForbidden
	}
	if exchange.Status != domain.ExchangeStatusPending {
		return domain.Exchange{}, domain.ErrExchangeTransition
	}
	return exchange, nil
}

func isExchangeParticipant(exchange domain.Exchange, userID int) bool {
	return exchange.RequesterID == userID || exchange.OwnerID == userID
}
