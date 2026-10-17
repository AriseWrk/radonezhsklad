package handler

import (
    "io"
    "os"
    "path/filepath"
    "strconv"
    "strings"

    "github.com/gin-gonic/gin"

    "github.com/radonezhsklad/shared/httpx"
)

type LogsHandler struct {
    dir string
}

func NewLogs(dir string) *LogsHandler {
    return &LogsHandler{dir: dir}
}

// Tail — GET /api/v1/audit/logs?kind=out|err&service=all|<name>&lines=N
func (h *LogsHandler) Tail(c *gin.Context) {
    kind := c.DefaultQuery("kind", "out")
    if kind != "out" && kind != "err" {
        kind = "out"
    }
    suffix := ".log"
    if kind == "err" {
        suffix = ".err.log"
    }
    service := c.DefaultQuery("service", "all")

    linesN := 200
    if s := c.Query("lines"); s != "" {
        if n, err := strconv.Atoi(s); err == nil && n > 0 && n <= 500 {
            linesN = n
        }
    }

    files, _ := filepath.Glob(filepath.Join(h.dir, "*"+suffix))

    perFile := linesN
    if service == "all" {
        perFile = linesN / 6
        if perFile < 20 {
            perFile = 20
        }
    }

    items := []gin.H{}
    for _, f := range files {
        base := filepath.Base(f)
        name := strings.TrimSuffix(base, suffix)
        if name == "" {
            continue
        }
        if service != "all" && name != service {
            continue
        }
        for _, ln := range tailFile(f, perFile) {
            items = append(items, gin.H{"service": name, "line": ln})
        }
    }

    httpx.OK(c, gin.H{"kind": kind, "service": service, "items": items})
}

func tailFile(path string, n int) []string {
    const maxBytes = 2 << 20
    f, err := os.Open(path)
    if err != nil {
        return nil
    }
    defer f.Close()
    if st, err := f.Stat(); err == nil && st.Size() > maxBytes {
        if _, err := f.Seek(-maxBytes, io.SeekEnd); err == nil {
            buf := make([]byte, 1)
            for {
                if _, err := f.Read(buf); err != nil || buf[0] == 0x0A {
                    break
                }
            }
        }
    }
    data, err := io.ReadAll(f)
    if err != nil {
        return nil
    }
    s := strings.TrimRight(string(data), "\r\n")
    if s == "" {
        return []string{}
    }
    parts := strings.Split(s, "\n")
    if len(parts) > n {
        parts = parts[len(parts)-n:]
    }
    for i, ln := range parts {
        parts[i] = strings.TrimRight(ln, "\r")
    }
    return parts
}
