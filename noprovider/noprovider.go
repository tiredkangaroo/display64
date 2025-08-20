package noprovider

type Provider struct{}

func (p *Provider) Authorized() bool { return true }
func (p *Provider) Start(onImageFunc func(u string)) {
	onImageFunc("https://hc-cdn.hel1.your-objectstorage.com/s/v3/81ddba41db872a37a630dbb071f57ba4f916b019_image.png")
}
func (p *Provider) Stop() {}
