package request

import (
	"pulse/helper/utils/errors"
	"pulse/helper/utils/toolkit/filex"
	"fmt"
	"github.com/zeromicro/go-zero/core/logx"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

var (
	MaxFileSize       int64 = 20 << 20
	MaxFileNumber     int   = 3
	StorageFolderName       = "storage"
	AllowExts               = []string{".jpg", ".jpeg", ".png", ".doc", ".docx", ".xls", ".xlsx", ".csv", ".pdf"}
)

type (
	UploadFile interface {
		GetFiles(key string) ([]filex.FileEntity, error)
		SaveFiles(key string) ([]*UploadFileInfo, error)
	}

	UploadFileInfo struct {
		FileName string
		FilePath string
	}

	UploadOption struct {
		MaxFileNumber    int
		MaxFileSize      int64
		FolderName       string
		AllowExts        []string
		OriginalFilename bool
		StrictMode       bool
	}

	fileUploadSvc struct {
		*UploadOption
		req *http.Request
	}

	ReqUploadOption func(s *UploadOption)
)

func WithMaxFileNumber(number int) ReqUploadOption {
	return func(m *UploadOption) {
		m.MaxFileNumber = number
	}
}

func WithOriginalFilename(s bool) ReqUploadOption {
	return func(m *UploadOption) {
		m.OriginalFilename = s
	}
}

func WithMaxFileSize(size int64) ReqUploadOption {
	return func(m *UploadOption) {
		m.MaxFileSize = size
	}
}

func WithFolder(folder string) ReqUploadOption {
	return func(m *UploadOption) {
		m.FolderName = filepath.Join(m.FolderName, folder)
	}
}

func WithExtensions(exts []string) ReqUploadOption {
	return func(m *UploadOption) {
		m.AllowExts = exts
	}
}

func WithStrictMode() ReqUploadOption {
	return func(m *UploadOption) {
		m.StrictMode = true
	}
}

func NewFileUpload(req *http.Request, opts ...ReqUploadOption) UploadFile {
	r := &fileUploadSvc{
		req: req,
		UploadOption: &UploadOption{
			MaxFileNumber: MaxFileNumber,
			MaxFileSize:   MaxFileSize,
			FolderName:    StorageFolderName,
			AllowExts:     AllowExts,
		},
	}
	for _, opt := range opts {
		opt(r.UploadOption)
	}
	return r
}

func (f *fileUploadSvc) mkdir() error {
	if _, err := os.Stat(f.FolderName); errors.Is(err, os.ErrNotExist) {
		if err = os.MkdirAll(f.FolderName, os.ModePerm); err != nil {
			return err
		}
	}
	return nil
}

func (f *fileUploadSvc) allowExts(fileName string) bool {
	ext := strings.ToLower(filepath.Ext(fileName))
	for _, validExt := range f.AllowExts {
		if strings.ToLower(ext) == validExt {
			return true
		}
	}
	return false
}

func (f *fileUploadSvc) GetFiles(key string) ([]filex.FileEntity, error) {
	err := f.req.ParseMultipartForm(int64(f.MaxFileNumber))
	if err != nil {
		return nil, err
	}

	files, ok := f.req.MultipartForm.File[key]
	if !ok {
		return nil, fmt.Errorf("the key name %s does not exists", key)
	}
	if len(files) > f.MaxFileNumber {
		return nil, fmt.Errorf("number of files exceeds the limit (%d)", f.MaxFileNumber)
	}

	var items []filex.FileEntity
	for _, val := range files {
		if !f.allowExts(val.Filename) {
			if f.StrictMode {
				return nil, fmt.Errorf("the file extension is not valid (%s)", val.Filename)
			}
			continue
		}

		flo, err := val.Open()
		if err != nil {
			logx.Error(err)
			continue
		}

		fileData, e := io.ReadAll(flo)
		if e != nil {
			_ = flo.Close()
			continue
		}
		_ = flo.Close()

		var fileName = val.Filename
		if !f.OriginalFilename {
			fileName = filex.RandFileName(fileName)
		}
		items = append(items, filex.NewFile(fileName, fileData))
	}
	return items, nil
}

func (f *fileUploadSvc) SaveFiles(key string) ([]*UploadFileInfo, error) {
	files, err := f.GetFiles(key)
	if err != nil {
		return nil, err
	}

	if err = f.mkdir(); err != nil {
		return nil, err
	}
	var results []*UploadFileInfo
	for _, file := range files {
		fPath := filepath.Join(f.FolderName, file.GetName())
		tpf, err := os.Create(fPath)
		if err != nil {
			_ = tpf.Close()
			return nil, fmt.Errorf("error creating output file: %v", err)
		}

		_, err = tpf.Write(file.GetData())
		if err != nil {
			_ = tpf.Close()
			return nil, fmt.Errorf("error copying file: %v", err)
		}
		file.FlushData()
		_ = tpf.Close()

		results = append(results, &UploadFileInfo{
			FileName: file.GetName(),
			FilePath: fPath,
		})
	}
	return results, nil
}
