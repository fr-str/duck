package er

type Type uint16

// ErrMissReqID = errors.New("invalid request - missing `RequestID`")
// ErrMissAction = errors.New("invalid request - missing `Action`")

const (
	InternalServerError Type = 100

	OK        Type = 200
	Error     Type = 300
	Forbbiden Type = 400
	Missing   Type = 500
	Timeout   Type = 600
	NotFound  Type = 700
)

const (
	ReqID Type = iota
	Action
	ContainerName
	Container
	UntilInLive
	Decode
)
