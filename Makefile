APP     := goscope
CMD     := ./cmd/goscope
OUT_DIR := dist

.PHONY: all linux windows darwin clean

all: linux windows darwin

linux:
	mkdir -p $(OUT_DIR)
	GOOS=linux GOARCH=amd64 go build -o $(OUT_DIR)/$(APP)-linux-amd64 $(CMD)

windows:
	mkdir -p $(OUT_DIR)
	GOOS=windows GOARCH=amd64 go build -o $(OUT_DIR)/$(APP)-windows-amd64.exe $(CMD)

darwin:
	mkdir -p $(OUT_DIR)
	GOOS=darwin GOARCH=amd64 go build -o $(OUT_DIR)/$(APP)-darwin-amd64 $(CMD)

clean:
	rm -rf $(OUT_DIR)
