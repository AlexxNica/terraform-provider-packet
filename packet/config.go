package packet

import (
	cleanhttp "github.com/hashicorp/go-cleanhttp"
	"github.com/hashicorp/terraform/helper/logging"
	"github.com/packethost/packngo"
)

const (
	consumerToken = "aZ9GmqHTPtxevvFq9SK3Pi2yr9YCbRzduCSXF2SNem5sjB91mDq7Th3ZwTtRqMWZ"
)

// Config types (currently only AuthToken set)
type Config struct {
	AuthToken string
}

// Client function returns a new client for accessing Packet's API.
func (c *Config) Client() *packngo.Client {
	client := cleanhttp.DefaultClient()
	client.Transport = logging.NewTransport("Packet", client.Transport)
	return packngo.NewClient(consumerToken, c.AuthToken, client)
}
