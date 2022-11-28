/*
Package simplejson provides a Go implementation for Grafana's SimpleJSON datasource: https://grafana.com/grafana/plugins/grafana-simple-json-datasource

# Overview

A simplejson server is an HTTP server that supports one or more handlers.  Each handler can support multiple targets,
each of which can be supported by a timeseries or table query.  Optionally tag can be used to alter the behaviour of the query
(e.g. filtering what data should be returned).  Finally, a handler can support annotation, i.e. a set of timestamps with associated text.

# Server

To create a SimpleJSON server, create a Server and run it:

	s := simplejson.Server{
		Handlers: map[string]simplejson.Handler{
			"my-target": myHandler,
		},
	}
	err := s.Run(8080)

This starts a server, listening on port 8080, with one target "my-target", served by myHandler.

# Handler

A handler serves incoming requests from Grafana, e.g. queries, requests for annotations or tag.
The Handler interface contains all functions a handler needs to implement. It contains only one function (Endpoints).
This function returns the Grafana SimpleJSON endpoints that the handler supports. Those can be:

  - Query()       implements the /query endpoint. handles both timeserie & table responses
  - Annotations() implements the /annotation endpoint
  - TagKeys()     implements the /tag-keys endpoint
  - TagValues()   implements the /tag-values endpoint

Here's an example of a handler that supports timeseries queries:

	type myHandler struct {
	}

	func (handler myHandler) Endpoints() simplejson.Endpoints {
		return simplejson.Endpoints{
			Query:  handler.Query
		}
	}

	func (handler *myHandler) Query(ctx context.Context, target string, target *simplejson.QueryArgs) (response *simplejson.QueryResponse, err error) {
		// build response
		return
	}

# Queries

SimpleJSON supports two types of query responses: timeseries responses and table responses.

Timeseries queries return values as a list of timestamp/value tuples. Here's an example of a timeseries query handler:

	func (handler *myHandler) Query(_ context.Context, _ string, _ query.QueryArgs) (response *simplejson.TimeSeriesResponse, err error) {
		response = &query.TimeSeriesResponse{
			Name: "A",
			DataPoints: []simplejson.DataPoint{
				{Timestamp: time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC), Value: 100},
				{Timestamp: time.Date(2020, 1, 1, 0, 1, 0, 0, time.UTC), Value: 101},
				{Timestamp: time.Date(2020, 1, 1, 0, 2, 0, 0, time.UTC), Value: 103},
			},
		}
		return
	}

Table Queries, on the other hand, return data organized in columns and rows.  Each column needs to have the same number of rows:

	func (handler *myHandler) TableQuery(_ context.Context, _ string, _ query.QueryArgs) (response *simplejson.TableResponse, err error) {
		response = &simplejson.TableResponse{
			Columns: []simplejson.Column{
				{ Text: "Time",     Data: query.TimeColumn{time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC), time.Date(2020, 1, 1, 0, 1, 0, 0, time.UTC)} },
				{ Text: "Label",    Data: query.StringColumn{"foo", "bar"}},
				{ Text: "Series A", Data: query.NumberColumn{42, 43}},
				{ Text: "Series B", Data: query.NumberColumn{64.5, 100.0}},
			},
		}
		return
	}

# Annotations

The /annotations endpoint returns Annotations:

	func (h *handler) Annotations(_ simplejson.QueryRequest) (annotations []simplejson.Annotation, err error) {
		annotations = []simplejson.Annotation{
			{
				Time:  time.Now().Add(-5 * time.Minute),
				Title: "foo",
				Text:  "bar",
				Tags:  []string{"A", "B"},
			},
		}
		return
	}

NOTE: this is only called when using the SimpleJSON datasource. simPod / GrafanaJsonDatasource does not use the /annotations endpoint.
Instead, it will call a regular /query and allows to configure its response as annotations instead.

# Tags

The /tag-keys and /tag-values endpoints return supported keys and key values respectively for your data source.
A Grafana dashboard can then be confirmed to show those keys and its possible values as a filter.

The following sets up a key & key value handler:

	func (h *handler) TagKeys(_ context.Context) (keys []string) {
		return []string{"some-key"}
	}

	func (h *handler) TagValues(_ context.Context, key string) (values []string, err error) {
		if key != "some-key" {
			return nil, fmt.Errorf("invalid key: %s", key)
		}
		return []string{"A", "B", "C"}, nil
	}

When the dashboard performs a query with a tag selected, that tag & value will be added in the request's AdHocFilters.

# Metrics

simplejson exports two Prometheus metrics for performance analytics:

	simplejson_query_duration_seconds: duration of query requests by target, in seconds
	simplejson_query_failed_count:     number of failed query requests

# Other topics

For information on query arguments and tags, refer to the documentation for those data structures.
*/
package simplejson
