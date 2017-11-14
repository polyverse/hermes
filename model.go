package hermes

import (
	"sort"
	"time"
	"strings"
)

const (
	CHILD_SEPARATOR = "/"
)

type ModelValue struct {
	Value string  `json:"value"`
	Age   float64 `json:"age"`
	TTL   float64 `json:"ttl"`
}

type ModelKeys []string

func (p ModelKeys) Len() int      { return len(p) }
func (p ModelKeys) Swap(i, j int) { p[i], p[j] = p[j], p[i] }
func (p ModelKeys) Less(i, j int) bool {

	// TODO: Fill in a better ordering function.
	s1 := p[i]
	s2 := p[j]

	s1slash := strings.Contains(s1, "/")
	s2slash := strings.Contains(s2, "/")

	if s1slash == s2slash {
		//either both contain slashes, or neither contains slashes
		return s1 < s2
	} else if s2slash && !s1slash {
		//s2 has a slash, but not s1. s1 comes first.
		return true
	}
	return false
}

// Sort is a convenience method.
func (p ModelKeys) Sort() { sort.Sort(p) }

type Model map[string]*ModelValue

func (m Model) SortedKeys() []string {
	keys := ModelKeys(make([]string, 0, len(m)))
	for key, _ := range m {
		keys = append(keys, key)
	}

	keys.Sort()

	return keys
}

var (
	cachedModel       Model
	cachedChangeIndex int64
)

func generateModel(prefix string, cached bool) Model {

	if cached == false || cachedModel == nil || cachedChangeIndex < GetChangeIndex() {
		sc, ci := shallowCopyStore()

		rm := make(map[string]*ModelValue, len(sc))
		for key, value := range sc {
			rm[prefix+key] = &ModelValue{
				Value: value.value,
				TTL:   value.ttl.Seconds(),
				Age:   time.Since(value.createdAt).Seconds(),
			}
		}

		cachedChangeIndex = ci
		cachedModel = rm
	}

	return cachedModel
}

func insertModel(childName string, m Model) {
	childPrefix := ""
	if childName != "" {
		childPrefix = childName + CHILD_SEPARATOR
	}

	for key, val := range m {
		// Go through official ReportStatus API so pushes to parent happen
		ReportStatus(childPrefix+key, val.Value, WithTTL(time.Duration(val.TTL-val.Age)*time.Second))
	}
}

func newEmptyModel() Model {
	return make(map[string]*ModelValue)
}
