package core

import (
	"io/fs"

	"github.com/muyi-zcy/tech-muyi-base-go/config"
	"github.com/muyi-zcy/tech-muyi-base-go/myLocale"
	"github.com/pkg/errors"
)

// LocaleOptionsFromEmbed 从 embed FS 构建 myLocale 选项
func LocaleOptionsFromEmbed(appCode string, localeFS fs.FS) myLocale.Options {
	opts := myLocale.Options{
		AppCode: appCode,
		FS:      localeFS,
	}
	if cfg := config.GetConfig(); cfg != nil {
		if opts.DefaultLocale == "" {
			opts.DefaultLocale = cfg.Locale.DefaultLocale
		}
		if len(opts.SupportedLocales) == 0 {
			opts.SupportedLocales = cfg.Locale.SupportedLocales
		}
	}
	return opts
}

// RegisterLocale 加载服务契约与文案，并注册 open API
func (s *Starter) RegisterLocale(opts myLocale.Options) error {
	if s.App.Config != nil {
		if opts.AppCode == "" {
			opts.AppCode = s.App.Config.AppCode
		}
		if opts.DefaultLocale == "" {
			opts.DefaultLocale = s.App.Config.Locale.DefaultLocale
		}
		if len(opts.SupportedLocales) == 0 {
			opts.SupportedLocales = s.App.Config.Locale.SupportedLocales
		}
	}
	if opts.AppCode == "" {
		return errors.New("RegisterLocale: AppCode 不能为空")
	}
	if err := myLocale.Init(opts); err != nil {
		return errors.Wrap(err, "初始化 myLocale 失败")
	}
	if s.App.Config == nil || s.App.Config.Locale.Enabled {
		myLocale.RegisterRoutes(s.GetAPIGroup())
	}
	return nil
}
