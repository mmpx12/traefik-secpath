# Traefik Sec Path

Traefik plugin to white list or black list path based on ip.


## config

config with dynamic config:

```yml
  middlewares:
    secpath:
      plugin:
        secpath:
          rules:
            redir:
              pathRule: ^/re/|/redir/
              ipRule: 172.17.0.1/10,192.168.75.207
              typeRule: redirection
              newPath: /redirect/
            admin:
              pathRule: ^/admin/|^/me/
              ipRule: 172.17.0.1/10,192.168.75.207
              typeRule: fake
              newPath: /fake/
            superadmin:
              pathRule: /super-admin/
              ipRule: 172.20.0.1/24
              typeRule: allow
            megaadmin:
              pathRule: /mega-admin/
              ipRule: 172.20.0.1-172.20.0.255
              typeRule: block
```

config with docker labels:

```yml
  nginx:
    labels:
      - traefik.enable=true
      - traefik.http.routers.nginx.rule=Host(`nginx.local.me`)
      - traefik.http.routers.nginx.entrypoints=web
      - traefik.http.services.nginx.loadbalancer.server.port=80
      # redirect
      - traefik.http.middlewares.secpath-plugin.plugin.secpath.rules.redir.pathRule=^/re/|/redir/
      - traefik.http.middlewares.secpath-plugin.plugin.secpath.rules.redir.ipRule=172.17.0.1/10,192.168.75.207
      - traefik.http.middlewares.secpath-plugin.plugin.secpath.rules.redir.typeRule=redirection
      - traefik.http.middlewares.secpath-plugin.plugin.secpath.rules.redir.newPath=/redirect/
      # allow
      - traefik.http.middlewares.secpath-plugin.plugin.secpath.rules.superadmin.pathRule=/super-admin/
      - traefik.http.middlewares.secpath-plugin.plugin.secpath.rules.superadmin.ipRule=172.20.0.1/24
      - traefik.http.middlewares.secpath-plugin.plugin.secpath.rules.superadmin.typeRule=allow
      # block
      - traefik.http.middlewares.secpath-plugin.plugin.secpath.rules.megaadmin.pathRule=/mega-admin/
      - traefik.http.middlewares.secpath-plugin.plugin.secpath.rules.megaadmin.ipRule=172.20.0.1-172.20.0.255
      - traefik.http.middlewares.secpath-plugin.plugin.secpath.rules.megaadmin.typeRule=block
      # fake
      - traefik.http.middlewares.secpath-plugin.plugin.secpath.rules.admin.pathRule=^/admin/|^/me/
      - traefik.http.middlewares.secpath-plugin.plugin.secpath.rules.admin.ipRule=172.17.0.1/10,192.168.75.207
      - traefik.http.middlewares.secpath-plugin.plugin.secpath.rules.admin.typeRule=fake
      - traefik.http.middlewares.secpath-plugin.plugin.secpath.rules.admin.newPath=/fake/
```

Path can use regex and should be speparayed by comma (without space)

Ips can be specify by:

  - single ip  172.17.0.1
  - cidr       172.17.0.1/27
  - nmap style 172.17.0.1-172.17.0.255

mulitple ips/range can be specify and need to be comma separated (without space)


4 rules are available:

- allow
- block
- redirect
- fake

Allow will only allow specify ips to acces path (return unauthorized if not), block will block specify ips to acces path (unauthorized), redirect will redirect specify ip to newPath (newPath should be set).

Fake will fake path ex: user request /admin/ but server while serve /fake-admin/ (url in browser while still show /admin/).


for local testing with docker-compose.yml you should clone this repo to: ./plugins-local/src/github.com/mmpx12/traefik-secpath: 


```
.
├── README.md
├── demo
│   └── html
│       ├── 404.html
│       ├── 50x.html
│       ├── admin
│       │   └── index.html
│       ├── fake
│       │   └── index.html
│       ├── index.html
│       └── redirect
│           └── index.html
├── docker-compose.yml
├── go.mod
├── plugins-local
│   └── src
│       └── github.com
│           └── mmpx12
│               └── traefik-secpath
│                   ├── go.mod
│                   └── secpath.go
└── secpath.go
```
