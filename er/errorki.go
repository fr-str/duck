package er

type Type uint16

// fix this
const (
	InternalServerError Type = 100

	OK        Type = 200
	Error     Type = 300
	NotFound  Type = 400
	Missing   Type = 500
	Timeout   Type = 600
	Forbbiden Type = 700
	Exists    Type = 800
	Invalid   Type = 900
)

const (
	ReqID Type = iota
	Action
	ActionArgs
	ContainerName
	Container
	ContainerIsRunning
	Decode
	Exited
)
