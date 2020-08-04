# dynflare

[![CI](https://github.com/jpillora/dynflare/workflows/CI/badge.svg)](https://github.com/jpillora/dynflare/actions?workflow=CI)

[Dynamic DNS](https://en.wikipedia.org/wiki/Dynamic_DNS) using Cloudflare. Periodically fetch your public IP address and ensure update the given `domain`'s A record to match it. Useful for keep

### Install

**Binaries**

[![Releases](https://img.shields.io/github/release/jpillora/dynflare.svg)](https://github.com/jpillora/dynflare/releases) [![Releases](https://img.shields.io/github/downloads/jpillora/dynflare/total.svg)](https://github.com/jpillora/dynflare/releases)

See [the latest release](https://github.com/jpillora/dynflare/releases/latest) or download and install it now with `curl https://i.jpillora.com/dynflare! | bash`

**Source**

``` sh
$ go get -v github.com/jpillora/dynflare
```

### Usage

```
$ dynflare --help

  Usage: dynflare [options]

  Options:
  --interval, -i  polling interval (default 5m0s)
  --token, -t     cloudflare token (env TOKEN)
  --domain, -d    domain (env DOMAIN)
  --dry-run       only print dns updates
  --help, -h      display help

```

Systemd

```
$ cat dynflare.service
[Unit]
Description=dynflare

[Service]
Environment=DOMAIN=foobar.jpillora.com
Environment=TOKEN=MySuperSecretToken
ExecStart=/usr/local/bin/dynflare
Restart=always
RestartSec=3

[Install]
WantedBy=default.target
```

### FAQ

* How do I get a token?
    * Visit https://dash.cloudflare.com/profile/api-tokens
    * Create an API Token with minimal scope (**not** an API Key)

#### MIT License

Copyright © 2020 Jaime Pillora &lt;dev@jpillora.com&gt;

Permission is hereby granted, free of charge, to any person obtaining
a copy of this software and associated documentation files (the
'Software'), to deal in the Software without restriction, including
without limitation the rights to use, copy, modify, merge, publish,
distribute, sublicense, and/or sell copies of the Software, and to
permit persons to whom the Software is furnished to do so, subject to
the following conditions:

The above copyright notice and this permission notice shall be
included in all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED 'AS IS', WITHOUT WARRANTY OF ANY KIND,
EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF
MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT.
IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY
CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT,
TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE
SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.
