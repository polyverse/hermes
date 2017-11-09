package hermes

import (
	"strings"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

type internalStatus struct {
	value     string
	ttl       time.Duration
	createdAt time.Time
	expiresAt time.Time
}

var (
	lock        sync.RWMutex
	gcInterval  time.Duration
	statusStore map[string]internalStatus
	logger      *logrus.Logger
	changeIndex int64
)

func init() {
	lock = sync.RWMutex{}
	gcInterval = time.Duration(30) * time.Second
	statusStore = make(map[string]internalStatus)
	logger = logrus.StandardLogger()

	go runPeriodicGc()
}

func SetGcInterval(interval time.Duration) {
	gcInterval = interval
}

func SetLogger(l *logrus.Logger) {
	logger = l
}

// ChangeIndex is a monotonically increasing integer
// that is incremented by one, for each change that happens.
// This index is useful in caching operations, to check whether
// a change has happened, before acting on it.
func GetChangeIndex() int64 {
	lock.RLock()
	defer lock.RUnlock()
	return changeIndex
}

func RunGC() {
	logger.Debugf("Hermes: Running GC cycle.")
	lock.Lock()
	defer lock.Unlock()

	changed := false
	for key, s := range statusStore {
		if s.expiresAt.After(time.Now()) {
			if logger.Level >= logrus.DebugLevel {
				logger.Debugf("Hermes: Key %s was created at %s with TTL %s (making expiry %s). It is now expired since the time is %s. Purging it.", key, s.createdAt.String(), s.ttl.String(), s.expiresAt.String(), time.Now().String())
			}
			delete(statusStore, key)
			changed = true
		}
	}

	if changed {
		changeIndex++
	}
}

func runPeriodicGc() {
	logger.Debugf("Hermes: Started background GC.")

	for {
		logger.Debugf("Hermes: Sleeping for %s before next GC.", gcInterval.String())
		time.Sleep(gcInterval)

		RunGC()
	}
}

func putStatus(key string, value string, ttl time.Duration) int64 {
	if ttl <= 0 {
		logger.Debugf("Hermes: Attempted to set key %s with a zero or negative TTL. Not setting it.", key)
		return changeIndex
	}

	lock.Lock()
	defer lock.Unlock()
	changeIndex++
	logger.Debugf("Hermes: storing key %s", key)

	creationTime := time.Now()
	s := internalStatus{
		createdAt: creationTime,
		ttl:       ttl,
		expiresAt: creationTime.Add(ttl),
		value:     value,
	}

	statusStore[key] = s
	return changeIndex
}

func getStatus(key string) (string, bool, int64) {
	lock.RLock()
	defer lock.RUnlock()
	value, ok := statusStore[key]
	return value.value, ok, changeIndex
}

func deleteStatus(key string) int64 {
	lock.Lock()
	defer lock.Unlock()
	changeIndex++
	logger.Debugf("Hermes: deleting key %s", key)
	delete(statusStore, key)
	return changeIndex
}

func listKeys(prefix string) ([]string, int64) {
	lock.RLock()
	defer lock.RUnlock()

	keys := make([]string, 0, len(statusStore))
	for key, _ := range statusStore {
		if strings.HasPrefix(key, prefix) {
			keys = append(keys, key)
		}
	}
	return keys, changeIndex
}

func shallowCopyStore() (map[string]internalStatus, int64) {
	lock.RLock()
	defer lock.RUnlock()

	sc := make(map[string]internalStatus, len(statusStore))

	for key, value := range statusStore {
		sc[key] = value
	}

	return sc, changeIndex
}
