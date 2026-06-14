package myLocale

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io/fs"
	"path"
	"strings"

	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
)

func loadContract(fsys fs.FS, contractsDir string) (*ErrorContract, error) {
	filePath := path.Join(contractsDir, "errors.yaml")
	data, err := fs.ReadFile(fsys, filePath)
	if err != nil {
		return nil, errors.Wrapf(err, "读取契约文件 %s 失败", filePath)
	}

	contract := &ErrorContract{}
	if err := yaml.Unmarshal(data, contract); err != nil {
		return nil, errors.Wrap(err, "解析契约文件失败")
	}
	if contract.AppCode == "" {
		return nil, fmt.Errorf("契约文件缺少 appCode")
	}

	for i := range contract.Codes {
		code := contract.Codes[i].Code
		prefix := contract.AppCode + "."
		if code != contract.AppCode && !strings.HasPrefix(code, prefix) {
			return nil, fmt.Errorf("错误码 %s 必须以 appCode %s 为前缀", code, contract.AppCode)
		}
	}

	return contract, nil
}

func loadMessages(fsys fs.FS, localesDir, locale string) (map[string]string, string, error) {
	filePath := path.Join(localesDir, locale, "errors.yaml")
	data, err := fs.ReadFile(fsys, filePath)
	if err != nil {
		return nil, "", errors.Wrapf(err, "读取语言文件 %s 失败", filePath)
	}

	messages := make(map[string]string)
	if err := yaml.Unmarshal(data, &messages); err != nil {
		return nil, "", errors.Wrap(err, "解析语言文件失败")
	}

	version := hashContent(data)
	return messages, version, nil
}

func listLocales(fsys fs.FS, localesDir string) ([]string, error) {
	entries, err := fs.ReadDir(fsys, localesDir)
	if err != nil {
		return nil, errors.Wrap(err, "读取 locales 目录失败")
	}

	locales := make([]string, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() {
			locales = append(locales, entry.Name())
		}
	}
	return locales, nil
}

func hashContent(data []byte) string {
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:8])
}
