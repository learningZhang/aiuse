BINARY := mylog
CONF   := ./conf/config/config.yaml

.PHONY: run build tidy fmt vet es-init es-health clean

run:
	go run main.go

build:
	CGO_ENABLED=0 go build -o $(BINARY) main.go

tidy:
	go mod tidy

fmt:
	go fmt ./...

vet:
	go vet ./...

# 创建 diary 索引（已存在会返回 400，属正常）
es-init:
	curl -X PUT "http://127.0.0.1:9200/diary" -H 'Content-Type: application/json' -d @scripts/mapping.json

es-health:
	curl "http://127.0.0.1:9200/_cat/indices/diary?v"

clean:
	rm -f $(BINARY)
