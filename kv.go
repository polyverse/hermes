package hermes

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/pkg/errors"
)

const (
	HERMES_MYNAME_ENV     = "HERMES_MYNAME"
	HERMES_PARENT_URL_ENV = "HERMES_PARENT_URL"
)

var (
	defaultttl          time.Duration = time.Duration(10) * time.Minute
	defaultPushToParent               = true

	hasParent bool = false
	ourName   string
	parentUrl *url.URL

	NoEnvFoundErr  = fmt.Errorf("No Hermes Environment settings found. Must set %s and %s.", HERMES_MYNAME_ENV, HERMES_PARENT_URL_ENV)
	NoParentSetErr = fmt.Errorf("No Hermes Parent set. Please call SetParent or SetParentFromEnv methods to set a parent before pushing a model to it.")
)

// Sets the Default TTL for each key when Status is reporting without
// a TTL.
func SetDefaultTTL(ttl time.Duration) {
	defaultttl = ttl
}

func SetDefaultPushToParent(pushToParent bool) {
	defaultPushToParent = pushToParent
}

// Sets the parent to report status to.
// The first parameter is "myname" which is the name of the current process
// as the parent should see it.
// The second parameter is the URL where the parent may be found (where to POST).
func SetParent(myname string, url *url.URL) {
	hasParent = true
	ourName = myname
	parentUrl = url
}

// This sets the Parent from ENV variables.
// This allows for instrumented process to not have a lot of
// glue logic when initializing.
// We are looking for two variables here:
// HERMES_MYNAME - sets the name of the current process
// HERMES_PARENT_URL - sets the URL to which to POST for the parent.
func SetParentFromEnv() error {
	myname, ok := os.LookupEnv(HERMES_MYNAME_ENV)
	if !ok || myname == "" {
		return NoEnvFoundErr
	}

	parenturlstr, ok := os.LookupEnv(HERMES_PARENT_URL_ENV)
	if !ok || parenturlstr == "" {
		return NoEnvFoundErr
	}

	parentUrl, err := url.Parse(parenturlstr)
	if err != nil {
		return errors.Wrapf(err, "Unable to parse %s value %s into a URL.", HERMES_PARENT_URL_ENV, parenturlstr)
	}

	SetParent(myname, parentUrl)

	return nil
}

type opts struct {
	ttl          time.Duration
	pushToParent bool
}

type StatusOpt func(o *opts)

func WithTTL(ttl time.Duration) StatusOpt {
	return func(o *opts) {
		o.ttl = ttl
	}
}

func WithPushToParent() StatusOpt {
	return func(o *opts) {
		o.pushToParent = true
	}
}

func ReportStatus(key string, value string, optlist ...StatusOpt) {
	opt := &opts{
		ttl:          defaultttl,
		pushToParent: defaultPushToParent,
	}

	for _, o := range optlist {
		o(opt)
	}

	putStatus(key, value, opt.ttl)

	if opt.pushToParent {
		go pushKeyToParent(key, value, opt.ttl)
	}
}

func GetStatus(key string) (string, bool) {
	val, exists, _ := getStatus(key)
	return val, exists
}

func GetStatusKeys() []string {
	keys, _ := listKeys("")
	return keys
}

func GetStatusKeysWithPrefix(prefix string) []string {
	keys, _ := listKeys(prefix)
	return keys
}

func GetStatusesWithPrefix(prefix string) map[string]string {
	keys, _ := listKeys(prefix)
	statuses := make(map[string]string, len(keys))

	for _, key := range keys {
		statuses[key], _, _ = getStatus(key)
	}

	return statuses
}

func pushKeyToParent(key string, value string, ttl time.Duration) {
	if !hasParent {
		return
	}

	var m Model
	m[key] = &ModelValue{
		Value: value,
		TTL:   ttl.Seconds(),
		Age:   0,
	}

	pushModelToParent(m)
}

func pushModelToParent(m Model) {
	if !hasParent {
		return
	}

	jstr, err := json.Marshal(m)
	if err != nil {
		logger.WithError(err).Errorf("Hermes: Unable to serialize Model %v into JSON.", m)
		return
	}

	resp, err := http.DefaultClient.Post(parentUrl.String(), CONTENT_TYPE_JSON, bytes.NewReader(jstr))
	if err != nil {
		logger.WithError(err).Errorf("Hermes: Error when writing to parent: %s", parentUrl.String())
	}

	if resp.StatusCode != http.StatusOK {
		logger.Warningf("Hermes: Non-Ok Status Code %d when writing to parent: %s", resp.StatusCode, parentUrl.String())
	}
}

func PushModelToParent() error {
	if !hasParent {
		return NoParentSetErr
	}

	m := generateModel()
	pushModelToParent(m)

	return nil
}

func PushModelToParentForever(ctx context.Context, interval time.Duration) error {
	if !hasParent {
		return NoParentSetErr
	}

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			PushModelToParent()
		}
	}
}

// Scrapes a Hermes endpoint, and pulls its keys
// as child keys in the current model
func ScrapeStatus(childName string, url url.URL) error {
	resp, err := http.DefaultClient.Get(url.String())
	if err != nil {
		logger.WithError(err).Errorf("Hermes: Error when scraping child %s at URL %s.", childName, url.String())
		return err
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		logger.WithError(err).Errorf("Hermes: Error when reading response body for child %s at URL %s.", childName, url.String())
		return err
	}

	var m Model
	err = json.Unmarshal(body, &m)
	if err != nil {
		logger.WithError(err).Errorf("Hermes: Error when JSON-parsing body for child %s at URL %s.", childName, url.String())
		return err
	}

	insertModel(childName, m)

	return nil
}

func ScrapeStatusForever(ctx context.Context, interval time.Duration, childName string, url url.URL) error {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			ScrapeStatus(childName, url)
		}
	}
}
