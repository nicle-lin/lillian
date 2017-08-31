package api

import (
	log "github.com/Sirupsen/logrus"
	"github.com/gorilla/context"
	"github.com/gorilla/mux"
	"github.com/nicle-lin/lillian/controller/manager"
	"github.com/nicle-lin/lillian/controller/middleware/access"
	"github.com/nicle-lin/lillian/controller/middleware/audit"
	"github.com/nicle-lin/lillian/controller/middleware/auth"
	"github.com/nicle-lin/lillian/helper/tlsutils"
	"github.com/urfave/negroni"
	"io/ioutil"
	"net/http"
)

type Api struct {
	listenAddr         string
	manager            manager.Manager
	authWhitelistCIDRS []string
	tlsCACertPath      string
	tlsCertPath        string
	tlsKeyPath         string
	allowInsecure      bool
}

type ApiConfig struct {
	ListenAddr         string
	Manager            manager.Manager
	AuthWhitelistCIDRS []string
	TLSCACertPath      string
	TLSCertPath        string
	TLSKeyPath         string
	AllowInsecure      bool
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
		tlsCACertPath:      config.TLSCACertPath,
		tlsCertPath:        config.TLSCertPath,
		tlsKeyPath:         config.TLSKeyPath,
		allowInsecure:      config.AllowInsecure,
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

	//if user config tls
	if a.tlsCertPath != "" && a.tlsKeyPath != "" {
		log.Infof("using TLS for communication:cert=%s key=%s", a.tlsCertPath, a.tlsKeyPath)

		var caCert []byte
		if a.tlsCACertPath != "" {
			ca, err := ioutil.ReadFile(a.tlsCACertPath)
			if err != nil {
				return err
			}
			caCert = ca
		}

		serverCert, err := ioutil.ReadFile(a.tlsCertPath)
		if err != nil {
			return err
		}

		serverKey, err := ioutil.ReadFile(a.tlsKeyPath)
		if err != nil {
			return err
		}
		tlsConfig, err := tlsutils.GetServerTLSConfig(caCert, serverCert, serverKey, a.allowInsecure)
		if err != nil {
			return err
		}
		s.TLSConfig = tlsConfig
		return s.ListenAndServeTLS(a.tlsCertPath, a.tlsKeyPath)
	}
	return s.ListenAndServe()
}
