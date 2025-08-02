package spotify

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"time"

	"github.com/tiredkangaroo/display64/env"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/spotify"
)

// provider flow -> kept provider in a struct and Init -> when its time to use the provider, check ok -> if no, error out -> if yes, set the on image
// when its time for outsies, set the on image to nil

const FETCH_INTERVAL = time.Duration(1500 * time.Millisecond)

type Provider struct {
	oAuthConfig   *oauth2.Config
	currentClient *http.Client
	working       bool
}

func (p *Provider) Init() error {
	p.oAuthConfig = &oauth2.Config{
		ClientID:     env.DefaultEnvironment.SpotifyClientID,
		ClientSecret: env.DefaultEnvironment.SpotifyClientSecret,
		RedirectURL:  env.DefaultEnvironment.SpotifyRedirectURI,
		Scopes:       []string{env.DefaultEnvironment.SpotifyScopes},
		Endpoint:     spotify.Endpoint,
	}
	p.currentClient = nil

	http.HandleFunc("/spotify/redirect", func(w http.ResponseWriter, r *http.Request) {
		code := r.URL.Query().Get("code")
		if code == "" {
			http.Error(w, "no code provided", http.StatusBadRequest)
			return
		}

		token, err := p.oAuthConfig.Exchange(r.Context(), code)
		if err != nil {
			http.Error(w, "failed to exchange code for token: "+err.Error(), http.StatusInternalServerError)
			return
		}

		client := p.oAuthConfig.Client(r.Context(), token)
		p.currentClient = client
	})

	http.HandleFunc("/spotify/auth", func(w http.ResponseWriter, r *http.Request) {
		location := p.oAuthConfig.AuthCodeURL("state", oauth2.AccessTypeOffline)
		w.WriteHeader(http.StatusFound)
		w.Header().Add("Location", location)
	})

	return nil
}

func (p *Provider) Authorized() bool {
	return p.currentClient != nil
}

func (p *Provider) Start(onImageFunc func(io.Reader)) {
	p.working = true

	var lastImageURL string
	for p.working {
		u, err := p.getCurrentlyPlayingImageURL()
		if err != nil {
			slog.Error("get currently playing image URL", "error", err)
			continue
		}
		if u == lastImageURL {
			continue
		}

		rd, err := getResponseBodyFromURL(u)
		if err != nil {
			if rd != nil { // ensure we close the reader if it was opened
				rd.Close()
			}
			slog.Error("get response body from URL", "error", err, "url", u)
			continue
		}

		onImageFunc(rd)
		rd.Close() // close the reader after passing it to the callback

		lastImageURL = u
		time.Sleep(FETCH_INTERVAL)
	}
}

func (p *Provider) Stop() {
	p.working = false
}

func (p *Provider) getCurrentlyPlayingImageURL() (string, error) {
	// obvious spotifyResponse isn't the whole response structure, it has the necessary fields for the image URL
	type spotifyResponse struct {
		Item struct {
			Album struct {
				Images []struct {
					URL string `json:"url"`
				} `json:"images"`
			} `json:"album"`
		} `json:"item"`
		Error struct {
			Status  int    `json:"status"`
			Message string `json:"message"`
		} `json:"error"`
	}
	resp, err := p.currentClient.Get("https://api.spotify.com/v1/me/player/currently-playing")
	if err != nil {
		return "", fmt.Errorf("get currently playing api call: %w", err)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("read response body: %w", err)
	}
	var spResp spotifyResponse
	if err := json.Unmarshal(body, &spResp); err != nil {
		slog.Error("decode response", "error", err, "body", string(body))
		return "", fmt.Errorf("decode response: %w", err)
	}
	if spResp.Error.Message != "" {
		return "", fmt.Errorf("spotify api error: %s", spResp.Error.Message)
	}
	if len(spResp.Item.Album.Images) == 0 {
		return "", fmt.Errorf("no images found in currently playing item")
	}
	imageURL := spResp.Item.Album.Images[0].URL
	if imageURL == "" {
		return "", fmt.Errorf("no image URL found in currently playing item")
	}
	return imageURL, nil
}
