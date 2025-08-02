package env

import (
	"os"

	"github.com/joho/godotenv"
)

type Environment struct {
	SpotifyClientID     string
	SpotifyClientSecret string
	SpotifyRedirectURI  string
	SpotifyScopes       string
}

var DefaultEnvironment = Environment{}

func Init() {
	godotenv.Load()
	DefaultEnvironment = Environment{
		SpotifyClientID:     os.Getenv("SPOTIFY_CLIENT_ID"),
		SpotifyClientSecret: os.Getenv("SPOTIFY_CLIENT_SECRET"),
		SpotifyRedirectURI:  dv(os.Getenv("SPOTIFY_REDIRECT_URI"), "http://127.0.0.1:9000/spotify/redirect"),
		SpotifyScopes:       dv(os.Getenv("SPOTIFY_SCOPES"), "user-read-playback-state"),
	}
}

func dv(a, b string) string {
	if a == "" {
		return b
	}
	return a
}
