package render

import (
	"context"
	"fmt"
	"mime"
	"net/http"
	"strings"
)

var (
	ContentTypeCtxKey = &contextKey{"ContentType"}
)

// ContentTypeSet is a ordered set of content types
type ContentTypeSet struct {
	set []ContentType
	pos int
}

func (set *ContentTypeSet) String() string {
	if set == nil || len(set.set) == 0 {
		return ""
	}
	strs := make([]string, len(set.set))
	for i := range set.set {
		strs[i] = string(set.set[i])
	}
	return strings.Join(strs, ",")
}

// Next returns if there is another content type waiting, and if there is
// advance to it
func (set *ContentTypeSet) Next() bool {
	if set == nil {
		return false
	}
	set.pos++
	return set.pos < len(set.set)
}

// Reset to the start of the content types
func (set *ContentTypeSet) Reset() {
	if set != nil {
		set.pos = -1
	}
}

// Type returns the current ContentType of the set
func (set *ContentTypeSet) Type() ContentType {
	if set == nil {
		return ""
	}
	p := set.pos
	if p >= len(set.set) {
		p = len(set.set) - 1
	} else if p <= 0 {
		p = 0
	}
	return set.set[p]
}

// Types returns a copy of the content types in order specified
func (set *ContentTypeSet) Types() (types []ContentType) {
	if set == nil || len(set.set) == 0 {
		return []ContentType{}
	}
	return append(make([]ContentType, 0, len(set.set)), set.set...)
}

// Has checks to see if the set contains the content type
func (set *ContentTypeSet) Has(contentType ContentType) bool {
	if set == nil {
		return false
	}
	for _, c := range set.set {
		if c == contentType {
			return true
		}
	}
	return false
}

// NewContentTypeSet returns a new set of ContentTypes based on the set of strings passed in. mime.ParseMediaType is
// used to prase each string. Empty strings and strings that do not parse are ignored.
func NewContentTypeSet(types ...string) *ContentTypeSet {
	if len(types) == 0 {
		return nil
	}
	set := &ContentTypeSet{
		set: make([]ContentType, 0, len(types)),
		pos: -1,
	}
allTypes:
	for _, t := range types {
		mediaType, _, err := mime.ParseMediaType(t)
		if err != nil {
			// skip types that can not be parsed
			continue
		}
		// Let's make sure we have not seen this type before.
		for _, tt := range set.set {
			if tt == ContentType(mediaType) {
				// Don't add it to the set, already exists
				continue allTypes
			}
		}
		set.set = append(set.set, ContentType(mediaType))
	}
	if len(set.set) == 0 {
		return nil
	}
	return set
}

// ContentType is an enumeration of common HTTP content types.
type ContentType string

func (contentType ContentType) String() string { return string(contentType) }

// GetContentType returns the base mimetype from the string. This uses mime.ParseMediaType to
// actually parse the string.
func GetContentType(str string) (ContentType, error) {
	mediaType, _, err := mime.ParseMediaType(str)
	return ContentType(mediaType), err
}

// ContentTypes that are commonly used
const (
	ContentTypeNone        = ContentType("")
	ContentTypeJSON        = ContentType("application/json")
	ContentTypeData        = ContentType("application/octet-stream")
	ContentTypeForm        = ContentType("multipart/form-data")
	ContentTypeEventStream = ContentType("text/event-stream")
	ContentTypeHTML        = ContentType("text/html")
	ContentTypePlainText   = ContentType("text/plain")
	ContentTypeXML         = ContentType("text/xml")
)

// SetContentType is a middleware that forces response Content-Type.
func SetContentType(contentType ContentType) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			r = r.WithContext(context.WithValue(r.Context(), ContentTypeCtxKey, contentType))
			next.ServeHTTP(w, r)
		}
		return http.HandlerFunc(fn)
	}
}

func AllowedContentTypes(contentTypes ContentTypeSet) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ct, _ := GetContentType(r.Header.Get("Content-Type"))
			if !contentTypes.Has(ct) {
				http.Error(w,
					fmt.Sprintf("invalid content type: accepted types are:%v", contentTypes),
					http.StatusNotAcceptable,
				)
			}
		})
	}
}

// GetRequestContentType is a helper function that returns ContentType based on
// context or "content-Type" request header.
func GetRequestContentType(r *http.Request, dflt ContentType) ContentType {
	if contentType, ok := r.Context().Value(ContentTypeCtxKey).(ContentType); ok {
		return contentType
	}
	ct, err := GetContentType(r.Header.Get("Content-Type"))
	if err != nil {
		return dflt
	}
	return ct
}

// GetAcceptedContentType is a helper function that returns a set of ContentTypes based
// on context or "Accept" request header.
func GetAcceptedContentType(r *http.Request) *ContentTypeSet {
	if contentType, ok := r.Context().Value(ContentTypeCtxKey).(ContentType); ok {
		return NewContentTypeSet(string(contentType))
	}

	// Parse request Accept header.
	fields := strings.Split(r.Header.Get("Accept"), ",")
	return NewContentTypeSet(fields...)
}
