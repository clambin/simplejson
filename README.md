# simplejson
![GitHub tag (latest by date)](https://img.shields.io/github/v/release/clambin/simplejson?color=green&label=Release&style=plastic)
![Codecov](https://img.shields.io/codecov/c/gh/clambin/simplejson?style=plastic)
![Test](https://github.com/clambin/simplejson/workflows/Test/badge.svg)
![Go Report Card](https://goreportcard.com/badge/github.com/clambin/simplejson)
![GitHub](https://img.shields.io/github/license/clambin/simplejson?style=plastic)
[![GoDoc](https://pkg.go.dev/badge/github.com/clambin/simplejson?utm_source=godoc)](http://pkg.go.dev/github.com/clambin/simplejson/v3)

Basic Go implementation of a Grafana SimpleJSON server.

Works with:

* [Grafana Simple JSON Datasource](https://grafana.com/grafana/plugins/grafana-simple-json-datasource)
* [JSON API Grafana Datasource](https://grafana.com/grafana/plugins/simpod-json-datasource)

Note: the latter does not call the /annotations endpoint. Return annotations through the /query endpoint instead. 
