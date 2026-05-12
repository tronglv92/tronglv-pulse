package filex

import (
	"pulse/helper/utils/toolkit/stringx"
	"fmt"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/utils"
	"net/url"
	"path"
	"path/filepath"
	"strings"
)

func RandFileName(fileName string) string {
	ext := filepath.Ext(fileName)
	return fmt.Sprintf(
		"%s-%s%s",
		stringx.Slugify(strings.TrimSuffix(fileName, ext)),
		utils.NewUuid(),
		ext,
	)
}

func GetFileName(rawUrl string) string {
	parts, err := url.Parse(rawUrl)
	if err != nil {
		logx.Errorf("Error parsing URL: %v", err)
		return ""
	}
	return path.Base(parts.Path)
}
