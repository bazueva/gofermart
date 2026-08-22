package helpers

import (
	"fmt"
	"math"
)

func SafeInt64ToInt32(v int64) (int32, error) {
	if v > math.MaxInt32 {
		return 0, fmt.Errorf("value %d exceeds int32 max limit %d", v, math.MaxInt32)
	}
	if v < math.MinInt32 {
		return 0, fmt.Errorf("value %d exceeds int32 min limit %d", v, math.MinInt32)
	}
	return int32(v), nil
}
