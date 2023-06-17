package main

import (
	"io"
	"net"
	"os"
)

type virtualHost interface {
	Name() string
	Handle(c net.Conn) error
}

type A struct {
	name string
}

func NewA(name string) *A {
	return &A{
		name: name,
	}
}

func (a *A) Name() string {
	return a.name
}

func (a *A) Handle(c net.Conn) error {
	defer c.Close()
	_, err := io.Copy(os.Stdout, c)
	return err
}
