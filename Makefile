
# make run mem=2000 maxmem=10000 path=server
run:
	go run main.go -mem=$(mem) -maxmem=$(maxmem) -path=$(path)