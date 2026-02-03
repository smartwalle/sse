package sse

import (
	"bufio"
	"context"
	"errors"
	"io"
	"net/http"
	"strconv"
	"strings"
	"sync"
)

type EventHandler func(ctx context.Context, event *Event) error
type CheckHandler func(ctx context.Context, response *http.Response) error

var ErrHandlerNotFound = errors.New("event handler not found")

type Client struct {
	req          *http.Request
	client       *http.Client
	eventHandler EventHandler
	checkHandler CheckHandler
	closed       chan struct{}
	closeOnce    sync.Once
}

type Option func(opts *Client)

func WithClient(client *http.Client) Option {
	return func(opts *Client) {
		if client != nil {
			opts.client = client
		}
	}
}

func NewClient(req *http.Request, opts ...Option) *Client {
	var client = &Client{}
	client.req = req
	client.client = http.DefaultClient
	client.closed = make(chan struct{})
	for _, opt := range opts {
		if opt != nil {
			opt(client)
		}
	}
	return client
}

func (c *Client) Closed() bool {
	select {
	case <-c.closed:
		return true
	default:
		return false
	}
}

func (c *Client) Close() error {
	c.closeOnce.Do(func() {
		if c.closed != nil {
			close(c.closed)
		}
	})
	return nil
}

func (c *Client) OnEvent(handler EventHandler) {
	c.eventHandler = handler
}

func (c *Client) OnCheck(handler CheckHandler) {
	c.checkHandler = handler
}

func (c *Client) Connect(ctx context.Context) error {
	select {
	case <-c.closed:
		return io.EOF
	default:
	}

	if c.eventHandler == nil {
		return ErrHandlerNotFound
	}

	var req = c.req.Clone(ctx)
	req.Header.Set("Accept", "text/event-stream")
	req.Header.Set("Cache-Control", "no-cache")
	req.Header.Set("Connection", "keep-alive")

	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if c.checkHandler != nil {
		if err = c.checkHandler(ctx, resp); err != nil {
			return err
		}
	}

	return c.handleResponse(ctx, resp)
}

func (c *Client) dispatchEvent(ctx context.Context, event *Event) error {
	if c.eventHandler != nil && event != nil && (event.Data != "" || event.Event != "" || event.ID != "") {
		return c.eventHandler(ctx, event)
	}
	return nil
}

func (c *Client) handleResponse(ctx context.Context, resp *http.Response) error {
	var reader = bufio.NewReader(resp.Body)
	var currentEvent *Event

	defer func() {
		_ = c.Close()
	}()

	for {
		select {
		case <-c.closed:
			return io.EOF
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		line, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				// 处理最后一个事件
				if nErr := c.dispatchEvent(ctx, currentEvent); nErr != nil {
					return nErr
				}
			}
			return err
		}

		line = strings.TrimRight(line, "\r\n")

		if line == "" {
			if err = c.dispatchEvent(ctx, currentEvent); err != nil {
				return err
			}
			currentEvent = nil
			continue
		}

		if len(line) > 0 && line[0] == ':' {
			continue
		}

		var field, value string
		if idx := strings.Index(line, ":"); idx >= 0 {
			field = line[:idx]
			value = line[idx+1:]
		} else {
			field = line
			value = ""
		}

		if currentEvent == nil {
			currentEvent = &Event{}
		}

		c.parseEvent(currentEvent, field, value)
	}
}

func (c *Client) parseEvent(event *Event, field, value string) {
	field = strings.TrimSpace(field)
	value = strings.TrimPrefix(value, " ")

	// 验证字段名不为空
	if field == "" {
		return
	}

	switch field {
	case "id":
		event.ID = value
	case "event":
		event.Event = value
	case "data":
		if event.Data != "" {
			event.Data += "\n"
		}
		event.Data += value
	case "retry":
		if retry, err := strconv.Atoi(value); err == nil {
			event.Retry = retry
		}
	default:
	}
}
