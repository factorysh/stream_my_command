package rfc7233

import (
	"fmt"
	"strconv"
	"strings"
)

type Range struct {
	min   int
	max   int
	start bool
	end   bool
}

func parseRange(raw string) (*Range, error) {
	r := strings.Split(raw, "-")
	if len(r) < 0 || len(r) > 2 {
		return nil, fmt.Errorf("Bad range length : %s", raw)
	}
	var err error
	rang := &Range{}
	if r[0] != "" {
		rang.min, err = strconv.Atoi(r[0])
		if err != nil {
			return nil, err
		}
		rang.start = true
	}
	if len(r) > 1 && r[1] != "" {
		rang.max, err = strconv.Atoi(r[1])
		if err != nil {
			return nil, err
		}
		rang.end = true
	}
	return rang, nil
}

type Ranges []*Range

func Parse(raw string) (Ranges, error) {
	if !strings.HasPrefix(raw, "bytes=") {
		return nil, fmt.Errorf("It doesn't start with bytes= : %s", raw)
	}
	r := strings.Split(raw[6:], ",")
	if len(r) == 0 {
		return nil, fmt.Errorf("Empty range %s", raw)
	}
	ranges := make(Ranges, len(r))
	for i := 0; i < len(r); i++ {
		var err error
		ranges[i], err = parseRange(r[i])
		if err != nil {
			return nil, err
		}
	}
	return ranges, nil
}
