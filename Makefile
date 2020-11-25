annie: fmt
	@go build -o annie

fmt: 
	@gofmt -w ./

clean:
	@-rm annie
	@-rm .annie.*

.PHONY: fmt clean
