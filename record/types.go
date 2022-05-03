package record

type ContainerInfo struct {
	Pid         string `json:"pid"`
	Id          string `json:"id"`
	Name        string `json:"name"`
	Command     string `json:"command"`
	Volume      string `json:"volume"`
	CreatedTime string `json:"created_time"`
	Status      string `json:"status"`
	PortMapping []string `json:"port_mapping"` //端口映射
}
