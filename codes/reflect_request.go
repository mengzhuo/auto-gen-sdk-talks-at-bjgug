// START OMIT
type StartRequest struct {
	Region string `xc:"required`
	HostId string `xc:"required`
}

// END OMIT

type StartResponse struct {
	Action string
	HostId string
}
