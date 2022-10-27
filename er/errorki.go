package er

type Type uint16

// ErrMissReqID = errors.New("invalid request - missing `RequestID`")
// ErrMissAction = errors.New("invalid request - missing `Action`")

const (
	OK                  Type = 200
	NotFound            Type = 100
	Missing             Type = 300
	Forbbiden           Type = 400
	Timeout             Type = 500
	InternalServerError Type = 1000
)

const (
	_ Type = iota
	ReqID
	Action
	ContainerName
	Container
	UntilInLive
)
