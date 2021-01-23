package aikucun_test

import (
	"os"
	"strings"
	"testing"

	"github.com/chekun/aikucun"
)

func setUpClient(t *testing.T) *aikucun.Client {
	appID := os.Getenv("APP_ID")
	appSecret := os.Getenv("APP_SECRET")
	gateway := os.Getenv("APP_GATEWAY")
	if appID == "" || appSecret == "" || gateway == "" {
		t.Fatalf("app_id, app_secret, gateway must be provided\n")
	}
	return aikucun.NewClient(appID, appSecret, gateway, nil)
}

func TestRegisterDistributor(t *testing.T) {
	client := setUpClient(t)
	phone := os.Getenv("TEST_PHONE")
	_, err := client.RegisterDistributor(phone, "测试")
	if err != nil && !strings.Contains(err.Error(), "已经注册过") {
		t.Fatalf("failed to get distributorID: %+v\n", err)
	}
	t.Log("register ok")
}

func TestGetAutoLoginURL(t *testing.T) {
	client := setUpClient(t)
	phone := os.Getenv("TEST_PHONE")
	url, err := client.GetAutoLoginURL(phone)
	if err != nil {
		t.Fatalf("failed to get auto login url: %+v\n", err)
	}
	t.Logf("got login url: %s", url)
}

func TestGetOrders(t *testing.T) {
	client := setUpClient(t)
	_, err := client.GetOrders(1, 20, "2021-01-01 00:00:00", "2021-02-01 00:00:00")
	if err != nil {
		t.Fatalf("failed to get orders: %+v\n", err)
	}
	t.Log("order req ok")
}

func TestGetOrderSettleInfo(t *testing.T) {
	client := setUpClient(t)
	_, err := client.GetOrderSettleInfo("20210107010153657795")
	if err != nil {
		t.Fatalf("failed to get order settle info: %+v\n", err)
	}
	t.Log("order settle info ok")
}
