# resolvr
simple wildcard dns server that will extract ip addresses from the hostname and resolve to them

similar to xip.io which appears to have gone offline with the recent basecamp exodus

```bash
# say you have an internal ip 10.10.10.1
> dig +short 10-10-10-1.resolvr.io
10.10.10.1
```
