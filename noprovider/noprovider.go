package noprovider

import (
	"embed"
	"io"
)

//go:embed nts.png
var embedFS embed.FS

type Provider struct{}

func (p *Provider) Authorized() bool { return true }
func (p *Provider) Start(onImageFunc func(io.Reader)) {
	file, err := embedFS.Open("nts.png")
	if err != nil {
		return
	}
	defer file.Close()
	onImageFunc(file)
}
func (p *Provider) Stop() {}
