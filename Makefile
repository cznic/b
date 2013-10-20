.PHONY: all todo clean cover generic

all: editor
	go build
	go vet
	go install
	make todo

editor:
	go fmt
	go test -i
	go test

todo:
	@grep -n ^[[:space:]]*_[[:space:]]*=[[:space:]][[:alpha:]][[:alnum:]]* *.go || true
	@grep -n TODO *.go || true
	@grep -n BUG *.go || true
	@grep -n println *.go || true

clean:
	@go clean
	rm -f *~

cover:
	t=$(shell tempfile) ; go test -coverprofile $$t && go tool cover -html $$t && unlink $$t

generic:
	@# writes to stdout a version where the type of key is KEY and the type
	@# of value is VALUE.
	@#
	@# Intended use is to replace all textual occurrences of KEY or VALUE in
	@# the output with your desired types.
	@sed -e 's|interface{}[^{]*/\*K\*/|KEY|g' -e 's|interface{}[^{]*/\*V\*/|VALUE|g' btree.go
