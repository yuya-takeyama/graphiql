# GraphiQL Handler fo Go

## Usage

See `./examples/server.go` for more details.

```
func main() {
	http.Handle("/", &graphiql.Handler{
		Endpoint: "/graphql", // Configure GraphQL endpoint
	})

	log.Fatal(http.ListenAndServe(":4000", nil))
}
```

## License

The MIT License
