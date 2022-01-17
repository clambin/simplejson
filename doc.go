/*
Package grafana_json provides a Go implementation for Grafana's SimpleJSON datasource: https://grafana.com/grafana/plugins/grafana-simple-json-datasource

Overview

A grafana_json server is an HTTP server that supports one or more handlers.  Each handler can support multiple targets,
each of which can be supported by a timeseries or table query.  Optionally tags can be used to alter the behaviour of the query
(e.g. filtering what data should be returned).  Finally, a handler can support annotations, i.e. a set of timestamps with associated text.

Server

To create a SimpleJSON server, create a Server and run it:

	s := grafana_json.Server{
		Handlers: []grafana_json.Handler{myHandler},
	}
	err := s.Run(8080)

This starts a server, listening on port 8080, with one handler (myHandler).

Handler

A handler groups a set of targets and supports timeseries queries, table queues and/or annotations, possibly with tags.
The Handler interface includes one function (Endpoints). This function returns the Grafana SimpleJSON endpoints that the handler supports.
Those can be:


	- Search()      implements the /search endpoint: it returns the list of supported targets
	- Query()       implements the /query endpoint for timeseries targets
	- TableQuery()  implements the /query endpoint for table targets
	- Annotations() implements the /annotations endpoint
	- TagKeys()     implements the /tag-keys endpoint
	- TagValues()   implements the /tag-values endpoint

Of these, Search() is mandatory. You will typically want to implement either Query or TableQuery.

Here's an example of a handler that supports timeseries queries:

	type myHandler struct {
	}

	func (handler myHandler) Endpoints() grafana_json.Endpoints {
		return grafana_json.Endpoints{
			Search: handler.Search,
			Query:  handler.Query
		}
	}

	func (handler myHandler) Search() []string {
		return []string{"myTarget"}
	}

	func (handler *myHandler) Query(ctx context.Context, target string, target *grafana_json.TimeSeriesQueryArgs) (response *grafana_json.QueryResponse, err error) {
		// build response
		return
	}

Timeseries Queries

Timeseries queries returns values as a list of timestamp/value tuples. Here's an example of a timeseries query handler:

	func (handler *myHandler) Query(_ context.Context, _ string, _ *grafana_json.TimeSeriesQueryArgs) (response *grafana_json.QueryResponse, err error) {
		response = &grafana_json.QueryResponse{
			Target: "A",
			DataPoints: []grafana_json.QueryResponseDataPoint{
				{Timestamp: time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC), Value: 100},
				{Timestamp: time.Date(2020, 1, 1, 0, 1, 0, 0, time.UTC), Value: 101},
				{Timestamp: time.Date(2020, 1, 1, 0, 2, 0, 0, time.UTC), Value: 103},
			},
		}
		return
	}

Table Queries

Table Queries, on the other hand, return data organized in columns & rows.  Each column needs to have the same number of rows:

	func (handler *myHandler) TableQuery(_ context.Context, _ string, _ *grafana_json.TableQueryArgs) (response *grafana_json.QueryResponse, err error) {
		response = &grafana_json.TableQueryResponse{
			Columns: []grafana_json.TableQueryResponseColumn{
				{ Text: "Time",     Data: grafana_json.TableQueryResponseTimeColumn{time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC), time.Date(2020, 1, 1, 0, 1, 0, 0, time.UTC)} },
				{ Text: "Label",    Data: grafana_json.TableQueryResponseStringColumn{"foo", "bar"}},
				{ Text: "Series A", Data: grafana_json.TableQueryResponseNumberColumn{42, 43}},
				{ Text: "Series B", Data: grafana_json.TableQueryResponseNumberColumn{64.5, 100.0}},
			},
		}
		return
	}

Other topics

For information on query arguments, annotations and tags, refer to the documentation for those data structures.

*/
package grafana_json
