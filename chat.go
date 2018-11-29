package chat

import (
	"container/list"
	"strings"
	"sync"
)

const (
	UserSplit  = "\x00"
	MsgSep     = "\x1f"
	InfoPrefix = "info:"
	MsgPrefix  = "msg:"
	ListPrefix = "list:"
)

type Chat struct {
	users *list.List
	sync.RWMutex
}

type User struct {
	name    string
	receive chan string
	send    chan string
}

func New() *Chat {
	return &Chat{users: list.New()}
}

func (c *Chat) AddUser(name string) (chan<- string, <-chan string) {
	c.Lock()
	defer c.Unlock()

	send := make(chan string)
	receive := make(chan string)

	el := c.users.PushBack(&User{name, receive, send})

	c.broadcast(c.getInfo(name+" enter") + MsgSep + c.GetList())

	go func() {
		for text := range send {
			c.broadcast(c.getMsg(name, text))
		}
		c.Lock()
		c.users.Remove(el)
		c.Unlock()
		close(receive)
		c.broadcast(c.getInfo(name + " leave"))
	}()
	return send, receive
}

func (c *Chat) GetList() string {
	list := c.userList()
	return ListPrefix + strings.Join(list, UserSplit)
}

func (c *Chat) broadcast(content string) {
	for el := c.users.Front(); el != nil; el = el.Next() {
		go func(element *list.Element) {
			element.Value.(*User).receive <- content
		}(el)
	}
}

func (c *Chat) getMsg(name, text string) string {
	return MsgPrefix + name + ":" + text
}
func (c *Chat) getInfo(text string) string {
	return InfoPrefix + text
}

func (c *Chat) userList() (list []string) {
	for el := c.users.Front(); el != nil; el = el.Next() {
		list = append(list, el.Value.(*User).name)
	}
	return
}
