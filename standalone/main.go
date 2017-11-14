package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/polyverse-security/hermes"
)

var (
	serverAddr string

	parentUrl     string
	ourName       string
	parentFromEnv bool
	pushToParent  bool

	children string

	generateFake bool

	help      bool
	shortHelp bool
)

func main() {
	flag.StringVar(&serverAddr, "serveraddr", ":9091", "Sets the address at which to expose Hermes status reports")

	flag.StringVar(&parentUrl, "parenturl", "", "Sets the Parent URL to push to.")
	flag.StringVar(&ourName, "ourname", "", "Sets our name for the parent to get updates from.")
	flag.BoolVar(&parentFromEnv, "parentfromenv", false, fmt.Sprintf("Sets our parent parameters from Environment. You must set two environment variables %s and %s.", hermes.HERMES_PARENT_URL_ENV, hermes.HERMES_MYNAME_ENV))
	flag.BoolVar(&pushToParent, "pushtoparent", true, "Should we push each key update to the parent immediately?")

	flag.StringVar(&children, "children", "", "Set scraping children with the form: 'childname1=http://childurl1,childname2=http://childurl2'")

	flag.BoolVar(&generateFake, "generate_fake_keys", false, "Generate fake keys (mainly used for testing connectivity.)")
	flag.BoolVar(&help, "help", false, "Display help/usage")
	flag.BoolVar(&shortHelp, "?", false, "Display help/usage")

	flag.Parse()

	if help || shortHelp {
		printUsage()
		return
	}

	if parentFromEnv {
		hermes.SetParentFromEnv()
	} else if parentUrl != "" && ourName != "" {
		u, err := url.Parse(parentUrl)
		if err != nil {
			fmt.Printf("Unable to parse Parent URL %s, due to error: %s\n", parentUrl, err.Error())
			return
		}

		hermes.SetParent(ourName, *u)
	}

	hermes.SetDefaultPushToParent(pushToParent)

	go hermes.PushModelToParentForever(context.Background(), time.Duration(30)*time.Second)

	if generateFake {
		go generateHermesStatusUpdates()
	}

	if children != "" {
		childs := strings.Split(children, ",")
		for _, child := range childs {
			childparams := strings.Split(child, "=")
			u, err := url.Parse(childparams[1])
			if err != nil {
				fmt.Printf("Error when parsing child url %s: %s\n", childparams[1], err.Error())
				return
			}
			go hermes.ScrapeStatusForever(context.Background(), time.Duration(20)*time.Second, childparams[0], *u)
		}
	}

	fmt.Printf("Listening and serving hermes statuses at address: %s\n", serverAddr)
	err := http.ListenAndServe(serverAddr, hermes.GetHandler())
	fmt.Printf("Error when listening for Hermes updates: %s\n", err.Error())
}

func generateHermesStatusUpdates() {
	hermes.ReportStatus("Allowing Timeout", "This key will expire.", hermes.WithTTL(time.Duration(5)*time.Minute))

	for {
		hermes.ReportStatus("pushedToParentImmediately", fmt.Sprintf("It is now %s", time.Now().String()), hermes.WithPushToParent())
		hermes.ReportStatus("pushedToParentAtInterval", fmt.Sprintf("It is now %s", time.Now().String()))

		time.Sleep(time.Duration(5) * time.Second)
	}
}

func printUsage() {
	fmt.Println(
		`Usage: Acts as a Hermes standalone collector/relayer.

Hermes is a framework to publish status updates across Supervision trees. For more details, please visit
the github repository at: https://github.com/polyverse-security/hermes

This standalone app can be set as the parent from any Hermes-instrumented component, and can act
as the root/default collector where you can view all statuses. As a parent, this standalone app
can either have statuses pushed to it, or scrape child URLs.

Optionally, this component can have a parent set to which it can push/relay all keys itself.

Flags:
`)

	flag.PrintDefaults()
}
