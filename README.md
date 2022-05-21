### inbox

Inbox is a small upload server with pushover notification.

`inbox.toml`  

``` toml

# 25478
# :25478
# http://127.0.0.1:25478
# https://127.0.0.1:25478

listen = ":25478"

# tls crt and key file

crt = "srv.crt"
key = "srv.key"

# pushover config

[pushover]
  app = "API Token"
  user = "Your User Key"
  
```

