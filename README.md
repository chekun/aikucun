# 爱库存API

[![Build Status](https://travis-ci.com/chekun/aikucun.svg?branch=main)](https://travis-ci.com/chekun/aikucun)
[![Go Report Card](https://goreportcard.com/badge/github.com/chekun/aikucun)](https://goreportcard.com/report/github.com/chekun/aikucun)
[![Go Doc](https://godoc.org/github.com/chekun/aikucun?status.svg)](http://godoc.org/github.com/chekun/aikucun)

> 本仓库目前为自用，不保证满足所有人需求。

## 快速使用

```go
// 创建client
client := aikucun.NewClient(appID, appSecret, gateway, nil)
// 调用接口
autoURL, err := client.GetAutoLoginURL("18888888888")
// 查看结果
fmt.Println(autoURL, err)
```

## 支持的接口列表

1. 会员三方登录 (aikucun.member.open.third.login)

    ```go
    GetAutoLoginURL(phone string) (string, error)
    ```

