package aikucun

import (
	"bytes"
	"crypto/sha1"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"time"
)

// Response 公共返回结构体
type Response struct {
	Code    string          `json:"code"`
	Success bool            `json:"success"`
	Message string          `json:"message"`
	Data    json.RawMessage `json:"data"`
}

// IsSuccessful 返回信息是否是成功的
func (res *Response) IsSuccessful() bool {
	return res.Success
}

// Error 转换错误为error对象
func (res *Response) Error() error {
	if res.Success {
		return nil
	}
	return fmt.Errorf("code:%s,%s", res.Code, res.Message)
}

// Client client
type Client struct {
	appID      string
	appSecret  string
	apiGateway string
	client     *http.Client
}

// NewClient 创建client
func NewClient(appID, appSecret, apiGateway string, client *http.Client) *Client {
	if client == nil {
		client = defaultHTTPClient()
	}
	return &Client{
		appID:      appID,
		appSecret:  appSecret,
		apiGateway: apiGateway,
		client:     client,
	}
}

func defaultHTTPClient() *http.Client {
	tr := &http.Transport{
		DialContext: (&net.Dialer{
			Timeout:   5 * time.Second,
			KeepAlive: 30 * time.Second,
		}).DialContext,
		MaxConnsPerHost:     200,
		MaxIdleConnsPerHost: 30,
		IdleConnTimeout:     30 * time.Second,
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
	}
	return &http.Client{Transport: tr}
}

func (c *Client) signParams(router string, params map[string]string, body map[string]interface{}) (string, []byte) {
	newParams := map[string]string{
		"appid":         "be6c3ca0d9d2480a8e30eec88f7de475",
		"appsecret":     "6a53203e07e3445eaf4438cda92ea854",
		"noncestr":      "1",
		"timestamp":     fmt.Sprintf("%d", time.Now().Unix()),
		"version":       "1",
		"format":        "JSON",
		"interfaceName": router,
	}
	for k, v := range params {
		newParams[k] = v
	}
	bodyJSON := make([]byte, 0)
	if body != nil && len(body) > 0 {
		bodyJSON, _ = json.Marshal(body)
		newParams["body"] = string(bodyJSON)
	}
	keys := make([]string, 0)
	for k := range newParams {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	originConetnt := make([]string, 0)
	for _, k := range keys {
		originConetnt = append(originConetnt, k+"="+newParams[k])
	}
	h := sha1.New()
	_, _ = h.Write([]byte(strings.Join(originConetnt, "&")))
	newParams["sign"] = fmt.Sprintf("%x", h.Sum(nil))
	delete(newParams, "appsecret")
	delete(newParams, "body")
	values := url.Values{}
	for k, v := range newParams {
		values.Set(k, v)
	}
	return values.Encode(), bodyJSON
}

// GetAutoLoginURL 三方联合登录接口
func (c *Client) GetAutoLoginURL(phone string) (string, error) {
	params := map[string]string{
		"accessToken": "",
	}
	postBody := map[string]interface{}{
		"phone": phone,
		"scene": 1,
	}
	qs, bodyBytes := c.signParams("aikucun.member.open.third.login", params, postBody)

	req, err := c.makeRequest("POST", qs, bodyBytes)
	if err != nil {
		return "", err
	}
	resBody, _, err := c.do(req)
	if err != nil {
		return "", err
	}
	var r Response
	err = json.Unmarshal(resBody, &r)
	if err != nil {
		return "", err
	}
	if !r.IsSuccessful() {
		return "", r.Error()
	}
	return string(r.Data), nil
}

func (c *Client) makeRequest(method string, params string, body []byte) (*http.Request, error) {
	toURL := c.apiGateway + "?" + params
	r, err := http.NewRequest(method, toURL, bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}
	r.Header.Add("Content-Type", "application/json")
	return r, err
}

func (c *Client) do(req *http.Request) (body []byte, resp *http.Response, err error) {
	req.Close = true
	resp, err = c.client.Do(req)
	if resp != nil {
		defer resp.Body.Close()
	}
	if err != nil {
		return nil, nil, err
	}
	bodyWriter := &bytes.Buffer{}
	_, err = bodyWriter.ReadFrom(resp.Body)
	body = bodyWriter.Bytes()
	return
}
