	
GOOS=windows GOARCH=amd64 go build -o ./bin/windows/amd64/animeman.exe ./cmd/service/main.go
GOOS=linux GOARCH=amd64 go build -o ./bin/linux/amd64/animeman ./cmd/service/main.go
GOOS=linux GOARCH=arm64 go build -o ./bin/linux/arm64/animeman ./cmd/service/main.go

mkdir releases

zip releases/animeman_Windows_amd64.zip bin/windows/amd64/animeman.exe
zip releases/animeman_Linux_amd64.zip bin/linux/amd64/animeman
zip releases/animeman_Linux_arm64.zip bin/linux/arm64/animeman