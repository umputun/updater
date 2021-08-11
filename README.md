# updater  [![Build Status](https://github.com/umputun/updater/workflows/build/badge.svg)](https://github.com/umputun/updater/actions)

Simple web-hook based receiver executing things via HTTP requests.

It was created to allow some simple way to invoke remote updates without exposing any sensitive info, like ssh key, passwords and so on. The updater usually invoked from CI/CD system (i.e. github action) and the actual call looks like `curl https://server/update/task/access-key`

List of tasks defined in the configuration file and each task has its own custom section for the command.

Example of `updater.yml`:

```yaml
tasks:

  - name: remark42-site
    command: |
      echo "update remark42-site"
      docker pull ghcr.io/umputun/remark24-site:master
      docker restart remark42-site

  - name: feed-master
    command: |
      echo "update feed-master"
      docker pull umputun/feed-master
      docker restart feed-master
```

By default the update call synchronous and can be switched to non-blocking mode with `async` query parameter, i.e. `curl https://server/update/task/access-key?async=1`

## all parameters

```
  -f, --file=   config file (default: updater.yml) [$CONF]
  -l, --listen= listen on host:port (default: localhost:8080) [$LISTEN]
  -k, --key=    secret key [$KEY]
  -b, --batch   batch mode for multi-line scripts
      --dbg     show debug info

Help Options:
  -h, --help    Show this help message

```