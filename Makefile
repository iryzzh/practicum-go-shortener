.PHONY: build
build:
	@cd ./cmd/shortener; go build -o shortener

.PHONY: test
test:
	@go test -v -race -timeout 30s ./...

.PHONY: statictest
statictest:
	@go vet -vettool="$(shell which statictest)" ./...

.PHONY: shortenertest
shortenertest: build increment1 increment2 increment3 increment4 increment5 increment6 increment7 increment8 \
	increment9 increment10 increment11 increment12 increment13 increment14

.PHONY: increment1
increment1:
	@shortenertest -test.v -test.run=^TestIteration1$$ \
                  -binary-path=cmd/shortener/shortener

.PHONY: increment2
increment2:
	@shortenertest -test.v -test.run=^TestIteration2$$ -source-path=.

.PHONY: increment3
increment3:
	@shortenertest -test.v -test.run=^TestIteration3$$ -source-path=.

.PHONY: increment4
increment4:
	@shortenertest -test.v -test.run=^TestIteration4$$ \
                  -source-path=. \
                  -binary-path=cmd/shortener/shortener

.PHONY: increment5
increment5:
	@SERVER_HOST="$(shell random domain)"; SERVER_PORT="$(shell random unused-port)"; shortenertest \
		-test.v -test.run=^TestIteration5$$ \
		-binary-path=cmd/shortener/shortener \
		-server-host=$$SERVER_HOST \
		-server-port=$$SERVER_PORT \
		-server-base-url="http://$$SERVER_HOST:$$SERVER_PORT"

.PHONY: increment6
increment6:
	@SERVER_PORT="$(shell random unused-port)"; TEMP_FILE="$(shell random tempfile)"; shortenertest \
		-test.v -test.run=^TestIteration6$$ \
		-binary-path=cmd/shortener/shortener \
		-server-port=$$SERVER_PORT \
		-file-storage-path=$$TEMP_FILE \
		-source-path=.

.PHONY: increment7
increment7:
	@SERVER_PORT="$(shell random unused-port)"; TEMP_FILE="$(shell random tempfile)"; shortenertest \
		-test.v -test.run=^TestIteration7$$ \
		-binary-path=cmd/shortener/shortener \
		-server-port=$$SERVER_PORT \
		-file-storage-path=$$TEMP_FILE \
		-source-path=.

.PHONY: increment8
increment8:
	@shortenertest -test.v -test.run=^TestIteration8$$ \
		-source-path=. \
		-binary-path=cmd/shortener/shortener

.PHONY: increment9
increment9:
	@shortenertest -test.v -test.run=^TestIteration9$$ \
		-source-path=. \
		-binary-path=cmd/shortener/shortener

.PHONY: increment10
increment10:
	@shortenertest -test.v -test.run=^TestIteration10$$ \
                  -source-path=. \
                  -binary-path=cmd/shortener/shortener \
                  -database-dsn='postgresql://shortener:my_super_password12345@localhost/shortener?sslmode=disable'

.PHONY: increment11
increment11:
	@shortenertest -test.v -test.run=^TestIteration11$$ \
                  -binary-path=cmd/shortener/shortener \
                  -database-dsn='postgresql://shortener:my_super_password12345@localhost/shortener?sslmode=disable'

.PHONY: increment12
increment12:
	@shortenertest -test.v -test.run=^TestIteration12$$ \
                  -binary-path=cmd/shortener/shortener \
                  -database-dsn='postgresql://shortener:my_super_password12345@localhost/shortener?sslmode=disable'

.PHONY: increment13
increment13:
	@shortenertest -test.v -test.run=^TestIteration13$$ \
                  -binary-path=cmd/shortener/shortener \
                  -database-dsn='postgresql://shortener:my_super_password12345@localhost/shortener?sslmode=disable'

.PHONY: increment14
increment14:
	@shortenertest -test.v -test.run=^TestIteration14$$ \
                  -binary-path=cmd/shortener/shortener \
                  -database-dsn='postgresql://shortener:my_super_password12345@localhost/shortener?sslmode=disable'

.DEFAULT_GOAL := build
