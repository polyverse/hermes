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

func generateModel(prefix string) Model {

	sc, _ := shallowCopyStore()

	rm := make(map[string]*ModelValue, len(sc))
	for key, value := range sc {
		rm[prefix+key] = &ModelValue{
			Value: value.value,
			TTL:   value.ttl.Seconds(),
			Age:   time.Since(value.createdAt).Seconds(),
		}
	}

	return rm
}

func insertModel(childName string, m Model) {
	childPrefix := ""
	if childName != "" {
		childPrefix = childName + CHILD_SEPARATOR
	}

	for key, val := range m {
		putStatus(childPrefix+key, val.Value, time.Duration(val.TTL-val.Age)*time.Second)
	}
}

func newEmptyModel() Model {
	return make(map[string]*ModelValue)
}
