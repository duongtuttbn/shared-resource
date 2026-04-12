package dt

type (
	LocaleCode string
	Locale     struct {
		Code LocaleCode `json:"code"`
		Name string     `json:"name"`
	}
)

const (
	DefaultLocale LocaleCode = "en_us"
)
