.PHONY: clean
clean:
	rm -rf build/*

.PHONY: build
build: clean
	CGO_ENABLED=1 GOOS=darwin GOARCH=amd64 go build -o build/sapphire-amd64
	CGO_ENABLED=1 GOOS=darwin GOARCH=arm64 go build -o build/sapphire-arm64
	lipo -create -output build/sapphire build/sapphire-amd64 build/sapphire-arm64
	mkdir -p Sapphire.app/Contents/MacOS/
	cp build/sapphire Sapphire.app/Contents/MacOS/sapphire

.PHONY: run
run: build
	./Sapphire.app/Contents/MacOS/sapphire

.PHONY: package
package: build
	zip -r build/sapphire.zip Sapphire.app
	hdiutil create -volname Sapphire -srcfolder Sapphire.app -ov -format UDZO build/sapphire.dmg

.PHONY: test
test:
	go test ./...

.PHONY: lint
lint:
	golangci-lint run
