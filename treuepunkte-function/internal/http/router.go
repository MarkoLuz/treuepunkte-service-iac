package httpx

import "net/http"

func Router(h *Handlers) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("POST /v1/points/accrue", h.PostAccrue)
	mux.HandleFunc("POST /v1/points/confirm", h.PostConfirm)
	mux.HandleFunc("POST /v1/points/revoke", h.PostRevoke)
	mux.HandleFunc("POST /v1/points/redeem", h.PostRedeem)
	mux.HandleFunc("POST /v1/points/restore", h.PostRestore)

	mux.HandleFunc("GET /v1/customers/", h.GetCustomerRoutes)

	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})

	return mux
}