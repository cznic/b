all:
	go fmt
	go test -i
	go test
	go build
	go vet
	go install
	make todo

todo:
	@grep -n ^[[:space:]]*_[[:space:]]*=[[:space:]][[:alpha:]][[:alnum:]]* *.go || true
	@grep -n TODO *.go || true
	@grep -n BUG *.go || true
	@grep -n println *.go || true

clean:
	@go clean
	rm -f *~ cov cov.html

gocov:
	gocov test $(COV) | gocov-html > cov.html

generic:
	@# writes to stdout a version where the type of key is KEY and the type
	@# of value is VALUE.
	@#
	@# Intended use is to replace all textual occurrences of KEY or VALUE in
	@# the output with your desired types.
	@sed -e 's|interface{}[^{]*/\*K\*/|KEY|g' -e 's|interface{}[^{]*/\*V\*/|VALUE|g' btree.go
