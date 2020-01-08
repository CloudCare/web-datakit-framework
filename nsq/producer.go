package nsq

import (
	"fmt"

	"github.com/nsqio/go-nsq"
)

type Producer struct {
	RemoteAddress    string   `json:"remote_address"`
	Hostname         string   `json:"hostname"`
	BroadcastAddress string   `json:"broadcast_address"`
	TCPPort          int      `json:"tcp_port"`
	HTTPPort         int      `json:"http_port"`
	Version          string   `json:"version"`
	Topics           []string `json:"topics"`

	// TODO: https://nsq.io/components/nsqlookupd.html#deletion-and-tombstones
	Tombstones []bool `json:"tombstones"`

	pub *nsq.Producer
	cfg *nsq.Config
}

func (p *Producer) Pub(topic string, msg []byte) error {
	var err error

	if p.pub == nil {
		p.cfg = nsq.NewConfig()
		p.pub, err = nsq.NewProducer(fmt.Sprintf("%s:%d", p.BroadcastAddress, p.TCPPort), p.cfg)
		if err != nil {
			return err
		}
	}

	if err := p.pub.Publish(topic, msg); err != nil {
		return err
	}

	return nil
}

func (p *Producer) Stop() {
	p.Stop()
}
