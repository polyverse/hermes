# DEPRECATION NOTICE

Please note that this repository has been deprecated and is no longer actively maintained by Polyverse Corporation.  It may be removed in the future, but for now remains public for the benefit of any users.

Importantly, as the repository has not been maintained, it may contain unpatched security issues and other critical issues.  Use at your own risk.

While it is not maintained, we would graciously consider any pull requests in accordance with our Individual Contributor License Agreement.  https://github.com/polyverse/contributor-license-agreement

For any other issues, please feel free to contact info@polyverse.com

---

# Hermes

Hermes is a Status collecting, publishing, and reporting framework in complex supervision trees, especially those created by massive microservice deployments.

The main challenge with monitoring fan-out services (services that start other services) is fundamentally the need for external infrastructure. You must either have a log sink or a metrics sink set up previously. You must then connect your "supervisor" (the service launching other services) to this sink to read status of what's going on. Alternatively, you could reach out to the fabric on which you launch those services (such as Kubernetes), in which case the best visibility you can expect is "Running", "Stopped", or "Not running."

This leads to a lot of side-channel coding and cumbersome workarounds. They lead to vendor-locking, but more so, they lead to infra-lockin! Infra-lockin is when you have to set pre-determined endpoints for your AWS CloudWatch region and so on and so forth. When running on localhost, you have to use other tricks. It's just a nightmare.

Hermes was inspired by Prometheus's simplicit of scrapable endpoints exposing metrics. Hermes exposes state of a system in a similar fashion.

This means when you instrument a component, it exposes it's Hermes status directly which you can query. If it is executed by a parent, the parent is able to receive updates from the child (and others) seamlessly exposing a consolidated status of the tree below it. If the parent is ever executed under a larger parent, that parent seamlessly gains access to the status tree.

For reactive components that are monitoring, changing, adapting, this ability does away with the need to go poll logs from children, or to do creative querying, etc. View the complete state of your system from your supervisor.

## Getting Started

### Monitoring a Hermes component

#### Plain Text
For any Hermes-instrumented component, run:

<pre>
curl http://localhost:9091

status: pending
child1/status: launching
child1/message: Starting logger...
</pre>

#### JSON

JSON-formatted output provides a bit more metadata, such as a key's creation time, along with its age and prescribed TTL. All durations are in seconds, and all consumers must support floating point numbers. They are not required to honor more than the integer part (this rounding UP to the nearst second), but they *MUST NOT* fail when they encounter a decimal.

<pre>
curl http://localhost:9091?type=json
{
  "status": {
    "age": 12,
    "ttl": 15,
    "value": "pending"
    },
  "child1/status": {
    "age": 1,
    "ttl": 30,
    "value": "launching"
  },
  "child1/message": {
    "age": 5,
    "ttl": 10,
    "value": "Starting logger..."
  }
}
</pre>

### Instrumenting with Hermes

All client bindings are only supported for Golang at the moment. We welcome contributions for bindings to other languages/platforms/frameworks.

#### Exposing Status

All Hermes components are required to expose direct status in a scrapable endpoint. Much like Prometheus, if pushing of status updates is required, this may be done using external/additional processes.

<pre>

package main

import (
	"flag"
	"log"
	"net/http"

	hermes "github.com/polyverse/hermes"
)

var addr = flag.String("hermes-address", ":9091", "The address to listen on for Hermes HTTP requests.")

func main() {
	flag.Parse()
	http.Handle("/status", hermes.GetHandler())
	log.Fatal(http.ListenAndServe(*addr, nil))
}
</pre>

#### Reporting status

<pre>

package runner

import (
	hermes "github.com/polyverse/hermes"
)

func run() {
  hermes.ReportStatus(
    "runner.status", // Key
    "running")  // Value
}
</pre>

# Purpose, Design, and implementation

## "Status" as a Fundamental Monitoring Channel for Reactive Applications

Status is the current "state" of a system with ZERO KNOWLEDGE of *WHY* the system is in that state, *HOW* it got there, or *WHAT* can make it change.

Status is a machine and human-understandable representation of the state a system/component is at any given time, for the intent and purpose of reacting to it. Traditional applications have two main channels for monitoring their state - logging, and some kind of aggregate metrics. However, when writing controllers or supervisors that operate on those applications, these controllers need to live outside the fabric of the apps, and must be connected through elaborate and complex channels to the logs/metrics destination. This means that supervision trees are difficult to set up.

Hermes provides an infrastructure-neutral status reporting channel, and more so, a consistent standard to do it. This means that when you write controllers or supervisors to launch more components (such as workers, nodes, etc.), those controllers can use Hermes to consistently monitor the state of those workers, and the controllers' controllers can monitor the tree.

## Data Scheme

The data obtained from each component is an aggregation of the status keys of itself, along with the collection of status keys obtained from its children, forming a tree.

As an observer, you are required only to ask the supervisor, and you should get the complete state of the tree that supervisor is supervising - either directly, or through proxy supervisors running as its children. All keys may have additional metadata such as *ttl* (time to live).

### Proc Keys
All State is stored as Key-Value pairs of strings. Each key must meet the following contract:
1. It must not begin or end with a "."
2. It may contain any letters (upper or lowercase), digits, underscore (\_) and dot (\.)
3. The regex for it is: <pre>^[\w][\w\.]*[\w]$</pre>

It is suggested that dots (.) be used to separate subsystems or namespaces within a process.

*Examples:*
<pre>
status: "launching"
message: "Starting logger..."
</pre>

### Child keys

The special separator slash (/) is used to indicate a child process's namespace.

*Examples:*
<pre>
status: "pending"
component1/status: "launching"
component1/message: "Starting logger..."
</pre>
