package grafana_json

import (
	"fmt"
	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	// log "github.com/sirupsen/logrus"
	"net/http"
)

// APIServer implements a generic frameworks for the Grafana simpleJson API datasource
type APIServer struct {
	apiHandler APIHandler
	port       int
}

// APIHandler implements the business logic of the Grafana API datasource so that
// APIServer can be limited to providing the generic search/query framework
type APIHandler interface {
	Search() []string
	Query(target string, request *QueryRequest) (*QueryResponse, error)
	QueryTable(target string, request *QueryRequest) (*QueryTableResponse, error)
}

// Create creates a APIServer object
func Create(apiHandler APIHandler, port int) *APIServer {
	return &APIServer{apiHandler: apiHandler, port: port}
}

// Run the API Server
func (apiServer *APIServer) Run() error {
	r := mux.NewRouter()
	r.Use(prometheusMiddleware)
	r.Path("/metrics").Handler(promhttp.Handler())
	r.HandleFunc("/", apiServer.hello)
	r.HandleFunc("/search", apiServer.search).Methods("POST")
	r.HandleFunc("/query", apiServer.query).Methods("POST")

	return http.ListenAndServe(fmt.Sprintf(":%d", apiServer.port), r)
}

// Prometheus metrics
var (
	httpDuration = promauto.NewSummaryVec(prometheus.SummaryOpts{
		Name: "grafana_api_duration_seconds",
		Help: "Grafana API duration of HTTP requests.",
	}, []string{"path"})
)

func prometheusMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		route := mux.CurrentRoute(r)
		path, _ := route.GetPathTemplate()
		timer := prometheus.NewTimer(httpDuration.WithLabelValues(path))
		next.ServeHTTP(w, r)
		timer.ObserveDuration()
	})
}
