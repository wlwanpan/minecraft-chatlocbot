
# make run mem=2048 maxmem=2048 path=server.jar
run:
	go run main.go -mem=$(mem) -maxmem=$(maxmem) -path=$(path)