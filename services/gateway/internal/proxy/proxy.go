package proxy

import (
"log/slog"
"net/http"
"net/http/httputil"
"net/url"

"github.com/gin-gonic/gin"

mw "github.com/radonezhsklad/shared/middleware"
)

func New(targetURL string) gin.HandlerFunc {
target, err := url.Parse(targetURL)
if err != nil {
panic("proxy: bad target URL " + targetURL + ": " + err.Error())
}

rp := httputil.NewSingleHostReverseProxy(target)
originalDirector := rp.Director
rp.Director = func(req *http.Request) {
originalDirector(req)
req.Host = target.Host

req.Header.Del("X-Forwarded-For")
req.Header.Del("X-Forwarded-Host")
req.Header.Del("X-Forwarded-Proto")
}

rp.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
slog.Error("gateway upstream error",
"target", targetURL,
"path", r.URL.Path,
"method", r.Method,
"error", err,
)
w.Header().Set("Content-Type", "application/json")
w.WriteHeader(http.StatusBadGateway)
_, _ = w.Write([]byte(`{"error":{"code":"bad_gateway","message":"upstream unavailable"}}`))
}

rp.ModifyResponse = func(resp *http.Response) error {
resp.Header.Del("Server")
return nil
}

return func(c *gin.Context) {
if rid := c.GetString(mw.CtxRequestID); rid != "" {
c.Request.Header.Set(mw.HeaderRequestID, rid)
}
c.Request.Header.Del("Connection")
c.Request.Header.Del("Proxy-Connection")
c.Request.Header.Del("Keep-Alive")
c.Request.Header.Del("Transfer-Encoding")
c.Request.Header.Del("Upgrade")

rp.ServeHTTP(c.Writer, c.Request)
}
}