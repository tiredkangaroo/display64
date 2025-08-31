package providers

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	"image/png"
	_ "image/png"
	"log/slog"
	"net/http"

	"golang.org/x/image/draw"
	_ "golang.org/x/image/webp"

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

	LastImageURL    string
	NewImageURLFunc func(string)
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
			return
		} else if p.NewImageURLFunc != nil {
			p.NewImageURLFunc(u)
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

	img, _, err := image.Decode(resp.Body)
	if err != nil {
		return fmt.Errorf("decode image: %w", err)
	}

	// img -> rgba conversion
	bounds := img.Bounds()
	rgbImg := image.NewRGBA(bounds)
	draw.Draw(rgbImg, bounds, img, bounds.Min, draw.Src)

	// the rgba image to 64x64
	thumb := image.NewRGBA(image.Rect(0, 0, 64, 64))
	draw.CatmullRom.Scale(thumb, thumb.Bounds(), rgbImg, bounds, draw.Over, nil)

	buf := new(bytes.Buffer)
	if err := png.Encode(buf, thumb); err != nil {
		return fmt.Errorf("encode image: %w", err)
	}

	var imgLengthData [8]byte
	binary.BigEndian.PutUint64(imgLengthData[:], uint64(buf.Len()))
	if _, err := env.DefaultEnvironment.DisplayConnection.Write(imgLengthData[:]); err != nil {
		return fmt.Errorf("send image length to display: %w", err)
	}
	if _, err := buf.WriteTo(env.DefaultEnvironment.DisplayConnection); err != nil {
		return fmt.Errorf("send image to display: %w", err)
	}
	return nil
}
