package milo

import (
	"net/http"

	"github.com/gorilla/sessions"
)

const (
	SessFlash    = "session_flash"
	FlashError   = "flasherror"
	FlashSuccess = "flashsuccess"
)

type FlashBase struct {
	*Renderer
	store sessions.Store
}

func NewFlashBase(r *Renderer, s sessions.Store) *FlashBase {
	return &FlashBase{Renderer: r, store: s}
}

func (fb *FlashBase) SetErrorFlash(w http.ResponseWriter, r *http.Request, message string) {
	fb.setFlashMessage(w, r, FlashError, message)
}

func (fb *FlashBase) SetSuccessFlash(w http.ResponseWriter, r *http.Request, message string) {
	fb.setFlashMessage(w, r, FlashSuccess, message)
}

func (fb *FlashBase) GetFlashes(w http.ResponseWriter, r *http.Request) ([]interface{}, []interface{}) {
	var errFlash []interface{}
	var successFlash []interface{}
	if sess, sessErr := fb.store.Get(r, SessFlash); sessErr == nil {
		errFlash = sess.Flashes(FlashError)
		successFlash = sess.Flashes(FlashSuccess)
		sess.Options.MaxAge = -1
		sess.Save(r, w)
	}
	return errFlash, successFlash
}

func (fb *FlashBase) setFlashMessage(w http.ResponseWriter, r *http.Request, key, message string) {
	if sess, sessErr := fb.store.Get(r, SessFlash); sessErr == nil {
		sess.AddFlash(message, key)
		sess.Save(r, w)
	}
}
