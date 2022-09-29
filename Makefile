build:
	go build -o clex main.go
 
install: build
	sudo mv ./clex /usr/local/bin
	sudo cp clex.service /etc/systemd/system/
	sudo systemctl daemon-reload

uninstall:
	sudo rm /usr/local/bin/clex
	sudo rm /etc/systemd/system/clex.service
	sudo systemctl daemon-reload

clean:
	go clean
