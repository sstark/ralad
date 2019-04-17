package main

type RaladError struct {
	s   string
	Err error
}

func NewRaladError(s string, e error) error {
	return &RaladError{s, e}
}

func (e *RaladError) Error() string {
	return e.s
}

func (e *RaladError) Wrapped() error {
	return e.Err
}

var ErrMaxRedirects = NewRaladError("maximum number of redirects reached", nil)
