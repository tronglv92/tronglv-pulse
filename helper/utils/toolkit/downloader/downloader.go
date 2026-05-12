package downloader

import (
	"pulse/helper/utils/toolkit/filex"
	"fmt"
	"github.com/zeromicro/go-zero/core/logx"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
)

const StorageDirectory = "storage"

type (
	Downloader interface {
		Download(urlFile string) (filex.FileEntity, error)
	}
)

type downloader struct {
	storageDir string
	httpClient *http.Client
}

func NewDownloader() Downloader {
	return New(
		&http.Client{
			CheckRedirect: func(r *http.Request, via []*http.Request) error {
				r.URL.Opaque = r.URL.Path
				return nil
			},
		},
		StorageDirectory,
	)
}

func New(client *http.Client, storageDir string) Downloader {
	r := &downloader{
		storageDir: storageDir,
		httpClient: client,
	}

	if err := os.MkdirAll(r.storageDir, os.ModePerm); err != nil {
		logx.Must(err)
	}
	return r
}

func (d *downloader) fileName(urlFile string) string {
	fileURL, err := url.Parse(urlFile)
	if err != nil {
		log.Fatal(err)
	}
	path := fileURL.Path
	segments := strings.Split(path, "/")

	return segments[len(segments)-1]
}

func (d *downloader) Download(urlFile string) (filex.FileEntity, error) {
	filePath := filepath.Join(d.storageDir, d.fileName(urlFile))
	file, err := os.Create(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to create file: %w", err)
	}
	defer func() {
		_ = file.Close()
	}()

	resp, err := d.httpClient.Get(urlFile)
	if err != nil {
		return nil, fmt.Errorf("failed to download file: %w", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("bad status: %s", resp.Status)
	}

	if _, err = io.Copy(file, resp.Body); err != nil {
		return nil, fmt.Errorf("failed to save file: %w", err)
	}

	return filex.NewFile(filePath, nil), nil
}

func (d *downloader) Remove(filePath string) error {
	if err := os.Remove(filePath); err != nil {
		return err
	}
	return nil
}
