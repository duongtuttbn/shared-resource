package core

import (
	"context"
	"github.com/duongtuttbn/shared-resource/go-kit/dt"

	"github.com/gin-gonic/gin"
	"github.com/samber/lo"
)

const (
	LocaleHeader     = "X-Locale-Code"
	localeContextKey = "locale_code"
)

func UseLocale(supportedCodes []dt.LocaleCode, defaultLocale dt.LocaleCode) gin.HandlerFunc {
	return func(c *gin.Context) {
		locale := dt.LocaleCode(c.GetHeader(LocaleHeader))
		isValid := lo.Contains(supportedCodes, locale)
		if !isValid {
			locale = defaultLocale
		}
		c.Set(localeContextKey, locale)
		c.Next()
	}
}

func GetLocale(ctx context.Context) dt.LocaleCode {
	locale, ok := ctx.Value(localeContextKey).(dt.LocaleCode)
	if !ok {
		return ""
	}
	return locale
}
