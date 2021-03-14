package contenttyper

import (
  "fmt"
  "path"
  "strings"

  "github.com/gin-gonic/gin"
  "github.com/spf13/pflag"
  go_service "github.com/tlowerison/go-service"
)

type Middleware struct {
  Mimetypes   map[string]string
  mimetypeArr []string
}

func New() *Middleware {
  return &Middleware{
    Mimetypes:   map[string]string{},
    mimetypeArr: []string{},
  }
}

func (m *Middleware) Register() {
  pflag.StringArrayVar(&m.mimetypeArr, "mime-type", []string{}, "Key/value pairs which map file extensions to mime-types. Ex: --mime-type .yaml=text/yaml")
}

func (m *Middleware) Handler() gin.HandlerFunc {
  for _, mimetype := range m.mimetypeArr {
    components := strings.Split(mimetype, "=")
    if len(components) != 2 {
      panic(fmt.Errorf("Invalid mime-type flag: must use format --mime-type .ext=type"))
    }
    m.AddMimeType(components[0], components[1])
  }
  return func(c *gin.Context) {
    go_service.SetStart(c)
    ext := path.Ext(c.Params.ByName("filepath"))
    if mimetype, ok := m.Mimetypes[ext]; ok {
      c.Writer.Header().Set("Content-Type", mimetype)
    }
		c.Next()
	}
}

func (m *Middleware) AddMimeType(ext string, mimetype string) {
  m.Mimetypes[ext] = mimetype
}

func (m *Middleware) AddMimeTypes(mimetypes map[string]string) {
  for ext, mimetype := range mimetypes {
    m.AddMimeType(ext, mimetype)
  }
}
