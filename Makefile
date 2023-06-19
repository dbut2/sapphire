.PHONY: clean
clean:
	rm -rf build/*

.PHONY: build
build: clean
	mkdir build/Sapphire.app
	./icon.sh

	go build -o build/Sapphire.app/Contents/MacOS/sapphire

.PHONY: package
package: build
	hdiutil create -volname Sapphire -srcfolder Sapphire.app -ov -format UDZO Sapphire.dmg