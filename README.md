# Hermes

Hermes is a Status collecting, publishing, and reporting framework in complex supervision trees, especially those created by massive microservice deployments.

## "Status" as a Fundamental Monitoring Channel for Reactive Applications

Status is the current "state" of a system with ZERO KNOWLEDGE of *WHY* the system is in that state, *HOW* it got there, or *WHAT* can make it change.

Status is a machine and human-understandable representation of the state a system/component is at any given time, for the intent and purpose of reacting to it. Traditional applications have two main channels for monitoring their state - logging, and some kind of aggregate metrics. However, when writing controllers or supervisors that operate on those applications, these controllers need to live outside the fabric of the apps, and must be connected through elaborate and complex channels to the logs/metrics destination. This means that supervision trees are difficult to set up.

Hermes provides an infrastructure-neutral status reporting channel, and more so, a consistent standard to do it. This means that when you write controllers or supervisors to launch more components (such as workers, nodes, etc.), those controllers can use Hermes to consistently monitor the state of those workers, and the controllers' controllers can monitor the tree.

## Data Schema

All State is stored as Key-Value pairs of strings. Each key must meet the following contract:
1. It must not begin or end with a "."
2. It may contain any letters (upper or lowercase), digits, underscore (\_) and dot (\.)
3. The regex for it is: <pre>^[\w][\w\.]*[\w]$</pre>

1. The last time the key was updated.
2. The TTL of the key (if not updated within that TTL again, the key is dropped and no longer honored.)

