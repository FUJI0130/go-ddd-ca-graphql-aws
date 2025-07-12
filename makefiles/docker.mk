# Docker関連Makefile
# makefiles/docker.mk

#----------------------------------------
# Docker関連コマンド
#----------------------------------------
.PHONY: docker-build docker-run
.PHONY: docker-build-api docker-build-graphql docker-build-grpc 
.PHONY: docker-run-api docker-run-graphql docker-run-grpc
.PHONY: test-docker-api test-docker-graphql test-docker-grpc test-docker-all

# 基本Dockerコマンド
docker-build:
	docker build -t go-ddd-ca .

docker-run:
	docker run -p 8080:8080 go-ddd-ca

# サービス別Dockerビルド
docker-build-api:
	docker build -t test-management-api --build-arg SERVICE_TYPE=api .

docker-build-graphql:
	docker build -t test-management-graphql --build-arg SERVICE_TYPE=graphql .

docker-build-grpc:
	docker build -t test-management-grpc --build-arg SERVICE_TYPE=grpc .

# サービス別Docker実行
docker-run-api:
	docker run -p 8080:8080 --name test-management-api-container test-management-api

docker-run-graphql:
	docker run -p 8080:8080 --name test-management-graphql-container test-management-graphql

docker-run-grpc:
	docker run -p 50051:50051 --name test-management-grpc-container test-management-grpc

# Dockerテスト
test-docker-api:
	@echo "REST APIサービスのテストを実行..."
	@chmod +x scripts/docker/test-api.sh
	@scripts/docker/test-api.sh

test-docker-graphql:
	@echo "GraphQLサービスのテストを実行..."
	@chmod +x scripts/docker/test-graphql.sh
	@scripts/docker/test-graphql.sh

test-docker-grpc:
	@echo "gRPCサービスのテストを実行..."
	@chmod +x scripts/docker/test-grpc.sh
	@scripts/docker/test-grpc.sh

test-docker-all: test-docker-api test-docker-graphql test-docker-grpc
	@echo "すべてのDocker化サービスのテストが完了しました"

# Docker Compose コマンド
docker-compose-up-api:
	SERVICE_TYPE=api PORT=8080 docker-compose -f deployments/docker/docker-compose.dev.yml up -d

docker-compose-up-graphql:
	SERVICE_TYPE=graphql PORT=8081 docker-compose -f deployments/docker/docker-compose.dev.yml up -d

docker-compose-up-grpc:
	SERVICE_TYPE=grpc GRPC_PORT=50051 docker-compose -f deployments/docker/docker-compose.dev.yml up -d

docker-compose-down:
	docker-compose -f deployments/docker/docker-compose.dev.yml down