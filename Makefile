.PHONY: run
run:
	go run ./main.go

.PHONY: dev
dev:
	# https://github.com/watchexec/watchexec
	CONFIG_PATH=dev-config.yaml watchexec -r -e go --wrap-process session -- "go run ./main.go"
