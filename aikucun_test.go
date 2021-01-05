package aikucun_test

import (
	"os"
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

func TestGetAutoLoginURL(t *testing.T) {
	client := setUpClient(t)
	phone := os.Getenv("TEST_PHONE")
	url, err := client.GetAutoLoginURL(phone)
	if err != nil {
		t.Fatalf("failed to get auto login url: %+v\n", err)
	}
	t.Logf("got login url: %s", url)
}