package respond

import (
	"sync"
	"net/http"
	"log"
	"fmt"
	"crypto/md5"
)

var (
	mutex     sync.RWMutex
	options   map[*http.Request]*Options
	responded map[*http.Request]bool
)

func init() {
	options   = make(map[*http.Request]*Options)
	responded = make(map[*http.Request]bool)
}

type Options struct {
	AllowMultiple bool

	OnError func(error)

	Encoder Encoder

	Before func(w http.ResponseWriter, r *http.Request, status int, data interface{}) (int, interface{})

	StatusFunc func(w http.ResponseWriter, r *http.Request, status int) interface{}
}

// Handler wraps an HTTP handler becoming the source of options for all
// containing With calls.
func (opts *Options) Handler(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mutex.Lock()
		options[r] = opts
		mutex.Unlock()

		defer func() {
			mutex.Lock()
			delete(options, r)
			delete(responded, r)
			mutex.Unlock()
		}()

		handler.ServeHTTP(w, r)
	})
}

func (opts *Options) log(err error) {
	if opts.OnError == nil {
		log.Println("respond: " + err.Error())
	}

	opts.OnError(err)
}

func about(r *http.Request) (opts *Options, multiple bool) {
	mutex.RLock()
	opts = options[r]
	multiple = responded[r]
	mutex.RUnlock()

	return opts, multiple
}

func with(w http.ResponseWriter, r *http.Request, status int, data interface{}, opts *Options, multiple bool) {
	if opts != nil && multiple && !opts.AllowMultiple {
		panic("respond: multiple responses")
	}

	encoder := JSON
	if opts != nil {
		if opts.Before != nil {
			status, data = opts.Before(w, r, status, data)
		}

		if opts.Encoder != nil {
			encoder = opts.Encoder
		}
	}

	out, err := encoder.Encode(data)
	if err != nil {
		opts.log(err)
		return
	}

	if status >= 200 && status < 300 {
		etag := fmt.Sprintf("\"%x\"", md5.Sum(out))

		w.Header().Set("Etag", etag)

		if match := r.Header.Get("if-none-match"); match == etag {
			w.WriteHeader(http.StatusNotModified)
			return
		}
	}

	w.Header().Set("Content-Type", encoder.ContentType())
	w.WriteHeader(status)
	w.Write(out)
}

func With(w http.ResponseWriter, r *http.Request, status int, data interface{}) {
	opts, multiple := about(r)

	with(w, r, status, data, opts, multiple)
}

func WithStatus(w http.ResponseWriter, r *http.Request, status int) {
	opts, multiple := about(r)

	var data interface{}
	if opts != nil && opts.StatusFunc != nil {
		data = opts.StatusFunc(w, r, status)
	} else {
		const (
			fieldStatus = "status"
			fieldCode   = "code"
		)
		data = map[string]interface{}{fieldStatus: http.StatusText(status), fieldCode: status}
	}

	with(w, r, status, data, opts, multiple)
}
