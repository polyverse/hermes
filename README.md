# Hermes

Hermes is a Status collecting, publishing, and reporting framework in complex supervision trees, especially those created by massive microservice deployments.

## Getting Started

### Monitoring a Hermes component

#### Plain Text
For any Hermes-instrumented component, run:

<pre>
curl http://<component>:9091

status: pending
child1/status: launching
child1/message: Starting logger...
</pre>

#### JSON

JSON-formatted output provides a bit more metadata, such as a key's creation time, along with its age and prescribed TTL. All durations are in seconds, and all consumers must support floating point numbers. They are not required to honor more than the integer part (this rounding UP to the nearst second), but they *MUST NOT* fail when they encounter a decimal.

<pre>
curl http://<component>:9091?format=json
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

	hermes "github.com/polyverse-security/hermes/client-go"
)

var addr = flag.String("hermes-address", ":9091", "The address to listen on for Hermes HTTP requests.")

func main() {
	flag.Parse()
	http.Handle("/status", hermes.Handler)
	log.Fatal(http.ListenAndServe(*addr, nil))
}
</pre>

#### Reporting status

<pre>

package runner

import (
	hermes "github.com/polyverse-security/hermes/client-go"
)

func run() {
  hermes.Report(
    "runner.status", // Key
    "running",  // Value
    15) //TTL
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
status: "launching"
message: "Starting logger..."

### Child keys

The special separator slash (/) is used to indicate a child process's namespace.

status: "pending"
component1/status: "launching"
component1/message: "Starting logger..."
