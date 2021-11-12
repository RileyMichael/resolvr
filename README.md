# resolvr

simple wildcard dns server that will extract ip addresses from the hostname and resolve to them -- no more editing /etc/hosts!

![terminal example](https://www.resolvr.io/terminal-dark.svg)

## config

[vrischmann/envconfig](https://github.com/vrischmann/envconfig) is used to load config from env variables. 
defaults are configured for local dev, however you can override them at runtime like so:

```bash
export RESOLVR_HOSTNAME=your.hostname.
export RESOLVR_BIND_ADDRESS=0.0.0.0:53
export RESOLVR_METRICS_ADDRESS=0.0.0.0:9091
export RESOLVR_ENV=prod
export RESOLVR_STATIC_TYPE_A_RECORDS={your.hostname.,10.10.10.2},{prod.your.hostname.,10.10.10.3}
export RESOLVR_STATIC_TYPE_AAAA_RECORDS={your.hostname.,::1}
export RESOLVR_STATIC_TYPE_CNAME_RECORDS={www.your.hostname.,your.hostname.}
export RESOLVR_NAMESERVERS={ns1.your.hostname.,10.10.10.2},{ns2.your.hostname.,10.10.10.3}
```
