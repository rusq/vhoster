package vhoster

import "net/url"

// URI is a wrapper around [pkg/net/url.URL] that implements custom
// json.Marshaler and json.Unmarshaler.
type URI url.URL

// convenience functions
func (u *URI) URL() *url.URL { return (*url.URL)(u) }
func ToURI(u *url.URL) *URI  { return (*URI)(u) }

func (u *URI) UnmarshalJSON(b []byte) error {
	if len(b) < 2 {
		return nil
	}
	b = b[1 : len(b)-1]
	return (*url.URL)(u).UnmarshalBinary(b)
}

func (u *URI) MarshalJSON() ([]byte, error) {
	b, err := (*url.URL)(u).MarshalBinary()
	if err != nil {
		return nil, err
	}
	return append([]byte{'"'}, append(b, '"')...), nil
}

func (u *URI) String() string {
	return (*url.URL)(u).String()
}

// Must panics if err is not nil.
func Must(u *URI, err error) *URI {
	if err != nil {
		panic(err)
	}
	return u
}

// Parse is a wrapper around [pkg/url.Parse].
func Parse(s string) (*URI, error) {
	u, err := url.Parse(s)
	if err != nil {
		return nil, err
	}
	return ToURI(u), nil
}
