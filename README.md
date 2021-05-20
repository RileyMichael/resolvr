# resolvr

simple wildcard dns server that will extract ip addresses from the hostname and resolve to them

similar to xip.io which appears to have gone offline with the recent basecamp exodus

```bash
# say you have an internal ip 10.10.10.1
> dig +short 10-10-10-1.resolvr.io
10.10.10.1
```

## config

[vrischmann/envconfig](https://github.com/vrischmann/envconfig) is used to load config from env variables. 
defaults are configured for local dev, however you can override them at runtime like so:

```bash
export RESOLVR_HOSTNAME=your.hostname.
export RESOLVR_ADDRESS=10.10.10.1
export RESOLVR_BIND_ADDRESS=0.0.0.0:53
export RESOLVR_ENV=prod
export RESOLVR_NAMESERVER={ns1.your.hostname.,10.10.10.2},{ns2.your.hostname.,10.10.10.3}
```