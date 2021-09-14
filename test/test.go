package test

var RedisOption map[string]string
var RabbitOption map[string]string

func init() {
	RedisOption = map[string]string{
		"Address":  "localhost:6379",
		"Username": "",
		"Password": "",
	}

	RabbitOption = map[string]string{
		"Server":   "9.134.114.13:5672",
		"Username": "muxing",
		"Password": "muxing",
		"Exchange": "muxing",
		"Vhost":    "muxing",
	}
}
