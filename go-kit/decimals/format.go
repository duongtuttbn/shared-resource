package decimals

import (
	"strconv"

	"github.com/dustin/go-humanize"
)

func FormatFloat(value float64) string {
	switch {
	case value < 0.00001:
		return "﹤0.00001"
	case value < 1:
		return humanize.FormatFloat("#.#####", value)
	case value < 10:
		return humanize.FormatFloat("#.####", value)
	case value < 100:
		return humanize.FormatFloat("#.###", value)
	default:
		return humanize.FormatFloat("#,###.##", value)
	}
}

func ToFloat(num any) float64 {
	switch v := num.(type) {
	case float64:
		return v
	case float32:
		return float64(v)
	case int64:
		return float64(v)
	case int32:
		return float64(v)
	case int:
		return float64(v)
	case uint64:
		return float64(v)
	case uint32:
		return float64(v)
	case uint:
		return float64(v)
	case string:
		f, _ := strconv.ParseFloat(v, 64)
		return f
	}

	return 0
}
