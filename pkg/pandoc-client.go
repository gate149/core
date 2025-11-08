package pkg

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

type Client struct {
	client  *http.Client
	address string
}

type PandocClient interface {
	ConvertLatexToHtml5(ctx context.Context, text string) (string, error)
	BatchConvertLatexToHtml5(ctx context.Context, texts []string) ([]string, error)
}

func NewPandocClient(client *http.Client, address string) *Client {
	return &Client{
		client:  client,
		address: address,
	}
}

type conversation struct {
	Text string `json:"text"`
	From string `json:"from"`
	To   string `json:"to"`
	Math string `json:"html-math-method"`
}

type message struct {
	Message   string `json:"message"`
	Verbosity string `json:"verbosity"`
}

type output struct {
	Error    string    `json:"error"`
	Output   string    `json:"output"`
	Base64   bool      `json:"base64"`
	Messages []message `json:"messages"`
}

func (client *Client) sendRaw(ctx context.Context, path string, body []byte) ([]byte, error) {
	path, err := url.JoinPath(client.address, path)
	if err != nil {
		return nil, err
	}

	buf := bytes.NewBuffer(body)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, path, buf)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.client.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	body, err = io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return body, nil
}

func (client *Client) convert(ctx context.Context, text, from, to, math string) (string, error) {
	body, err := json.Marshal(conversation{
		Text: text,
		From: from,
		To:   to,
		Math: math,
	})
	if err != nil {
		return "", err
	}

	resp, err := client.sendRaw(ctx, "/", body)
	if err != nil {
		return "", err
	}

	return string(resp), nil
}

func (client *Client) batchConvert(ctx context.Context, texts []string, from, to, math string) ([]string, error) {
	list := make([]conversation, len(texts))
	for i, text := range texts {
		list[i] = conversation{
			Text: text,
			From: from,
			To:   to,
			Math: math,
		}
	}

	body, err := json.Marshal(list)
	if err != nil {
		return nil, err
	}

	resp, err := client.sendRaw(ctx, "/batch", body)
	if err != nil {
		return nil, err
	}

	var result []output
	err = json.Unmarshal(resp, &result)
	if err != nil {
		return nil, err
	}

	if len(result) != len(texts) {
		return nil, fmt.Errorf("wrong number of fieilds returned: %d", len(result))
	}

	err = nil
	for _, o := range result {
		if o.Error != "" {
			err = errors.Join(err, errors.New(o.Error))
		}
	}

	if err != nil {
		return nil, Wrap(ErrBadInput, err, "BatchConvertLatexToHtml5", "invalid input")
	}

	res := make([]string, len(result))
	for i, o := range result {
		res[i] = o.Output
	}

	return res, nil
}

func (client *Client) ConvertLatexToHtml5(ctx context.Context, text string) (string, error) {
	return client.convert(ctx, text, "latex", "html5", "katex")
}

func (client *Client) BatchConvertLatexToHtml5(ctx context.Context, texts []string) ([]string, error) {
	return client.batchConvert(ctx, texts, "latex", "html5", "katex")
}
