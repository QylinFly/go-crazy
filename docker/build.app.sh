# rm vendor/golang.org/x
# ln -s $PWD/vendor/github.com/golang vendor/golang.org/x

cd ..

cp .env.yml ./docker
CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o ./docker/server.exe .

cd ./docker

git commit -a -m"编译提交"
git push

