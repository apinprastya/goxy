package server

type BasicAuth struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type VHost struct {
	Domain         []string   `json:"domain"`
	Target         string     `json:"target"`
	AllowNonSecure bool       `json:"allowNonSecure"`
	BasicAuth      *BasicAuth `json:"basicAuth"`
}

type Config struct {
	Cache  string  `json:"cache"`
	VHosts []VHost `json:"vhosts"`
}

func (c *Config) AllVHost() []string {
	ret := []string{}
	for _, v := range c.VHosts {
		ret = append(ret, v.Domain...)
	}
	return ret
}
