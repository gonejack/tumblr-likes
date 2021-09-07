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
1. get credentials from
```
https://api.tumblr.com/console/calls/user/info 
```
2. build config.json
```shell
> tumblr-likes -t > config.json
```
3. fetch likes
```shell
> tumblr-likes -v
```

```
Usage:
  tumblr-likes [flags]

Flags:
  -c, --config string   config file (default "config.json")
  -o, --outdir string   output directory (default "likes")
  -t, --template        print config template
  -v, --verbose         verbose
  -h, --help            help for tumblr-likes
```
