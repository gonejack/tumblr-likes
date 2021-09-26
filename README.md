# tumblr-likes
Command line tool for download tumblr like images.

![GitHub go.mod Go version](https://img.shields.io/github/go-mod/go-version/gonejack/tumblr-likes)
![Build](https://github.com/gonejack/tumblr-likes/actions/workflows/go.yml/badge.svg)
[![GitHub license](https://img.shields.io/github/license/gonejack/tumblr-likes.svg?color=blue)](LICENSE)

### Install
```shell
> go get github.com/gonejack/tumblr-likes
```

### Usage
1. Get credentials
```
https://api.tumblr.com/console/calls/user/info 
```
2. Edit config.json
```shell
> tumblr-likes -t > config.json && open config.json
```
3. Get likes
```shell
> tumblr-likes -v
```

### Flags
```
  -h, --help                    Show context-sensitive help.
  -c, --config="config.json"    Config file.
  -o, --output="likes"          Output directory.
  -m, --max-fetch=200           How many likes to be fetched.
  -t, --template                Print config template.
  -v, --verbose                 Verbose printing.
      --about                   Show about.
```
