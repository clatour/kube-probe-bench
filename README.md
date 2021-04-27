# kube-probe-bench

Idea: some people have a hard time deciding what to use for liveness probe startup times and resource limits. If you can recreate the situation locally, this can potentially plot the startup time vs the resource requirements. May not work very well when there are external dependencies that add to startup time.

```bash
$ go run main.go nginx:latest 80 /
2021/04/27 12:44:14 ensuring image 'nginx:latest' is pulled
2021/04/27 12:44:14 container starting
2021/04/27 12:44:15 /heuristic_moore (500m cpu): starting probe against 'http://0.0.0.0:49244/'
2021/04/27 12:44:16 /heuristic_moore (500m cpu): success after '1.486137439s'
2021/04/27 12:44:16 container starting
2021/04/27 12:44:17 /silly_joliot (1000m cpu): starting probe against 'http://0.0.0.0:49245/'
2021/04/27 12:44:18 /silly_joliot (1000m cpu): success after '1.515524651s'
2021/04/27 12:44:18 container starting
2021/04/27 12:44:19 /distracted_solomon (1500m cpu): starting probe against 'http://0.0.0.0:49246/'
2021/04/27 12:44:20 /distracted_solomon (1500m cpu): success after '1.523629582s'
2021/04/27 12:44:20 container starting
2021/04/27 12:44:21 /awesome_tharp (2000m cpu): starting probe against 'http://0.0.0.0:49247/'
2021/04/27 12:44:22 /awesome_tharp (2000m cpu): success after '1.497264121s'
2021/04/27 12:44:22 container starting
2021/04/27 12:44:23 /angry_keldysh (4000m cpu): starting probe against 'http://0.0.0.0:49248/'
2021/04/27 12:44:24 /angry_keldysh (4000m cpu): success after '1.480436569s'
```

### TODO:
- ~use kubelet http client~ this requires kubernetes internal packages
- ~pass ports/urls/actual probe specs~ dumb `os.Args`-based for now
- ~make sure to pull images before starting containers~
- etc...
