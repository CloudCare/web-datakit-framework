package nsq

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"math/rand"
	"net/http"
)

type NSQLookupd struct {
	host          string              `json:"-"`
	c             *http.Client        `json:"-"`
	topicChannels map[string][]string `json:"-"`

	Topics    []string    `json:"topics,omitempty"`
	Channels  []string    `json:"channels,omitempty"`
	Producers []*Producer `json:"producers,omitempty"`
}

func New(host string) *NSQLookupd {
	return &NSQLookupd{
		host: host,
		c:    &http.Client{},

		topicChannels: map[string][]string{},
	}
}

func (d *NSQLookupd) Refresh() error {
	if err := d.refreshNodes(); err != nil {
		return err
	}

	if err := d.refreshTopics(); err != nil {
		return err
	}

	for _, t := range d.Topics {
		if err := d.refreshChannels(t); err != nil {
			return err
		}
	}

	return nil
}

func (d *NSQLookupd) RandomNode() *Producer {
	idx := rand.Intn(len(d.Producers))
	return d.Producers[idx]
}

func (d *NSQLookupd) refreshNodes() error {
	body, err := d.apiReq(`GET`, `nodes`, nil)
	if err != nil {
		return err
	}

	if err := json.Unmarshal(body, d); err != nil {
		return err
	}

	return nil
}

func (d *NSQLookupd) refreshTopics() error {
	body, err := d.apiReq(`GET`, `topics`, nil)
	if err != nil {
		return err
	}

	if err := json.Unmarshal(body, d); err != nil {
		return err
	}

	return nil
}

func (d *NSQLookupd) refreshChannels(topic string) error {
	body, err := d.apiReq(`GET`, `channels?topic=`+topic, nil)
	if err != nil {
		return err
	}

	if err := json.Unmarshal(body, d); err != nil {
		return err
	}

	d.topicChannels[topic] = d.Channels
	return nil
}

func (d *NSQLookupd) apiReq(method, path string, r io.Reader) ([]byte, error) {
	url := fmt.Sprintf("%s/%s", d.host, path)

	req, err := http.NewRequest(method, url, r)
	if err != nil {
		return nil, err
	}

	resp, err := d.c.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, errors.New(string(body))
	}

	return body, nil
}
