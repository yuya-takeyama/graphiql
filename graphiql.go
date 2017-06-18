package graphiql

import (
	"bytes"
	"encoding/json"
	"html/template"
	"net/http"
	"strings"
)

var graphiqlVersion = "0.10.2"
var fetchVersion = "2.0.1"
var reactVersion = "15.5.4"
var reactDomVersion = "15.5.4"

var compiledTmpl *template.Template

type Handler struct {
	Endpoint string
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	page = new(bytes.Buffer)
	compiledTmpl.Execute(page, &graphiqlData{
		Endpoint:        h.Endpoint,
		GraphiqlVersion: graphiqlVersion,
		FetchVersion:    fetchVersion,
		ReactVersion:    reactVersion,
		ReactDomVersion: reactDomVersion,
		request:         r,
	})

	w.Header().Set("Content-Type", "text/html")
	w.Write(page.Bytes())
}

var tmpl = `<!DOCTYPE html>
<html>
<head>
  <meta charset="utf-8" />
  <title>GraphiQL</title>
  <meta name="robots" content="noindex" />
  <style>
    html, body {
      height: 100%;
      margin: 0;
      overflow: hidden;
      width: 100%;
    }
  </style>
  <link href="//cdn.jsdelivr.net/graphiql/{{ .GraphiqlVersion }}/graphiql.css" rel="stylesheet" />
  <script src="//cdn.jsdelivr.net/fetch/{{ .FetchVersion }}/fetch.min.js"></script>
  <script src="//cdn.jsdelivr.net/react/{{ .ReactVersion }}/react.min.js"></script>
  <script src="//cdn.jsdelivr.net/react/{{ .ReactDomVersion }}/react-dom.min.js"></script>
  <script src="//cdn.jsdelivr.net/graphiql/{{ .GraphiqlVersion }}/graphiql.min.js"></script>
</head>
<body>
  <script>
    // Collect the URL parameters
    var parameters = {};
    window.location.search.substr(1).split('&').forEach(function (entry) {
      var eq = entry.indexOf('=');
      if (eq >= 0) {
        parameters[decodeURIComponent(entry.slice(0, eq))] =
          decodeURIComponent(entry.slice(eq + 1));
      }
    });

    function locationURL(params) {
      return '{{ .Endpoint }}' + locationQuery(params);
    }

    // Produce a Location query string from a parameter object.
    function locationQuery(params) {
      return '?' + Object.keys(params).map(function (key) {
        return encodeURIComponent(key) + '=' +
          encodeURIComponent(params[key]);
      }).join('&');
    }

    // Derive a fetch URL from the current URL, sans the GraphQL parameters.
    var graphqlParamNames = {
      query: true,
      variables: true,
      operationName: true
    };

    var otherParams = {};
    for (var k in parameters) {
      if (parameters.hasOwnProperty(k) && graphqlParamNames[k] !== true) {
        otherParams[k] = parameters[k];
      }
    }
    var fetchURL = locationURL(otherParams);

    // Defines a GraphQL fetcher using the fetch API.
    function graphQLFetcher(graphQLParams) {
      return fetch(fetchURL, {
        method: 'post',
        headers: {
          'Accept': 'application/json',
          'Content-Type': 'application/json'
        },
        body: JSON.stringify(graphQLParams),
        credentials: 'include',
      }).then(function (response) {
        return response.text();
      }).then(function (responseBody) {
        try {
          return JSON.parse(responseBody);
        } catch (error) {
          return responseBody;
        }
      });
    }

    // When the query and variables string is edited, update the URL bar so
    // that it can be easily shared.
    function onEditQuery(newQuery) {
      parameters.query = newQuery;
      updateURL();
    }

    function onEditVariables(newVariables) {
      parameters.variables = newVariables;
      updateURL();
    }

    function onEditOperationName(newOperationName) {
      parameters.operationName = newOperationName || '';
      updateURL();
    }

    function updateURL() {
      history.replaceState(null, null, locationQuery(parameters));
    }

    // Render <GraphiQL /> into the body.
    ReactDOM.render(
      React.createElement(GraphiQL, {
        fetcher: graphQLFetcher,
        onEditQuery: onEditQuery,
        onEditVariables: onEditVariables,
        onEditOperationName: onEditOperationName,
        query: '{{ .Query }}' || undefined,
        variables: '{{ .Variables | toSafeJson }}',
        operationName: '{{ .OperationName }}' || undefined,
      }),
      document.body
    );
  </script>
</body>
</html>
`

type graphiqlData struct {
	Endpoint        string
	GraphiqlVersion string
	FetchVersion    string
	ReactVersion    string
	ReactDomVersion string
	request         *http.Request
}

func (d *graphiqlData) Query() string {
	return d.request.URL.Query().Get("query")
}

func (d *graphiqlData) OperationName() string {
	return d.request.URL.Query().Get("operationName")
}

func (d *graphiqlData) Variables() map[string]interface{} {
	var result map[string]interface{}
	_ = json.Unmarshal([]byte(d.request.URL.Query().Get("variables")), &result)
	return result
}

var page *bytes.Buffer

func init() {
	var err error
	funcMap := template.FuncMap{
		"toSafeJson": toSafeJson,
	}
	compiledTmpl, err = template.New("graphiql").Funcs(funcMap).Parse(tmpl)
	if err != nil {
		panic(err)
	}
}

func toJson(d interface{}) string {
	j, err := json.Marshal(d)
	if err != nil {
		return ""
	}

	if string(j) == "null" {
		return ""
	}

	var indented bytes.Buffer
	indentErr := json.Indent(&indented, j, "", "  ")
	if indentErr != nil {
		return ""
	}

	return indented.String()
}

func toSafeJson(d interface{}) string {
	j := toJson(d)
	return strings.Replace(j, "/", "\\/", -1)
}
