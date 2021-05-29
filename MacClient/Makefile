build:
	go build -o ledofication main.go
distribution:
	rm -rf ledofication.app
	mkdir -p ledofication.app/Contents/MacOS
	mkdir -p ledofication.app/Contents/Resources
	cp Info.plist ledofication.app/Contents/
	cp ledofication ledofication.app/Contents/MacOS/
	cp led.icns ledofication.app/Contents/Resources/Icon.icns
	cp config.json ledofication.app/Contents/MacOS/
	codesign --force --deep --sign - ledofication.app
