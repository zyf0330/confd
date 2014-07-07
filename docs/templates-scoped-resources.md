# Templates - Template Prefix Example

Using confd to manage nginx proxy config of several apps in subdomains

## Add two apps with upstream servers to etcd

myapp
```
curl http://127.0.0.1:4001/v2/keys/myapp/upstream -XPUT -d dir=true
curl http://127.0.0.1:4001/v2/keys/myapp/subdomain -XPUT -d value="myapp"
curl http://127.0.0.1:4001/v2/keys/myapp/upstream/app2 -XPUT -d value="10.0.1.101:80"
curl http://127.0.0.1:4001/v2/keys/myapp/upstream/app1 -XPUT -d value="10.0.1.100:80"
```

yourapp
```
curl http://127.0.0.1:4001/v2/keys/yourapp/upstream -XPUT -d dir=true
curl http://127.0.0.1:4001/v2/keys/yourapp/subdomain -XPUT -d value="yourapp"
curl http://127.0.0.1:4001/v2/keys/yourapp/upstream/app2 -XPUT -d value="10.0.1.103:80"
curl http://127.0.0.1:4001/v2/keys/yourapp/upstream/app1 -XPUT -d value="10.0.1.102:80"
```

## Create template resources

/etc/confd/conf.d/myapp-nginx.toml

```TOML
prefix = "myapp"
src = "nginx.tmpl"
dest = "/tmp/myapp.conf"
owner = "nginx"
mode = "0644"
keys = [
  "subdomain",
  "upstream",
]
check_cmd = "/usr/sbin/nginx -t -c {{.src}}"
reload_cmd = "/usr/sbin/service nginx reload"
```

/etc/confd/conf.d/yourapp-nginx.toml

```TOML
prefix = "yourapp"
src = "nginx.tmpl"
dest = "/tmp/yourapp.conf"
owner = "nginx"
mode = "0644"
keys = [
  "subdomain",
  "upstream",
]
check_cmd = "/usr/sbin/nginx -t -c {{.src}}"
reload_cmd = "/usr/sbin/service nginx reload"
```

## Create a source template

/etc/confd/templates/nginx.tmpl

```
upstream {{getv "subdomain"}} {
{{range getvs "upstream"}}
    server {{.}};
{{end}}
}

server {
    server_name  {{getv "subdomain"}}.example.com;
    location / {
        proxy_pass        http://{{getv "subdomain"}};
        proxy_redirect    off;
        proxy_set_header  Host             $host;
        proxy_set_header  X-Real-IP        $remote_addr;
        proxy_set_header  X-Forwarded-For  $proxy_add_x_forwarded_for;
   }
}
```