package respond

import (
	"sync"
	"net/http"
	"log"
	"fmt"
	"crypto/md5"
	"io"
	"bytes"
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

type ReadHasher interface {
	io.Reader
	Hash() []byte
}

type rhImpl struct {
	io.Reader
	hash []byte
}

func (rh *rhImpl) Hash() []byte {
	return rh.hash
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

func with(w http.ResponseWriter, r *http.Request, status int, in io.Reader, headers []string, opts *Options, multiple bool) {
	if opts != nil && multiple && !opts.AllowMultiple {
		panic("respond: multiple responses")
	}

	if hasher, ok := in.(ReadHasher); ok {
		if status >= 200 && status < 300 {
			etag := fmt.Sprintf("\"%s\"", hasher.Hash())

			w.Header().Set("Etag", etag)

			if match := r.Header.Get("if-none-match"); match == etag {
				w.WriteHeader(http.StatusNotModified)
				return
			}
		}
	}

	for i := 0; i < len(headers)/2; i++ {
		w.Header().Set(headers[i*2], headers[i*2 + 1])
	}

	w.WriteHeader(status)
	io.Copy(w, in)
}

func With(w http.ResponseWriter, r *http.Request, status int, data interface{}, headers ...string) {
	opts, multiple := about(r)

	var out []byte
	if bs, ok := data.([]byte); ok {
		out = bs
	} else {
		encoder := JSON
		if opts != nil {
			if opts.Before != nil {
				status, data = opts.Before(w, r, status, data)
			}

			if opts.Encoder != nil {
				encoder = opts.Encoder
			}
		}

		bs, err := encoder.Encode(data)
		if err != nil {
			opts.log(err)
			return
		}

		headers = append(headers, "Content-Type", encoder.ContentType())
		out = bs
	}

	with(w, r, status, &rhImpl{bytes.NewBuffer(out), md5.New().Sum(out)}, headers, opts, multiple)
}

func WithReader(w http.ResponseWriter, r *http.Request, status int, reader io.Reader, headers ...string) {
	opts, multiple := about(r)
	with(w, r, status, reader, headers, opts, multiple)
}

func WithStatus(w http.ResponseWriter, r *http.Request, status int, headers ...string) {
	const (
		fieldStatus = "status"
		fieldCode   = "code"
	)
	With(w, r, status, map[string]interface{}{fieldStatus: http.StatusText(status), fieldCode: status}, headers...)
}
