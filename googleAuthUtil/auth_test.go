package googleAuthUtil

import "testing"

func TestGetSecret(t *testing.T) {
	secret, err := GetSecret()
	if err != nil {
		t.Fatal("error:", err.Error())
		return
	}

	t.Log(secret)
}

func TestAuth(t *testing.T) {
	secret := "BEWWZOPO6ANXHZD3O6IA2AKAMK6MH2TO"
	nowCode, err := GetNowAuthCode(secret)
	if err != nil {
		t.Fatal("error:", err.Error())
		return
	}

	t.Log(nowCode)
}
