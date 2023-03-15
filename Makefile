build:
	go build -o bin/clex main.go
 
install: build
	mv bin/clex $HOME/.local/bin
	cp clex.service $HOME/.config/systemd/user/ 
	chmod +x $HOME/.local/bin/clex 
	systemctl --user daemon-reload
	systemctl --user enable clex.service
	systemctl --user start clex.service 

uninstall:
	systemctl --user stop clex.service 
	systemctl --user disable clex.service 
	rm $HOME/.local/bin/clex
	rm $HOME/.config/systemd/user/clex.service
	systemctl --user daemon-reload

clean:
	go clean
