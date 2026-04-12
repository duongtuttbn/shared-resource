package scope

import "gorm.io/gorm"

var (
	DeletedColumnName = "deleted"
	HiddenColumnName  = "hidden"
)

func NotDeleted(table ...string) func(db *gorm.DB) *gorm.DB {
	field := DeletedColumnName
	if len(table) > 0 {
		field = table[0] + "." + field
	}

	return FilterBoolean(field, false)
}

func NotHidden(table ...string) func(db *gorm.DB) *gorm.DB {
	field := HiddenColumnName
	if len(table) > 0 {
		field = table[0] + "." + field
	}
	return FilterBoolean(field, false)
}

func FilterBoolean(field string, value bool) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where(field+" = ?", value)
	}
}
