package api

import (
	log "github.com/Sirupsen/logrus"
	"github.com/gorilla/context"
	"github.com/gorilla/mux"
	"github.com/nicle-lin/lillian/controller/manager"
	"github.com/nicle-lin/lillian/controller/middleware/access"
	"github.com/nicle-lin/lillian/controller/middleware/audit"
	"github.com/nicle-lin/lillian/controller/middleware/auth"
	"github.com/urfave/negroni"
	"net/http"
)

type Api struct {
	listenAddr         string
	manager            manager.Manager
	authWhitelistCIDRS []string
}

type ApiConfig struct {
	ListenAddr         string
	Manager            manager.Manager
	AuthWhitelistCIDRS []string
}

type Credentials struct {
	Username string `json:"username,omitempty"`
	Password string `json:"password,omitempty"`
}

func writeCorsHeaders(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Access-Control-Allow-Origin", "*")
	w.Header().Add("Access-Control-Allow-Headers", "Origin, X-Requested-With, Content-Type, Accept")
	w.Header().Add("Access-Control-Allow-Methods", "GET, POST, DELETE, PUT, OPTIONS")
}

func NewApi(config ApiConfig) *Api {
	return &Api{
		listenAddr:         config.ListenAddr,
		manager:            config.Manager,
		authWhitelistCIDRS: config.AuthWhitelistCIDRS,
	}
}

func (a *Api) Run() error {
	globalMux := http.NewServeMux()
	controllerManager := a.manager

	apiRouter := mux.NewRouter()
	apiRouter.HandleFunc("/api/accounts", a.accounts).Methods("GET")
	apiRouter.HandleFunc("/api/accounts", a.saveAccount).Methods("POST")
	apiRouter.HandleFunc("/api/accounts/{username}", a.account).Methods("GET")
	apiRouter.HandleFunc("/api/accounts/{username}", a.deleteAccount).Methods("DELETE")

	auditExcludes := []string{
		"/networks",
		"/images/json",
	}

	apiAuthRouter := negroni.New()
	apiAuthRequired := auth.NewAuthRequired(controllerManager, a.authWhitelistCIDRS)
	apiAccessRequired := access.NewAccessRequired(controllerManager)
	apiAuditor := audit.NewAuditor(controllerManager, auditExcludes)

	apiAuthRouter.Use(negroni.HandlerFunc(apiAuthRequired.HandlerFuncWithNext))
	apiAuthRouter.Use(negroni.HandlerFunc(apiAccessRequired.HandlerFuncWithNext))
	apiAuthRouter.Use(negroni.HandlerFunc(apiAuthRequired.HandlerFuncWithNext))
	apiAuthRouter.Use(negroni.HandlerFunc(apiAuditor.HandlerFuncWithNext))

	apiAuthRouter.UseHandler(apiRouter)
	globalMux.Handle("/api/", apiAuthRouter)

	s := &http.Server{
		Addr:    a.listenAddr,
		Handler: context.ClearHandler(globalMux),
	}

	log.Printf("listening on %s\n", a.listenAddr)
	return s.ListenAndServe()
}
