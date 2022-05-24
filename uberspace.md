# Uberspace setup

## clone the repository and build the binary

``` bash
cd tmp
git clone https://github.com/kernelschmelze/inbox.git
cd inbox
go build
```

## copy the binary and config template

``` bash
cp inbox ~/bin/.
cp inbox.toml.example ~/etc/inbox.toml
```

## set the backend to an application listening on port 25478

See [manual](https://manual.uberspace.de/web-backends/) for more information about backend configuration.

``` bash
uberspace web backend set /inbox --http --port 25478
```

## edit your inbox config file

`~/etc/inbox.toml`

``` toml
listen = ":25478"

[jsonstore]
path = "~/inbox/data"
days = 30

[pushover]
  app = "API Token"
  user = "Your User Key"
```
## try it

``` bash
cd ~/bin
./inbox -f ~/etc/inbox.toml 
```

## setup supervisord

See [manual](https://manual.uberspace.de/daemons-supervisord/) for more information about supervisord configuration.

`~/etc/services.d/inbox.ini`

``` bash
[program:inbox]
command=inbox -f %(ENV_HOME)s/etc/inbox.toml
autostart=yes
autorestart=yes
```

## refresh supervisord configuration

``` bash
supervisorctl reread
supervisorctl update
```
