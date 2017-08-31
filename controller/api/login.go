package api

import (
	"encoding/json"
	log "github.com/Sirupsen/logrus"
	"github.com/nicle-lin/lillian/controller/manager"
	"github.com/nicle-lin/lillian/helper/auth"
	"github.com/nicle-lin/lillian/helper/auth/ldap"
	"net/http"
)

func (a *Api) register(w http.ResponseWriter, r *http.Request) {

}

func (a *Api) login(w http.ResponseWriter, r *http.Request) {
	var creds *Credentials
	if err := json.NewDecoder(r.Body).Decode(&creds); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	loginSuccessful, err := a.manager.Authenticate(creds.Username, creds.Password)
	if err != nil {
		log.Errorf("登陆出错 %s from %s: %s", creds.Username, r.RemoteAddr, err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if !loginSuccessful {
		log.Warnf("无效的登陆r %s from %s", creds.Username, r.RemoteAddr)
		http.Error(w, "无效的用户名和密码", http.StatusForbidden)
		//http.Error(w, "invalid username/password", http.StatusForbidden)
		return
	}

	// check for ldap and autocreate for users
	if a.manager.GetAuthenticator().Name() == "ldap" {
		if a.manager.GetAuthenticator().(*ldap.LdapAuthenticator).AutocreateUsers {
			defaultAccessLevel := a.manager.GetAuthenticator().(*ldap.LdapAuthenticator).DefaultAccessLevel
			log.Debug("ldap: checking for existing user account and creating if necessary")
			// give default users readonly access to containers
			acct := &auth.Account{
				Username: creds.Username,
				Roles:    []string{defaultAccessLevel},
			}

			// check for existing account
			if _, err := a.manager.Account(creds.Username); err != nil {
				if err == manager.ErrAccountDoesNotExist {
					log.Debugf("autocreating user for ldap: username=%s access=%s", creds.Username, defaultAccessLevel)
					if err := a.manager.SaveAccount(acct); err != nil {
						log.Errorf("error autocreating ldap user %s: %s", creds.Username, err)
						http.Error(w, err.Error(), http.StatusInternalServerError)
						return
					}
				} else {
					log.Errorf("error checking user for autocreate: %s", err)
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}
			}

		}
	}

	// return token
	token, err := a.manager.NewAuthToken(creds.Username, r.UserAgent())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if err := json.NewEncoder(w).Encode(token); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (a *Api) changePassword(w http.ResponseWriter, r *http.Request) {
	var creds *Credentials
	if err := json.NewDecoder(r.Body).Decode(&creds); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	username := a.manager.Store(w,r).Get("username").(string)
	if username == "" {
		http.Error(w, "没有认证", http.StatusInternalServerError)
		//http.Error(w, "unauthorized", http.StatusInternalServerError)
		return
	}
	if err := a.manager.ChangePassword(username, creds.Password); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
