# proplo
ProxyProtocol Logger daemon


## Usage

```console
$ proplo
proplo [local_host:port] [upstream_host:port]
```

## Example

Run `proplo` on 127.0.0.1:9876 which proxy to 10.8.0.1:80.

```console
$ proplo 127.0.0.1:9876 10.8.0.1:80
2020/02/25 00:52:07 [info] Upstream 10.8.0.1:80
2020/02/25 00:52:07 [info] Listening 127.0.0.1:9876
```

Try connect to `proplo` with Proxy Protocol and send some payloads.

```
$ telnet 127.0.0.1 9876
Trying 127.0.0.1...
Connected to localhost.
Escape character is '^]'.
PROXY TCP4 192.168.1.1 172.16.0.1 9999 8888
GET / HTTP/1.0

```

`proplo` proxies the payloads to the upstream and outputs logs of the connection.

```json
{"id":"e4640056-4811-4fce-9948-d00231cf5454","type":"connect","time":"2020-02-25T00:52:36.68708+09:00","client_addr":"192.168.1.1:9999","proxy_addr":"10.8.0.6:61424","upstream_addr":"10.8.0.1:80","status":"connected","client_at":"2020-02-25T00:52:17.000776+09:00","upstream_at":"2020-02-25T00:52:36.687079+09:00"}
{"id":"e4640056-4811-4fce-9948-d00231cf5454","type":"transfer","time":"2020-02-25T00:52:44.835855+09:00","src_addr":"10.8.0.1:80","proxy_addr":"10.8.0.6:61424","dest_addr":"192.168.1.1:9999","bytes":337,"duration":27.834918903,"error":null}
{"id":"e4640056-4811-4fce-9948-d00231cf5454","type":"transfer","time":"2020-02-25T00:52:44.836469+09:00","src_addr":"192.168.1.1:9999","proxy_addr":"10.8.0.6:61424","dest_addr":"10.8.0.1:80","bytes":18,"duration":27.835533649,"error":null}
```

### Ignoring health check access

When you ignore health check accesses from specified CIDR, set `-ignore-cidr` flag.

`proplo -ignore-cidr 192.168.0.0/24` ignores logging and proxying to upstream connect from `192.168.0.0/24`.

## LICENSE

MIT
