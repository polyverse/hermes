package hermes

import "time"

const (
	CHILD_SEPARATOR = "/"
)

type ModelValue struct {
	Value string  `json:"value"`
	Age   float64 `json:"age"`
	TTL   float64 `json:"ttl"`
}

type Model map[string]*ModelValue

var (
	cachedModel Model
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
