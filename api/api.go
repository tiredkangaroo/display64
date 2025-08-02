package api

import (
	"net/http"

	"github.com/tiredkangaroo/display64/providers"
)

func UseHandler(providers *providers.Providers) http.Handler {
	mux := http.NewServeMux()

	// Define API routes
	mux.HandleFunc("/v1/providers/start", func(w http.ResponseWriter, r *http.Request) {
		providerName := r.URL.Query().Get("name")

		provider, ok := providers.GetProvider(providerName)
		if !ok {
			http.Error(w, "provider not found", http.StatusNotFound)
			return
		}
		err := providers.Start(provider)
		if err != nil {
			http.Error(w, err.Error(), http.StatusUnprocessableEntity)
			return
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Provider started successfully"))
	})

	return mux
}
