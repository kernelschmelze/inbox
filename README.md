# inbox

Inbox is a simple upload server with pushover notification.

``` bash
echo "Hallo Freund" | \
  curl -F file=@- -F "from=Kay" -F "subject=re: inbox" http://localhost:25478/inbox
```

Persist to disk.  

`./data/f6ab2023-c582-4939-ba77-d758e3b85aee`

``` json
{
  "time":"2022-05-21T18:33:51.562755+02:00",
  "id":"f6ab2023-c582-4939-ba77-d758e3b85aee",
  "from":"Kay",
  "subject":"re: inbox",
  "payload":"SGFsbG8gRnJldW5kCg=="
}

```

With optional pushover notification.  

![](screenshot/screenshot1.jpeg)



## Usage

``` bash
./inbox -f inbox.toml
```

## Config

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

