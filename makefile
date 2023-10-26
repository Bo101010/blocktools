build:
	rm -rf bin
	mkdir bin
	go build -o bin/task1 task1/main.go
	go build -o bin/task2 task2/main.go
	go build -o bin/task3 task3/main.go

