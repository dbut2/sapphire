.PHONY: clean
clean:
	rm -rf build/*

.PHONY: generate
generate:
	openssl enc -d -aes-256-cbc -pbkdf2 -in enc/bios.gba.enc -out gba/bios.gba -pass pass:$$SAPPHIRE_KEY
	openssl enc -d -aes-256-cbc -pbkdf2 -in enc/sapphire.gba.enc -out cmd/web/gamepak.gba -pass pass:$$SAPPHIRE_KEY

.PHONY: build
build: clean generate
	CGO_ENABLED=1 GOOS=darwin GOARCH=amd64 go build -tags embed_bios -o build/sapphire-amd64 .
	CGO_ENABLED=1 GOOS=darwin GOARCH=arm64 go build -tags embed_bios -o build/sapphire-arm64 .
	lipo -create -output build/sapphire build/sapphire-amd64 build/sapphire-arm64
	mkdir -p Sapphire.app/Contents/MacOS/
	cp build/sapphire Sapphire.app/Contents/MacOS/sapphire

.PHONY: wasm
wasm: generate
	GOOS=js GOARCH=wasm go build -tags "embed_bios embed_gamepak" -o build/sapphire.wasm ./cmd/web

.PHONY: run
run: build
	./Sapphire.app/Contents/MacOS/sapphire

.PHONY: package
package: build wasm
	zip -r build/sapphire.zip Sapphire.app
	hdiutil create -volname Sapphire -srcfolder Sapphire.app -ov -format UDZO build/sapphire.dmg

.PHONY: test
test:
	go test ./...

.PHONY: lint
lint:
	golangci-lint run
