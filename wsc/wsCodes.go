package wsc

import (
	"reflect"
)

type Major uint16
type Minor uint16

//go:generate stringer -type=Minor
//go:generate stringer -type=Major
const (
	InternalServerError Major = 100
	OK                  Major = 200
	Error               Major = 300
	NotFound            Major = 400
	Missing             Major = 500
	Timeout             Major = 600
	Forbbiden           Major = 700
	Exists              Major = 800
	Invalid             Major = 900
)

const (
	ReqID Minor = iota
	Action
	ActionArgs
	Container
	ContainerIsRunning
	Decode
	Exited
	Inspect
	ID
	Name
	Image
	Delete
)

type Code interface {
	String() string
	Get() uint16
}

type Codes []Code

func (m Major) Get() uint16 {
	return uint16(m)
}
func (m Minor) Get() uint16 {
	return uint16(m)
}

func (c Codes) Sum() uint16 {
	var sum uint16
	for _, code := range c {
		sum = sum + code.Get()
	}
	return sum
}

func (codes Codes) ToString() (errString string) {
	for _, c := range codes {
		if reflect.TypeOf(c) == reflect.TypeOf(Major(0)) {
			errString = c.String() + errString
			continue
		}
		errString += c.String()
	}
	return
}

func Wrap(codes ...Code) Codes {
	return codes
}
