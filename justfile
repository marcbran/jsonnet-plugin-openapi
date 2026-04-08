test:
    @go test ./...
    @jpoet test .

build: test
    @jpoet pkg build

push: build
    @jpoet pkg push

test-go-e2e:
    go test -v -tags e2e ./tests/...

test-e2e:
    just test-go-e2e
