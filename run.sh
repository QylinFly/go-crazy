# rm vendor/golang.org/x
# ln -s $PWD/vendor/github.com/golang vendor/golang.org/x
go build server.go  &&  ./server