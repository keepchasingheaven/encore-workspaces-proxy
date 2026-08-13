package auth

import (
	"fmt"
	"net/http"
	"strings"
	"time"
)

const (
	SessionCookieName = "encore-workspaces-session"
)

func checkIfOldCookieExists(r *http.Request, config *Config) bool {
	cookie, err := r.Cookie(SessionCookieName)
	if err != nil {
		return false
	}

	if cookie.Value == "" {
		return false
	}

	claim, err := getOldCookieJWT(config.SigningKey, cookie.Value)
	if err != nil {
		// isso pode acontecer apenas se o cookie antigo é
		// por algum motivo mal formado. não se preocupar com isso
		return false
	}

	return !claim.NewFormat
}

func checkIfValidCookieExists(r *http.Request, config *Config, workspaceID string) bool {
	cookie, err := r.Cookie(SessionCookieName)
	if err != nil {
		return false
	}

	if cookie.Value == "" {
		return false
	}

	valid, _ := validateCookieJWT(config.SigningKey, cookie.Value, workspaceID)

	return valid
}

func setCookie(w http.ResponseWriter, value string, expires int, secure bool) {
	cookie := &http.Cookie{
		Path:     "/",
		Name:     SessionCookieName,
		Value:    value,
		Expires:  time.Now().Add(time.Duration(expires) * time.Second),
		Secure:   secure,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode
	}
	
	http.SetCookie(w, cookie)
}

func setOldCookie(w http.ResponseWriter, domain string) {
	domainElements := strings.Split(domain, ":")[0]
	wildcardDomain := strings.Join(strings.Split(domainElements, ".")[1:], ".")
	
	cookie := &http.Cookie{
		Path:    "/",
		Domain:  fmt.Sprintf(".%s", wildcardDomain),
		Name:    SessionCookieName,
		Value:   "",
		Expires: time.Now().Add(-1 * time.Second),
		Secure:  false
	}
	
	http.SetCookie(w, cookie)
}
