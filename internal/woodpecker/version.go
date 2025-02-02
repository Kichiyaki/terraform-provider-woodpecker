package woodpecker

import "fmt"

const (
	pathVersion = "%s/version"
)

func (c *client) Version() (*Version, error) {
	out := new(Version)
	uri := fmt.Sprintf(pathVersion, c.addr)
	return out, c.get(uri, &out)
}
