# install gvm

```sh
# https://go.dev/dl/

# ubuntu
apt install -y golang
apt install -y bison

# centos
yum install -y epel-release
yum install -y golang
yum install -y bison

# https://github.com/moovweb/gvm
wget -c https://raw.githubusercontent.com/moovweb/gvm/master/binscripts/gvm-installer
bash gvm-installer
```

# switch environment

```sh
gvm version
gvm listall
gvm install go1.18
gvm use go1.18
go version
```

# config env

```sh
go env -w GO111MODULE=on
go env -w GOPROXY=https://goproxy.cn,direct
```

# run & build

```sh
go run hello.go
go build hello.go
```

# go module

```sh
go mod init test
cat go.mod
```

# gore

```sh
go install github.com/x-motemen/gore/cmd/gore@latest
gore
```

# delve

```sh
go install github.com/go-delve/delve/cmd/dlv@latest
xcode-select --install

dlv debug main.go -- --arg1 value1 --arg2 value2
```

# go project layout

```
https://github.com/golang-standards/project-layout
https://makeoptim.com/golang/standards/project-layout
```

# go-swagger

```sh
brew tap go-swagger/go-swagger
brew install go-swagger

swagger serve ./api.json
```
