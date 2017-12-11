package gotalk

//GtMan Structure to hold the Manager
type GtMan struct {
	localTCPServerName string
	localWSServerName  string

	handlers Handlers
}

//NewManager Creates a New Manager for gotalk Connections
func NewManager() GtMan {
	return GtMan{localTCPServerName: "", localWSServerName: ""}
}
//AddHandler Add a new Handler in this server
func (gm GtMan) AddHandler(op string, interface{}) error {
gm.handlers.Handle(op , fn )
}
