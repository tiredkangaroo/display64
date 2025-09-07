package env

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
	"github.com/tiredkangaroo/display64/display"
)

type Environment struct {
	SpotifyClientID       string
	SpotifyClientSecret   string
	SpotifyRedirectURI    string
	SpotifyScopes         string
	DisplayServerHostport string
	CertFile              string
	KeyFile               string
	Debug                 bool

	DisplayConnection *display.Connection
}

var DefaultEnvironment = Environment{}

func Init() error {
	godotenv.Load()
	DefaultEnvironment = Environment{
		SpotifyClientID:       os.Getenv("SPOTIFY_CLIENT_ID"),
		SpotifyClientSecret:   os.Getenv("SPOTIFY_CLIENT_SECRET"),
		SpotifyRedirectURI:    dv(os.Getenv("SPOTIFY_REDIRECT_URI"), "https://127.0.0.1:9000/api/v1/spotify/redirect"),
		SpotifyScopes:         dv(os.Getenv("SPOTIFY_SCOPES"), "user-read-playback-state"),
		DisplayServerHostport: dv(os.Getenv("DISPLAY_SERVER_HOSTPORT"), "127.0.1:14366"),
		CertFile:              os.Getenv("CERT_FILE"),
		KeyFile:               os.Getenv("KEY_FILE"),
		Debug:                 os.Getenv("DEBUG") == "true",
	}
	DefaultEnvironment.DisplayConnection = &display.Connection{
		Hostport: DefaultEnvironment.DisplayServerHostport,
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
