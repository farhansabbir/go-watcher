# export Version=`git rev-list -1 HEAD` && go build -o bin/fsswatcher -ldflags "-X main.Version=$Version" main.go
make all
