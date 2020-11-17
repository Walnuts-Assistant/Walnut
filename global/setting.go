package global

import (
	"Walnut/internal/service"
)

//CommonSetting ...
type CommonSetting struct {
	//ServerHost can be domain name or an ip address
	//(but port number must be carried)
	//e.g. https://sh1luo.gitee.io/
	//	   http://127.0.0.1:8080/
	ServerHost string

	//Manager ...
	Manager *service.Manager
}


