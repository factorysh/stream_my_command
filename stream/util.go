package stream

import (
	"fmt"
	"math"
	"path"
)

func BucketPath(home string, n int) string {
	return path.Join(home, fmt.Sprintf("bucket_%d", n))
}

func min(a, b int) int {
	if a > b {
		return b
	}
	return a
}

func div(x, y int) int {
	if x < y { // Early optimization
		return 0
	}
	return int(math.Floor(float64(x) / float64(y)))
}
