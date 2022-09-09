package main

import (
	"flag"
	"fmt"
	"github.com/carto-run/fake-jenkins/constants"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/goji/httpauth"
	"github.com/gorilla/mux"
)

var srv *http.Server
var servingScheme string

const (
	jobWithParametersUsingDefaults  = 101
	jobWithParametersWithParameters = 102
	jobWithoutParameters            = 103
)

func main() {
	var certFile string
	var keyFile string
	var port int
	var bindLocalhost bool

	flag.StringVar(&certFile, "cert", "", "Location of a file containing the TLS certificate")
	flag.StringVar(&keyFile, "key", "", "Location of the file containing the key of the TLS certificate")
	flag.IntVar(&port, "port", constants.DefaultPort, "Port to run the server on (8443 will be used if TLS certs are specified)")
	flag.BoolVar(&bindLocalhost, "local", false, "Bind to localhost (127.0.0.1) only")

	flag.Parse()

	r := mux.NewRouter()
	r.HandleFunc("/crumbIssuer/api/json", CrumbHandler)
	r.HandleFunc("/job/{name}/api/json", JobInfoHandler)
	r.HandleFunc("/job/{name}/build", BuildHandler)
	r.HandleFunc("/job/{name}/buildWithParameters", BuildHandlerWithParameters)
	r.HandleFunc("/queue/{id}/api/json", QueueInfoHandler)
	r.HandleFunc("/job/{name}/{id}/api/json", BuildInfoHandler)
	r.HandleFunc("/job/{name}/{id}/logText/progressiveText", BuildLogHandler)

	// log all requests
	r.Use(loggingMiddleware)

	// auth
	r.Use(httpauth.SimpleBasicAuth("jenkins", "token"))

	http.Handle("/", r)

	// use a TLS port if certs are specified
	if certFile != "" && keyFile != "" && port == 8080 {
		port = 8443
	}

	var addr string
	if bindLocalhost {
		addr = fmt.Sprintf("127.0.0.1:%d", port)
	} else {
		addr = fmt.Sprintf(":%d", port)
	}

	log.Printf("Binding to %v.\n", addr)

	srv = &http.Server{
		Handler:           r,
		Addr:              addr,
		WriteTimeout:      15 * time.Second,
		ReadTimeout:       15 * time.Second,
		ReadHeaderTimeout: 15 * time.Second,
	}

	fmt.Printf("Starting on %d...\n", port)

	if certFile != "" && keyFile != "" {
		servingScheme = "https"
		log.Fatal(srv.ListenAndServeTLS(certFile, keyFile))
	} else {
		servingScheme = "http"
		log.Fatal(srv.ListenAndServe())
	}
}

func CrumbHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, `{"crumb":"some_crumb", "crumbRequestField": "some_crumb_request_field"}`)
}

const jobInfoWithoutParameters = `{
	"name": "job-no-parameters",
	"url": "http://%s/job/job-no-parameters/",
	"property": []
}`

const jobInfoWithParameters = `{
	"name": "job-with-parameters",
	"url": "http://%s/job/job-with-parameters/",
	"property": [
		{
			"_class": "hudson.model.ParametersDefinitionProperty",
			"parameterDefinitions": [
				{
					"_class": "hudson.model.StringParameterDefinition",
					"defaultParameterValue": {
						"_class": "hudson.model.StringParameterValue",
						"name": "anything",
						"value": "something"
					},
					"description": null,
					"name": "anything",
					"type": "StringParameterDefinition"
				}
			]
		}
	]
}`

func JobInfoHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	name := vars["name"]

	if name == "job-no-parameters" {
		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		body := fmt.Sprintf(jobInfoWithoutParameters, srv.Addr)
		fmt.Println(body)
		fmt.Fprint(w, body)
		return
	}

	if name == "job-with-parameters" {
		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		body := fmt.Sprintf(jobInfoWithParameters, srv.Addr)
		fmt.Println(body)
		fmt.Fprint(w, body)
		return
	}

	fmt.Println("[WARN] unknown job", name)
	w.WriteHeader(http.StatusNotFound)
}

func BuildHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	name := vars["name"]
	log.Println("scheduling job", name)

	location := url.URL{
		Scheme: servingScheme,
		Host:   r.Host,
		Path:   fmt.Sprintf("queue/%d", jobWithoutParameters),
	}

	log.Println(location.String())
	w.Header().Set("Location", location.String())
	w.WriteHeader(http.StatusOK)
}

func BuildHandlerWithParameters(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	name := vars["name"]
	log.Println("scheduling job with parameters", name)
	var location url.URL

	err := r.ParseForm()
	if err != nil {
		panic(err)
	}
	log.Println("Params", len(r.Form))

	if len(r.Form) == 0 {
		location = url.URL{
			Scheme: servingScheme,
			Host:   r.Host,
			Path:   fmt.Sprintf("queue/%d", jobWithParametersUsingDefaults),
		}
	} else {
		location = url.URL{
			Scheme: servingScheme,
			Host:   r.Host,
			Path:   fmt.Sprintf("queue/%d", jobWithParametersWithParameters),
		}
	}

	log.Println(location.String())
	w.Header().Set("Location", location.String())
	w.WriteHeader(http.StatusOK)
}

func QueueInfoHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	body := fmt.Sprintf(`{"executable":{"number":%s}}`, id)
	fmt.Println(body)
	fmt.Fprint(w, body)
}

func BuildInfoHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	body := `{"building": false, "result": "SUCCESS"}`
	fmt.Println(body)
	fmt.Fprint(w, body)
}

func BuildLogHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, _ := strconv.Atoi(vars["id"])

	w.Header().Add("Content-Type", "text/plain")
	w.Header().Add("X-Text-Size", "20")

	w.WriteHeader(http.StatusOK)

	var body string
	if id == jobWithParametersUsingDefaults {
		body = `Hello, world!`
	} else if id == jobWithParametersWithParameters {
		body = `Hello, buddy!`
	} else {
		body = `Hello, everyone!`
	}

	fmt.Println(body)
	fmt.Fprint(w, body)
}

func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Do stuff here
		log.Println(r.RequestURI)
		// Call the next handler, which can be another middleware in the chain, or the final handler.
		next.ServeHTTP(w, r)
	})
}
