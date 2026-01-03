package httpapi

import "net/http"

func NewRouter(handler *InvoiceHandler) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("POST /invoices", handler.CreateInvoice)
	mux.HandleFunc("POST /invoices/{id}/pay", handler.RequestPayment)

	return mux
}
