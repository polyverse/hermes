# Hermes

Hermes is a Status collecting, publishing, and reporting framework in complex supervision trees, especially those created by massive microservice deployments.

## Status as a Fundamental Monitoring Type

Status is the current "state" of a system with ZERO KNOWLEDGE of *WHY* the system is in that state, *HOW* it got there, or *WHAT* can make it change.

Status is a machine and human-understandable representation of the state a system/component is at any given time.

Status differs in content, intent and interpretation from logging and metrics. *Logs* are a stream of temporal activity that only makes sense when analyzed over a window. For instance, reading one single line of log doesn't tell you what the current state of the system that emitted it is. *Metrics* make sense as an aggregated macro snapshot of system behavior and performance. Metrics tell you what tends to happen or what has tended to happen in the past.

Status an important concept due to the meteoric rise of self-correcting large numbers of small servies. When developing or deploying services locally, for instance, you may not necessarily have a complex analytics engine to capture those logs and feedback behavior. When launching a quick and dirty demo you might not find aggregated logs over a complete "set" of services, and even when available, might not be easy to interpret.

