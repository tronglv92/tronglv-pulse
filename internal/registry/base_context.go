package registry

import "pulse/helper/utils/toolkit/downloader"

// BaseContext is the minimum interface embedded by all registry contexts.
type BaseContext interface {
	GetDownloader() downloader.Downloader
}

type baseContext struct{}

func newBaseContext() *baseContext { return &baseContext{} }

func (b *baseContext) GetDownloader() downloader.Downloader {
	return downloader.NewDownloader()
}
