# kube-probe-bench

Idea: some people have a hard time deciding what to use for liveness probe startup times and resource limits. If you can recreate the situation locally, this can potentially plot the startup time vs the resource requirements. May not work very well when there are external dependencies that add to startup time.

### TODO:
- use kubelet http client
- pass ports/urls/actual probe specs
- make sure to pull images before starting containers
- etc...
