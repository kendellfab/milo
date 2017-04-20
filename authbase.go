package milo

import (
	"context"
	"net/http"
)

const (
	idKey       = 22
	tokenKey    = 24
	sessAuthKey = "sessauthkey"
	sessID      = "sessid"
	xUserToken  = "X-User-Token"
)

type AuthBase struct {
	*FlashBase
	ac       AuthCheck
	loginURL string
	authKey  string
	xToken   string
}

type AuthCheck interface {
	IsValid(id string) (bool, error)
	IsTokenValid(token string) (bool, error)
}

func NewAuthBase(fb *FlashBase, ac AuthCheck, loginURL string) *AuthBase {
	return &AuthBase{FlashBase: fb, ac: ac, loginURL: loginURL, authKey: sessAuthKey, xToken: xUserToken}
}

func NewAuthBaseCustom(fb *FlashBase, ac AuthCheck, loginURL string, authKey string, xToken string) *AuthBase {
	return &AuthBase{FlashBase: fb, ac: ac, loginURL: loginURL, authKey: authKey, xToken: xToken}
}

func (ab *AuthBase) AuthMiddleware(fn http.HandlerFunc, overrideAuthCheck ...AuthCheck) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		sess, sessErr := ab.store.Get(r, ab.authKey)
		token := r.Header.Get(ab.xToken)
		if token == "" && sessErr != nil {
			ab.SetErrorFlash(w, r, "Error: authorization required")
			ab.Redirect(w, r, ab.loginURL, http.StatusSeeOther)
			return
		}

		id, idOk := sess.Values[sessID]
		if token == "" && !idOk {
			ab.SetErrorFlash(w, r, "Error: authorization required")
			ab.Redirect(w, r, ab.loginURL, http.StatusSeeOther)
			return
		}

		if overrideAuthCheck == nil || len(overrideAuthCheck) == 0 {
			overrideAuthCheck = append(overrideAuthCheck, ab.ac)
		}

		if token != "" {
			for _, oac := range overrideAuthCheck {
				valid, err := oac.IsTokenValid(token)
				if err != nil || !valid {
					w.WriteHeader(http.StatusUnauthorized)
					return
				}
			}

			ctx := r.Context()
			ctx = contextWithToken(ctx, token)
			r = r.WithContext(ctx)
		} else {
			for _, oac := range overrideAuthCheck {
				valid, err := oac.IsValid(id.(string))
				if err != nil || !valid {
					ab.SetErrorFlash(w, r, r.RequestURI+" requires authentication.")
					ab.Redirect(w, r, ab.loginURL, http.StatusSeeOther)
					return
				}
			}

			ctx := r.Context()
			ctx = contextWithId(ctx, id.(string))
			r = r.WithContext(ctx)
		}

		fn(w, r)
	}
}

func (ab *AuthBase) AuthMiddlewareCookie(fn http.HandlerFunc, overrideAuthCheck ...AuthCheck) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		sess, sessErr := ab.store.Get(r, ab.authKey)
		if sessErr != nil {
			ab.SetErrorFlash(w, r, sessErr.Error())
			ab.Redirect(w, r, ab.loginURL, http.StatusSeeOther)
			return
		}

		id, idOk := sess.Values[sessID]
		if !idOk {
			ab.SetErrorFlash(w, r, r.RequestURI+" requires authentication.")
			ab.Redirect(w, r, ab.loginURL, http.StatusSeeOther)
			return
		}

		if overrideAuthCheck == nil || len(overrideAuthCheck) == 0 {
			overrideAuthCheck = append(overrideAuthCheck, ab.ac)
		}

		for _, oac := range overrideAuthCheck {
			valid, err := oac.IsValid(id.(string))
			if err != nil || !valid {
				ab.SetErrorFlash(w, r, r.RequestURI+" requires authentication.")
				ab.Redirect(w, r, ab.loginURL, http.StatusSeeOther)
				return
			}
		}

		ctx := r.Context()
		ctx = contextWithId(ctx, id.(string))
		r = r.WithContext(ctx)

		fn(w, r)
	}
}

func (ab *AuthBase) AuthMiddlewareToken(fn http.HandlerFunc, overrideAuthCheck ...AuthCheck) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		token := r.Header.Get(ab.xToken)
		if token == "" {
			ab.RenderError(w, r, http.StatusForbidden, "Authorization required.")
			return
		}

		if overrideAuthCheck == nil || len(overrideAuthCheck) == 0 {
			overrideAuthCheck = append(overrideAuthCheck, ab.ac)
		}

		for _, oac := range overrideAuthCheck {
			valid, err := oac.IsTokenValid(token)
			if err != nil || !valid {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
		}

		ctx := r.Context()
		ctx = contextWithId(ctx, token)
		r = r.WithContext(ctx)

		fn(w, r)
	}
}

func contextWithToken(ctx context.Context, token string) context.Context {
	return context.WithValue(ctx, tokenKey, token)
}

func TokenFromContext(ctx context.Context) (string, bool) {
	token, ok := ctx.Value(tokenKey).(string)
	return token, ok
}

func contextWithId(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, idKey, id)
}

func IdFromContext(ctx context.Context) (string, bool) {
	id, ok := ctx.Value(idKey).(string)
	return id, ok
}

func (ab *AuthBase) DoLogin(w http.ResponseWriter, r *http.Request, id string) error {
	sess, sessErr := ab.store.Get(r, ab.authKey)
	if sessErr != nil {
		return sessErr
	}
	sess.Values[sessID] = id
	sess.Options.MaxAge = 60 * 60 * 2
	sess.Save(r, w)
	return nil
}

func (ab *AuthBase) DoLogout(w http.ResponseWriter, r *http.Request) error {
	sess, sessErr := ab.store.Get(r, ab.authKey)
	if sessErr != nil {
		return sessErr
	}
	sess.Options.MaxAge = -1
	sess.Save(r, w)
	return nil
}
