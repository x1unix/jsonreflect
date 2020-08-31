package jsonx

import (
	"fmt"
	"regexp"
	"sort"
	"strconv"
)

type GroupedNumericKey struct {
	Order []int
	Key   string
}

func (g GroupedNumericKey) lessOf(gg GroupedNumericKey) bool {
	for i, x := range g.Order {
		if x < gg.Order[i] {
			return true
		}
	}
	return false
}

type GroupedNumbericKeys []GroupedNumericKey

func (gks GroupedNumbericKeys) Len() int {
	return len(gks)
}

func (gks GroupedNumbericKeys) Swap(i, j int) {
	gks[i], gks[j] = gks[j], gks[i]
}

func (gks GroupedNumbericKeys) Less(i, j int) bool {
	return gks[i].lessOf(gks[j])
}

// GroupNumericKeys groups set of keys with similar numeric prefix or suffix as array by pattern.
//
// Example of keys:
//	"fan1", "fan2", "fan1_2"
func (o Object) GroupNumericKeys(regex *regexp.Regexp, matchCount int) (GroupedNumbericKeys, error) {
	groupCount := matchCount + 1
	var out GroupedNumbericKeys
	for k := range o.Items {
		match := regex.FindStringSubmatch(k)
		if len(match) < groupCount {
			continue
		}

		segments := make([]int, 0, matchCount)
		for i := 1; i < groupCount; i++ {
			segment := match[i]
			if segment == "" {
				segments = append(segments, 0)
				continue
			}

			intVal, err := strconv.Atoi(segment)
			if err != nil {
				return nil, fmt.Errorf("segment %q of key %q is not a number (%w)", segment, k, err)
			}

			segments = append(segments, intVal)
		}

		out = append(out, GroupedNumericKey{Order: segments, Key: k})
	}

	sort.Sort(out)
	return out, nil
}
