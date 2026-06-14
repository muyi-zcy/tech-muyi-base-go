package myLocale

import (
	"io/fs"
	"sync"

	"github.com/pkg/errors"
)

type store struct {
	mu               sync.RWMutex
	appCode          string
	defaultLocale    string
	supportedLocales []string
	contract         *ErrorContract
	messages         map[string]map[string]string // locale -> code -> template
	versions         map[string]string            // locale -> hash
	httpHints        map[string]int
	fsys             fs.FS
	contractsDir     string
	localesDir       string
}

var defaultStore *store

// Options myLocale 初始化选项
type Options struct {
	AppCode          string
	DefaultLocale    string
	SupportedLocales []string
	FS               fs.FS
	ContractsDir     string
	LocalesDir       string
}

// Init 从 embed 或文件系统加载契约与文案
func Init(opts Options) error {
	if opts.FS == nil {
		return errors.New("myLocale.Init: FS 不能为空，请使用 //go:embed contracts locales")
	}
	if opts.AppCode == "" {
		return errors.New("myLocale.Init: AppCode 不能为空")
	}

	contractsDir := opts.ContractsDir
	if contractsDir == "" {
		contractsDir = "contracts"
	}
	localesDir := opts.LocalesDir
	if localesDir == "" {
		localesDir = "locales"
	}
	defaultLocale := opts.DefaultLocale
	if defaultLocale == "" {
		defaultLocale = "zh-CN"
	}

	platformContract, err := loadContract(platformFS, platformRoot+"/contracts")
	if err != nil {
		return errors.Wrap(err, "加载平台默认契约失败")
	}

	contract, err := loadContract(opts.FS, contractsDir)
	if err != nil {
		return err
	}
	if contract.AppCode != opts.AppCode {
		return errors.Errorf("契约 appCode=%s 与配置 appCode=%s 不一致", contract.AppCode, opts.AppCode)
	}

	supported, err := listLocales(opts.FS, localesDir)
	if err != nil {
		return err
	}

	platformLocales, err := listLocales(platformFS, platformRoot+"/locales")
	if err != nil {
		return err
	}
	if len(opts.SupportedLocales) > 0 {
		supported = opts.SupportedLocales
	}
	supported = mergeLocaleList(supported, platformLocales)
	if len(supported) == 0 {
		supported = []string{defaultLocale}
	}

	messages := make(map[string]map[string]string, len(supported))
	versions := make(map[string]string, len(supported))
	httpHints := mergeHTTPHints(buildHTTPHints(platformContract), buildHTTPHints(contract))

	serviceLocales, err := listLocales(opts.FS, localesDir)
	if err != nil {
		return err
	}

	for _, locale := range supported {
		platformMsgs := map[string]string{}
		if containsLocale(platformLocales, locale) {
			platformMsgs, _, err = loadMessages(platformFS, platformRoot+"/locales", locale)
			if err != nil {
				return err
			}
		}

		serviceMsgs := map[string]string{}
		versionSeed := locale
		if containsLocale(serviceLocales, locale) {
			serviceMsgs, versionSeed, err = loadMessages(opts.FS, localesDir, locale)
			if err != nil {
				return err
			}
		}

		messages[locale] = mergeMessages(platformMsgs, serviceMsgs)
		versions[locale] = hashContent([]byte(versionSeed + locale))
	}

	defaultStore = &store{
		appCode:          opts.AppCode,
		defaultLocale:    defaultLocale,
		supportedLocales: supported,
		contract:         contract,
		messages:         messages,
		versions:         versions,
		httpHints:        httpHints,
		fsys:             opts.FS,
		contractsDir:     contractsDir,
		localesDir:       localesDir,
	}
	return nil
}

func mergeLocaleList(primary, extra []string) []string {
	seen := make(map[string]struct{}, len(primary)+len(extra))
	result := make([]string, 0, len(primary)+len(extra))
	for _, locale := range primary {
		if _, ok := seen[locale]; ok {
			continue
		}
		seen[locale] = struct{}{}
		result = append(result, locale)
	}
	for _, locale := range extra {
		if _, ok := seen[locale]; ok {
			continue
		}
		seen[locale] = struct{}{}
		result = append(result, locale)
	}
	return result
}

func containsLocale(locales []string, locale string) bool {
	for _, item := range locales {
		if item == locale {
			return true
		}
	}
	return false
}

// GetContract 获取错误码契约
func GetContract() (*ErrorContract, error) {
	if defaultStore == nil {
		return nil, errors.New("myLocale 未初始化")
	}
	return defaultStore.contract, nil
}

// GetMessages 获取指定语言的文案包
func GetMessages(locale string) (*MessagesBundle, error) {
	if defaultStore == nil {
		return nil, errors.New("myLocale 未初始化")
	}
	return defaultStore.getMessages(locale)
}

// SupportedLocales 返回支持的语言列表
func SupportedLocales() []string {
	if defaultStore == nil {
		return nil
	}
	return append([]string(nil), defaultStore.supportedLocales...)
}

// AppCode 返回当前服务 appCode
func AppCode() string {
	if defaultStore == nil {
		return ""
	}
	return defaultStore.appCode
}

func (s *store) getMessages(locale string) (*MessagesBundle, error) {
	loc := locale
	if loc == "" {
		loc = s.defaultLocale
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	msgs, ok := s.messages[loc]
	if !ok {
		return nil, errors.Errorf("不支持的语言: %s", loc)
	}

	copied := make(map[string]string, len(msgs))
	for k, v := range msgs {
		copied[k] = v
	}

	return &MessagesBundle{
		AppCode:  s.appCode,
		Locale:   loc,
		Version:  s.versions[loc],
		Messages: copied,
	}, nil
}
