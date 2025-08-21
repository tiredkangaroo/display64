package spotify

import (
	"context"
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

const NTS_LINK = "https://hc-cdn.hel1.your-objectstorage.com/s/v3/81ddba41db872a37a630dbb071f57ba4f916b019_image.png"

// provider flow -> kept provider in a struct and Init -> when its time to use the provider, check ok -> if no, error out -> if yes, set the on image
// when its time for outsies, set the on image to nil

const FETCH_INTERVAL = time.Duration(1700 * time.Millisecond)

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

	http.HandleFunc("/api/v1/spotify/redirect", func(w http.ResponseWriter, r *http.Request) {
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

		client := p.oAuthConfig.Client(context.Background(), token)
		p.currentClient = client

		w.Header().Set("Content-Type", "text/html")
		fmt.Fprintln(w,
			"<html><body><h1>Authorization successful!</h1><p>You can close this window.</p></body><script>window.open(\"\", \"_self\");window.close();</script></html>")
	})

	http.HandleFunc("/api/v1/spotify/auth", func(w http.ResponseWriter, r *http.Request) {
		location := p.oAuthConfig.AuthCodeURL("state", oauth2.AccessTypeOffline)
		http.Redirect(w, r, location, http.StatusFound)
	})

	return nil
}

func (p *Provider) Authorized() bool {
	return p.currentClient != nil
}
func (p *Provider) AuthorizationURL() string {
	if env.DefaultEnvironment.Debug {
		return "http://localhost:9000/api/v1/spotify/auth"
	}
	return "/api/v1/spotify/auth"
}

func (p *Provider) Start(onImageFunc func(u string)) {
	p.working = true

	for p.working {
		time.Sleep(FETCH_INTERVAL)
		u, err := p.getCurrentlyPlayingImageURL()
		if err != nil {
			slog.Error("get currently playing image URL", "error", err)
			continue
		}
		onImageFunc(u)
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
	if len(body) == 0 { // nothing playing atm
		return NTS_LINK, nil
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
