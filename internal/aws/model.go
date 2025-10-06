package aws

type Instance struct {
	InstanceId  string `json:"instance_id"`
	Name        string `json:"name"`
	ServerGroup string `json:"server_group"`
	TimeAdded   int64  `json:"time_added"`
}
