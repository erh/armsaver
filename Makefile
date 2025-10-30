
main: *.go go.*
	go build -o main *.go

demo: main
	./main -host painter-demo01-main.j7u7m261t1.viam.cloud -cmd replay data/image.json data/middle.json grab data/pass1.json open data/middle.json data/image.json
