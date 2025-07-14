.PHONY: test test-utils test-service test-handler bench bench-utils bench-service bench-handler cover cover-utils cover-service cover-handler clean

# 运行所有测试
test:
	go test ./...

# 运行 utils 包测试
test-utils:
	go test -v ./internal/utils

# 运行 service 包测试
test-service:
	go test -v ./internal/service

# 运行 handler 包测试
test-handler:
	go test -v ./internal/handler

# 运行所有性能测试
bench:
	go test -bench=. ./internal/utils
	go test -bench=. ./internal/service
	go test -bench=. ./internal/handler

# 运行 utils 包性能测试
bench-utils:
	go test -bench=. ./internal/utils

# 运行 service 包性能测试
bench-service:
	go test -bench=. ./internal/service

# 运行 handler 包性能测试
bench-handler:
	go test -bench=. ./internal/handler

# 运行所有覆盖率测试
cover:
	go test -cover ./...
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

# 运行 utils 包覆盖率测试
cover-utils:
	go test -cover ./internal/utils
	go test -coverprofile=coverage-utils.out ./internal/utils
	go tool cover -html=coverage-utils.out -o coverage-utils.html

# 运行 service 包覆盖率测试
cover-service:
	go test -cover ./internal/service
	go test -coverprofile=coverage-service.out ./internal/service
	go tool cover -html=coverage-service.out -o coverage-service.html

# 运行 handler 包覆盖率测试
cover-handler:
	go test -cover ./internal/handler
	go test -coverprofile=coverage-handler.out ./internal/handler
	go tool cover -html=coverage-handler.out -o coverage-handler.html

# 清理测试文件
clean:
	rm -f coverage*.out coverage*.html

# 运行完整测试套件
test-full: test bench cover
	@echo "所有测试完成"

# 运行单个包的完整测试
test-utils-full: test-utils bench-utils cover-utils
	@echo "utils 包测试完成"

test-service-full: test-service bench-service cover-service
	@echo "service 包测试完成"

test-handler-full: test-handler bench-handler cover-handler
	@echo "handler 包测试完成"