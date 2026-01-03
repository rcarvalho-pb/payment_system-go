package httpapi

import (
	"encoding/json"
	"net/http"

	invoiceApplication "github.com/rcarvalho-pb/payment_system-go/internal/application/invoice"
)

type InvoiceHandler struct {
	Service *invoiceApplication.Service
}

type CreateInvoiceRequest struct {
	ID     string `json:"id"`
	Amount int64  `json:"amount"`
}

func (h *InvoiceHandler) CreateInvoice(w http.ResponseWriter, r *http.Request) {
	var req CreateInvoiceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	inv, err := h.Service.CreateInvoice(req.ID, req.Amount)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(inv)
}

func (h *InvoiceHandler) RequestPayment(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	if err := h.Service.RequestPayment(id); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusAccepted)
}
