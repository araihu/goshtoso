package server

import "net/http"

const (
	storagePreferenceCookie = "gt_storage"
	storagePreferenceDenied = "denied"
)

func storageAllowed(r *http.Request) bool {
	c, err := r.Cookie(storagePreferenceCookie)
	return err != nil || c.Value != storagePreferenceDenied
}
