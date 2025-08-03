package env

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

type Environment struct {
	SpotifyClientID     string
	SpotifyClientSecret string
	SpotifyRedirectURI  string
	SpotifyScopes       string
	DisplayServerURL    string
	Debug               bool
}

var DefaultEnvironment = Environment{}

func Init() error {
	godotenv.Load()
	DefaultEnvironment = Environment{
		SpotifyClientID:     os.Getenv("SPOTIFY_CLIENT_ID"),
		SpotifyClientSecret: os.Getenv("SPOTIFY_CLIENT_SECRET"),
		SpotifyRedirectURI:  dv(os.Getenv("SPOTIFY_REDIRECT_URI"), "http://127.0.0.1:9000/api/v1/spotify/redirect"),
		SpotifyScopes:       dv(os.Getenv("SPOTIFY_SCOPES"), "user-read-playback-state"),
		DisplayServerURL:    dv(os.Getenv("DISPLAY_SERVER_URL"), "http://127.0.1:14366"),
		Debug:               os.Getenv("DEBUG") == "true",
	}
	if DefaultEnvironment.SpotifyClientID == "" ||
		DefaultEnvironment.SpotifyClientSecret == "" {
		return fmt.Errorf("missing required environment variables: SPOTIFY_CLIENT_ID and SPOTIFY_CLIENT_SECRET")
	}
	return nil
}

func dv(a, b string) string {
	if a == "" {
		return b
	}
	return a
}
