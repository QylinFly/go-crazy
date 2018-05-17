# go-crazy

A Golang Framework For Web Artisans

#install
1.Create your project folder and cd inside

    $ mkdir -p $GOPATH/src/github.com/myusername/project && cd "$_"

2.Download src:

    git clone https://github.com/QylinFly/go-crazy.git ./

3.Vendor init your project

    govendor init

    Package Management for Golang
    https://github.com/Masterminds/glide

    install
    curl https://glide.sh/get | sh

    glide get github.com/Masterminds/cookoo

    glide install

    $ glide create                            # Start a new workspace
    $ open glide.yaml                         # and edit away!
    $ glide get github.com/Masterminds/cookoo # Get a package and add to glide.yaml
    $ glide install                           # Install packages and dependencies
    # work, work, work
    $ go build                                # Go tools work normally
    $ glide up                                # Update to newest versions of the package

    ### glide
    $ glide mirror set https://golang.org/x/mobile https://github.com/golang/mobile --vcs git
    $ glide mirror set https://golang.org/x/crypto https://github.com/golang/crypto --vcs git
    $ glide mirror set https://golang.org/x/net https://github.com/golang/net --vcs git
    $ glide mirror set https://golang.org/x/tools https://github.com/golang/tools --vcs git
    $ glide mirror set https://golang.org/x/text https://github.com/golang/text --vcs git
    $ glide mirror set https://golang.org/x/image https://github.com/golang/image --vcs git
    $ glide mirror set https://golang.org/x/sys https://github.com/golang/sys --vcs git
    $ glide mirror set https://google.golang.org/grpc https://github.com/grpc/grpc-go --vcs git
    $ glide mirror set https://google.golang.org/genproto https://github.com/google/go-genproto --vcs git




    google.golang.org/grpc

4.Run your project

    $ go run main.go

### vendor

##### config file

     https://github.com/jinzhu/configor

##### ORM:gorm

    https://github.com/jinzhu/gorm

##### log

    https://github.com/uber-go/zap

### manage

#### graceful-shutdown

    http://localhost:8080/down

### elastic 6.1.1

#### github.com/olivere/elastic

    https://github.com/olivere/elastic

### shell

    govendor fetch github.com/gorilla/websocket@v1.2.0
    go build index.go  &&  ./index

    go get -u golang.org/x/net

### 中国安装方法 go get -u golang.org/x

    https://www.golangtc.com/download/package

### Cache Store

    https://github.com/kataras/iris/tree/master/sessions/sessiondb

    /*
    |--------------------------------------------------------------------------
    | Default Cache Store
    |--------------------------------------------------------------------------
    |
    | This option controls the default cache connection that gets used while
    | using this caching library. This connection is used when another is
    | not explicitly specified when executing a given caching function.
    |
    | Supported: "apc", "array", "database", "file", "memcached", "redis"
    |
    */

### 类似框架

    https://github.com/kataras/iris/tree/master/sessions/sessiondb
    https://github.com/laravel/framework/tree/5.5/src/Illuminate
    glide

### 命令行

    "flag"
    http://www.01happy.com/golang-command-line-arguments/
