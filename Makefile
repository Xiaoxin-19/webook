.PHONY: docker clean build

# 检测操作系统类型
UNAME_S := $(shell uname -s)

# 定义适配命令
ifeq ($(OS), Windows_NT)
    RM = del /F /Q
else
    RM = rm -f
endif
# 编译目标
TARGET := webook


ifeq ($(OS), Windows_NT)
    BUILD_CMD = powershell -Command "& { \
         $$env:GOOS='linux'; \
         $$env:GOARCH='arm'; \
        go build -tags=k8s -o $(TARGET) .; \
    }"
else
    BUILD_CMD = GOOS=linux GOARCH=arm go build -tags=k8s -o $(TARGET) .
endif

# Docker 镜像信息
DOCKER_IMAGE := xiaoxina/webook:v0.0.1



docker: clean build
	@echo "Building Docker image: $(DOCKER_IMAGE)"
	@docker rmi -f $(DOCKER_IMAGE)
	@docker build -t $(DOCKER_IMAGE) .

clean:
	@echo "Cleaning up..."
	@$(RM) $(TARGET) || true

build:
	@echo "Tidying up Go modules..."
	@go mod tidy
	@echo "Building Go binary..."
	@$(BUILD_CMD)

mock:
	@go generate ./...

.PHONY: mock



