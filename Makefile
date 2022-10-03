build:
	go build -o bin/clex main.go
 
install: build
	sudo mv bin/clex /usr/local/bin
	sudo cp clex.service /etc/systemd/system/
	sudo systemctl daemon-reload

uninstall:
	sudo rm /usr/local/bin/clex
	sudo rm /etc/systemd/system/clex.service
	sudo systemctl daemon-reload

clean:
	go clean
