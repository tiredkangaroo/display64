package providers

import (
	"bytes"
	"fmt"
	"io"
	"log/slog"
	"mime/multipart"
	"net/http"

	"github.com/tiredkangaroo/display64/env"
	"github.com/tiredkangaroo/display64/noprovider"
	"github.com/tiredkangaroo/display64/spotify"
)

var (
	ErrUnauthorized = fmt.Errorf("provider not authorized")
)

type Providers struct {
	Current    Provider
	NoProvider Provider
	Spotify    *spotify.Provider

	LastImageURL string
}

type Provider interface {
	Authorized() bool
	Start(func(string))
	Stop()
}

func (p *Providers) Init() error {
	p.Spotify = &spotify.Provider{}
	err := p.Spotify.Init()
	if err != nil {
		return fmt.Errorf("initialize Spotify provider: %w", err)
	}

	p.NoProvider = &noprovider.Provider{}

	return nil
}

func (p *Providers) Start(provider Provider) error {
	if !provider.Authorized() {
		slog.Error("provider not authorized", "provider", fmt.Sprintf("%T", provider))
		return ErrUnauthorized
	}
	if p.Current != nil {
		p.Current.Stop()
	}
	p.Current = provider

	slog.Info("provider starting", "provider", fmt.Sprintf("%T", provider))

	go p.Current.Start(func(u string) {
		if u == p.LastImageURL {
			slog.Info("image URL is the same as last, skip", "url", u)
			return
		}
		slog.Info("sending image to display", "url", u)
		if err := sendURLToDisplay(u); err != nil {
			slog.Error("sending image to display", "error", err)
		} else {
			slog.Info("image sent to display successfully")
			p.LastImageURL = u
		}
	})
	return nil
}

func (p *Providers) GetProvider(name string) (Provider, bool) {
	switch name {
	case "Spotify":
		return p.Spotify, true
	case "None":
		return p.NoProvider, true
	default:
		return nil, false
	}
}

func (p *Providers) List() []any {
	type providerInfo struct {
		Name             string `json:"name"`
		Authorized       bool   `json:"authorized"`
		AuthorizationURL string `json:"authorization_url,omitempty"`
		IsCurrent        bool   `json:"is_current,omitempty"`
	}
	return []any{
		providerInfo{
			Name:             "Spotify",
			Authorized:       p.Spotify.Authorized(),
			AuthorizationURL: p.Spotify.AuthorizationURL(),
			IsCurrent:        p.Current == p.Spotify,
		},
		providerInfo{
			Name:             "None",
			Authorized:       p.NoProvider.Authorized(),
			AuthorizationURL: "",
			IsCurrent:        p.Current == p.NoProvider,
		},
	}
}

func sendURLToDisplay(u string) error {
	req, err := http.NewRequest("GET", u, nil)
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("send request: %w", err)
	}
	defer resp.Body.Close()

	rd := resp.Body
	// create a multipart form with rd
	buf := new(bytes.Buffer)
	writer := multipart.NewWriter(buf)
	part, err := writer.CreateFormFile("file", "image.png")
	if err != nil {
		return fmt.Errorf("create form file: %w", err)
	}
	if _, err := io.Copy(part, rd); err != nil {
		return fmt.Errorf("copy image data: %w", err)
	}
	writer.Close()

	req, err = http.NewRequest("POST", env.DefaultEnvironment.DisplayServerURL+"/use", buf)
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())

	resp, err = http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("send request: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("server error: %s", body)
	}
	return nil
}
