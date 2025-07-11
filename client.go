package sse

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
)

// EventHandler 事件处理回调函数
type EventHandler func(event *Event) error

// Client 是 SSE 客户端
type Client struct {
	req    *http.Request
	client *http.Client

	ctx    context.Context
	cancel context.CancelFunc

	eventHandler EventHandler
}

type Option func(opts *Client)

func WithClient(client *http.Client) Option {
	return func(opts *Client) {
		if client != nil {
			opts.client = client
		}
	}
}

// NewClient 创建 SSE 客户端
func NewClient(req *http.Request, opts ...Option) *Client {
	ctx, cancel := context.WithCancel(context.Background())

	var client = &Client{}
	client.req = req
	client.client = http.DefaultClient
	for _, opt := range opts {
		if opt != nil {
			opt(client)
		}
	}
	client.ctx = ctx
	client.cancel = cancel
	return client
}

// OnEvent 设置事件处理回调函数
func (c *Client) OnEvent(handler EventHandler) {
	c.eventHandler = handler
}

func (c *Client) Connect() error {
	var req = c.req.Clone(c.ctx)
	req.Header.Set("Accept", "text/event-stream")
	req.Header.Set("Cache-Control", "no-cache")
	req.Header.Set("Connection", "keep-alive")

	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	return c.handleResponse(resp)
}

// dispatchEvent 分发事件到处理器（如果事件有效且处理器存在），返回 error
func (c *Client) dispatchEvent(event *Event) error {
	if event != nil && (event.Data != "" || event.Event != "" || event.ID != "") {
		if c.eventHandler != nil {
			return c.eventHandler(event)
		}
	}
	return nil
}

func (c *Client) handleResponse(resp *http.Response) error {
	var reader = bufio.NewReader(resp.Body)
	var currentEvent *Event

	for {
		select {
		case <-c.ctx.Done():
			return nil
		default:
		}

		line, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				// 处理最后一个事件
				if err = c.dispatchEvent(currentEvent); err != nil {
					return err
				}
				return nil
			}
			return err
		}

		// 去除行尾的 \r\n 或 \n
		line = strings.TrimRight(line, "\r\n")

		// 空行表示事件结束，发送当前事件
		if line == "" {
			if err = c.dispatchEvent(currentEvent); err != nil {
				return err
			}
			// 重置当前事件
			currentEvent = nil
			continue
		}

		if strings.HasPrefix(line, ":") {
			// 注释行，跳过
			continue
		}

		// 检查是否为有效的SSE行格式
		var parts = strings.SplitN(line, ":", 2)
		if len(parts) != 2 {
			continue // 忽略无效行
		}

		// 懒初始化事件对象
		if currentEvent == nil {
			currentEvent = &Event{}
		}

		// 解析事件行并累积到当前事件
		c.parseEventLine(parts[0], parts[1], currentEvent)
	}
}

// parseEventLine 解析单行事件数据并累积到事件中
func (c *Client) parseEventLine(field, value string, event *Event) {
	field = strings.TrimSpace(field)
	value = strings.TrimSpace(value)

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
		// 忽略未知字段
	}
}

// Close 关闭连接
func (c *Client) Close() error {
	c.cancel()
	return nil
}
