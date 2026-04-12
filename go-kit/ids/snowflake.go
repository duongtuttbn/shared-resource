package ids

import (
	"fmt"
	"time"

	"github.com/godruoyi/go-snowflake"
)

const SnowflakeStartTime = 1288834974657

func NewSnowflake() string {
	return fmt.Sprintf("%d", NewSnowflakeNumber())
}

func NewSnowflakeNumber() uint64 {
	snowflake.SetStartTime(time.UnixMilli(SnowflakeStartTime))
	return snowflake.ID()
}
