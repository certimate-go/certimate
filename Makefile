# 定义变量
BINARY_NAME=certimate
BUILD_DIR=dist

# 支持的操作系统和架构列表
OS_ARCH=\
    linux/amd64 \
    linux/arm64 \
    darwin/amd64 \
    darwin/arm64 \
    windows/amd64 \
    windows/arm64

# 默认目标
all: build

# 构建所有平台的二进制文件
build: $(OS_ARCH)
$(OS_ARCH):
	@mkdir -p $(BUILD_DIR)
	GOOS=$(word 1,$(subst /, ,$@)) \
	GOARCH=$(word 2,$(subst /, ,$@)) \
	CGO_ENABLED=0 \
	go build -trimpath -ldflags="-s -w" -o $(BUILD_DIR)/$(BINARY_NAME)_$(word 1,$(subst /, ,$@))_$(word 2,$(subst /, ,$@)) .

# 清理构建文件
clean:
	rm -rf $(BUILD_DIR)

# 帮助信息
help:
	@echo "Usage:"
	@echo "  make        - 编译所有平台的二进制文件"
	@echo "  make clean  - 清理构建文件"
	@echo "  make help   - 显示此帮助信息"

.PHONY: all build clean help

local.run:
	go mod vendor&& npm --prefix=./ui install && npm --prefix=./ui run build && go run main.go serve --http 127.0.0.1:8090
