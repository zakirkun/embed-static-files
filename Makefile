.PHONY: generate

update-pkg-cache:
	GOPROXY=https://proxy.golang.org GO111MODULE=on \
	go get github.com/$(USER)/$(PACKAGE)@v$(VERSION)

# ⚠️
generate:
	@go generate ./...
	@echo "[OK] Files added to embed box!"

security:
	@gosec ./..
	@echo "[OK] Go security check was completed!"

build: generate 
	@go build -o ./build/server ./cmd/app/*.go
	@echo "[OK] App binary was created!"

run:
	@./build/server -config ./cmd/app/config.yml