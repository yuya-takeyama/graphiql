package main

import (
	"log"
	"net/http"

	"github.com/neelance/graphql-go"
	"github.com/neelance/graphql-go/example/starwars"
	"github.com/neelance/graphql-go/relay"
	"github.com/yuya-takeyama/graphiql"
)

var schema *graphql.Schema

func init() {
	var err error
	schema, err = graphql.ParseSchema(starwars.Schema, &starwars.Resolver{})
	if err != nil {
		panic(err)
	}
}

func main() {
	http.Handle("/graphql", &relay.Handler{Schema: schema})
	http.Handle("/", &graphiql.Handler{
		Endpoint: "/graphql",
	})

	log.Fatal(http.ListenAndServe(":4000", nil))
}
