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
	Code    interface{}     `json:"code"` //can be string or integer, weird!
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
	now := time.Now()
	newParams := map[string]string{
		"appid":         c.appID,
		"appsecret":     c.appSecret,
		"noncestr":      now.Format("150405"),
		"timestamp":     fmt.Sprintf("%d", now.Unix()),
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

// RegisterDistributor 店长注册
func (c *Client) RegisterDistributor(phone, name string) (string, error) {
	params := map[string]string{
		"accessToken": "",
	}
	postBody := map[string]interface{}{
		"phone": phone,
		"name":  name,
	}
	qs, bodyBytes := c.signParams("aikucun.member.open.register.distributor", params, postBody)

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
	var distributor uint64
	_ = json.Unmarshal(r.Data, &distributor)
	return fmt.Sprintf("%d", distributor), nil
}

// OrderItem 三方订单结构体
type OrderItem struct {
	OrderNo                 string  `json:"orderNo"`
	OrderDetailNo           string  `json:"orderDetailNo"`
	OrderDetailID           string  `json:"orderDetailId"`
	ProductID               string  `json:"productId"`
	SkuID                   string  `json:"skuId"`
	ProductName             string  `json:"productName"`
	Barcode                 string  `json:"barcode"`
	ModelNo                 string  `json:"modelNo"`
	AfterSaleStatus         int     `json:"afterSaleStatus"`
	PayStatus               int     `json:"payStatus"`
	ThreeOrderPaymentAmount float64 `json:"threeOrderPaymentAmount"`
}

// Order 二方订单结构体
type Order struct {
	ShopNo          int64        `json:"shopNo"`
	SellerID        string       `json:"sellerId"`
	OrderNo         string       `json:"orderNo"`
	BrandID         string       `json:"brandId"`
	BrandURL        string       `json:"brandUrl"`
	BrandName       string       `json:"brandName"`
	PaymentAmount   float64      `json:"paymentAmount"`
	OrderStatus     int          `json:"orderStatus"`
	OrderChannel    string       `json:"orderChannel"`
	OrderSource     string       `json:"orderSource"`
	TotalCommission float64      `json:"totalCommission"`
	OrderTime       string       `json:"orderTime"`
	Freight         float64      `json:"freight"`
	ThreeOrderList  []*OrderItem `json:"threeOrderList"`
}

// OrderResponse 订单接口返回数据结构
type OrderResponse struct {
	PageIndex int      `json:"pageIndex"`
	PageSize  int      `json:"pageSize"`
	StartRow  int      `json:"startRow"`
	EndRow    int      `json:"endRow"`
	Total     int      `json:"total"`
	Pages     int      `json:"pages"`
	Result    []*Order `json:"result"`
}

// 拉取订单
func (c *Client) GetOrders(page int, pageSize int, from, to string) (*OrderResponse, error) {
	params := map[string]string{
		"accessToken": "",
	}
	postBody := map[string]interface{}{
		"currentPage": page,
		"pageSize":    pageSize,
		"data": map[string]interface{}{
			"beginTime": from,
			"endTime":   to,
		},
	}
	qs, bodyBytes := c.signParams("aikucun.order.seller.order.list", params, postBody)
	req, err := c.makeRequest("POST", qs, bodyBytes)
	if err != nil {
		return nil, err
	}
	resBody, _, err := c.do(req)
	if err != nil {
		return nil, err
	}
	var r Response
	err = json.Unmarshal(resBody, &r)
	if err != nil {
		return nil, err
	}
	if !r.IsSuccessful() {
		return nil, r.Error()
	}
	var orderRes OrderResponse
	_ = json.Unmarshal(r.Data, &orderRes)
	return &orderRes, nil
}

// 结算订单数据结构
type OrderSettleInfo struct {
	IncomeAmount float64 `json:"incomeAmount"`
	SettleDate   string  `json:"settle_date"`
	SettleStatus string  `json:"settleStatus"`
}

// 按照订单号拉取订单结算时间
func (c *Client) GetOrderSettleInfo(orderNo string) (*OrderSettleInfo, error) {
	params := map[string]string{
		"accessToken":   "",
		"secondOrderNo": orderNo,
	}
	qs, bodyBytes := c.signParams("aikucun.settle.shop.income.detail", params, nil)
	req, err := c.makeRequest("GET", qs, bodyBytes)
	if err != nil {
		return nil, err
	}
	resBody, _, err := c.do(req)
	if err != nil {
		return nil, err
	}
	var r Response
	err = json.Unmarshal(resBody, &r)
	if err != nil {
		return nil, err
	}
	if !r.IsSuccessful() {
		return nil, r.Error()
	}
	var info OrderSettleInfo
	_ = json.Unmarshal(r.Data, &info)
	return &info, nil
}

func (c *Client) makeRequest(method string, params string, body []byte) (*http.Request, error) {
	toURL := c.apiGateway + "?" + params
	fmt.Println(toURL, string(body))
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
