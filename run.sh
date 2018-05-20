# rm vendor/golang.org/x
# ln -s $PWD/vendor/github.com/golang vendor/golang.org/x

# CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o server.exe .

clear
go build server.go   &&  ./server -Detcd.url=10.99.2.116:2379 -Dserver.port=8080
