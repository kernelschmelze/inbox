### inbox

Inbox is a simple upload server with pushover notification.

``` bash
echo "Hallo Freund" | curl -F file=@- -F "from=Kay" -F "subject=re: inbox" http://localhost:25478/inbox

```

![](screenshot/screenshot1.jpeg)

``` bash
./inbox -f inbox.toml
```

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

