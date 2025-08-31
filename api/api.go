package api

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/tiredkangaroo/display64/env"
	"github.com/tiredkangaroo/display64/providers"
	"github.com/tiredkangaroo/websocket"
)

func UseHandler(providers *providers.Providers) http.Handler {
	mux := http.NewServeMux()

	// Define API routes

	mux.HandleFunc("GET /v1/providers", func(w http.ResponseWriter, r *http.Request) {
		cors(w)

		data, err := json.Marshal(providers.List())
		if err != nil {
			http.Error(w, "failed to marshal providers: "+err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(data)
	})

	mux.HandleFunc("PUT /v1/providers/start", func(w http.ResponseWriter, r *http.Request) {
		cors(w)

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

	mux.HandleFunc("GET /v1/imageURL", func(w http.ResponseWriter, r *http.Request) {
		cors(w)

		conn, err := websocket.AcceptHTTP(w, r)
		if err != nil {
			http.Error(w, "failed to upgrade to websocket: "+err.Error(), http.StatusInternalServerError)
			return
		}
		if providers.LastImageURL != "" {
			conn.Write(&websocket.Message{
				Type: websocket.MessageText,
				Data: []byte(providers.LastImageURL),
			})
		}
		ctx, cancel := context.WithCancel(context.Background())
		providers.NewImageURLFunc = func(url string) {
			if err := conn.Write(&websocket.Message{
				Type: websocket.MessageText,
				Data: []byte(url),
			}); err != nil {
				providers.NewImageURLFunc = nil
				cancel()
			}
		}
		<-ctx.Done()
	})

	mux.HandleFunc("OPTIONS /", func(w http.ResponseWriter, r *http.Request) {
		cors(w)
		w.WriteHeader(http.StatusOK)
	})
	return mux
}

func cors(w http.ResponseWriter) {
	if env.DefaultEnvironment.Debug {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
	}
}
