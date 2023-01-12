package hub

import (
	"io"
	"net/http"
	"strings"

	"golang.org/x/net/webdav"
)

var _ http.Handler = (*Hub)(nil)

func (s *Hub) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var host Host = getHostFromReqHost(r.Host, s.LocalDomain)

	client, err := s.AllowDir(host, r.URL.Path)
	if err != nil {
		http.Error(w, "get dir allow access failed", http.StatusInternalServerError)
		return
	}
	if client == nil {
		if r.URL.Path == "/version" {
			io.WriteString(w, "lcode-hub\n")
			io.WriteString(w, s.bash.VERSION+"\n")
			return
		}
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}

	handler := s.GetWebdavHandler(client)
	if handler == nil {
		http.Error(w, "client not ready", http.StatusNotFound)
		return
	}

	handler.ServeHTTP(w, r)
}

func getHostFromReqHost(h string, localdomain string) (host Host) {
	h = strings.Split(h, ":")[0]
	h = strings.TrimSuffix(h, localdomain)
	return parseIDRaw(h)
}

func (s *Hub) GetWebdavHandler(id *Client) *webdav.Handler {
	if id == nil {
		return nil
	}
	fs := s.bash.Get(id.String())
	if fs == nil {
		return nil
	}
	item := s.lockers.Get(id.ToHost().String())
	if item == nil {
		return nil
	}
	return &webdav.Handler{
		FileSystem: fs,
		LockSystem: item.Value(),
	}
}
