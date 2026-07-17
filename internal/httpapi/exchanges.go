package httpapi

import (
	"net/http"

	"github.com/CharlesLLM/BarterSwap/internal/domain"
)

func (handler Handler) createExchange(responseWriter http.ResponseWriter, request *http.Request) {
	userID, valid := requireUserID(responseWriter, request)
	if !valid {
		return
	}
	var input domain.CreateExchangeInput
	if !decodeJSON(responseWriter, request, &input) {
		return
	}

	exchange, err := handler.exchanges.Create(request.Context(), userID, input)
	if err != nil {
		writeApplicationError(responseWriter, err, "création de l'échange")
		return
	}
	writeJSON(responseWriter, http.StatusCreated, exchange)
}

func (handler Handler) listExchanges(responseWriter http.ResponseWriter, request *http.Request) {
	userID, valid := requireUserID(responseWriter, request)
	if !valid {
		return
	}
	filter := domain.ExchangeFilter{Status: request.URL.Query().Get("status")}
	exchanges, err := handler.exchanges.List(request.Context(), userID, filter)
	if err != nil {
		writeApplicationError(responseWriter, err, "liste des échanges")
		return
	}
	writeJSON(responseWriter, http.StatusOK, exchanges)
}

func (handler Handler) getExchange(responseWriter http.ResponseWriter, request *http.Request, exchangeID int) {
	userID, valid := requireUserID(responseWriter, request)
	if !valid {
		return
	}
	exchange, err := handler.exchanges.Get(request.Context(), userID, exchangeID)
	if err != nil {
		writeApplicationError(responseWriter, err, "lecture de l'échange")
		return
	}
	writeJSON(responseWriter, http.StatusOK, exchange)
}

func (handler Handler) acceptExchange(responseWriter http.ResponseWriter, request *http.Request, exchangeID int) {
	handler.updateExchangeStatus(responseWriter, request, exchangeID, "accept")
}

func (handler Handler) rejectExchange(responseWriter http.ResponseWriter, request *http.Request, exchangeID int) {
	handler.updateExchangeStatus(responseWriter, request, exchangeID, "reject")
}

func (handler Handler) completeExchange(responseWriter http.ResponseWriter, request *http.Request, exchangeID int) {
	handler.updateExchangeStatus(responseWriter, request, exchangeID, "complete")
}

func (handler Handler) cancelExchange(responseWriter http.ResponseWriter, request *http.Request, exchangeID int) {
	handler.updateExchangeStatus(responseWriter, request, exchangeID, "cancel")
}

func (handler Handler) updateExchangeStatus(responseWriter http.ResponseWriter, request *http.Request, exchangeID int, action string) {
	userID, valid := requireUserID(responseWriter, request)
	if !valid {
		return
	}

	var (
		exchange domain.Exchange
		err      error
	)
	switch action {
	case "accept":
		exchange, err = handler.exchanges.Accept(request.Context(), userID, exchangeID)
	case "reject":
		exchange, err = handler.exchanges.Reject(request.Context(), userID, exchangeID)
	case "complete":
		exchange, err = handler.exchanges.Complete(request.Context(), userID, exchangeID)
	case "cancel":
		exchange, err = handler.exchanges.Cancel(request.Context(), userID, exchangeID)
	default:
		writeError(responseWriter, http.StatusNotFound, "route introuvable")
		return
	}
	if err != nil {
		writeApplicationError(responseWriter, err, "modification de l'échange")
		return
	}
	writeJSON(responseWriter, http.StatusOK, exchange)
}
