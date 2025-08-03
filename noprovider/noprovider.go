package noprovider

type Provider struct{}

func (p *Provider) Authorized() bool { return true }
func (p *Provider) Start(onImageFunc func(u string)) {
	onImageFunc("http://127.0.0.1:9000/nts.png")
}
func (p *Provider) Stop() {}
