package model

import "github.com/google/uuid"

type Client struct {
	UUID string
	Who string
	Send chan *Message
	Tags []string
}

func NewClient(Who string) *Client {
	return &Client{
		UUID: uuid.NewString(),
		Who: Who,
		Send: make(chan *Message, 16),
	}
}

func (c *Client) Contains(e string) bool {
    for _, a := range c.Tags {
        if a == e {
            return true
        }
    }
    return false
}