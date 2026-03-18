package protocol

import "testing"

func TestDecodeSSURLWithPlainUserinfo(t *testing.T) {
	url := "ss://2022-blake3-aes-128-gcm:%2FS8blFRGE3o%2FaDSN93iTmA%3D%3D@3.115.244.62:18898#JP-AWS"

	ss, err := DecodeSSURL(url)
	if err != nil {
		t.Fatalf("decode failed: %v", err)
	}

	assertEqualString(t, "Server", "3.115.244.62", ss.Server)
	assertEqualIntInterface(t, "Port", 18898, ss.Port)
	assertEqualString(t, "Cipher", "2022-blake3-aes-128-gcm", ss.Param.Cipher)
	assertEqualString(t, "Password", "/S8blFRGE3o/aDSN93iTmA==", ss.Param.Password)
	assertEqualString(t, "Name", "JP-AWS", ss.Name)
}

func TestDecodeSSURLWithPlainUserinfoContainingColon(t *testing.T) {
	url := "ss://2022-blake3-aes-256-gcm:n3RRAL3KF%2FzeWa1O722wd9UlNR%2BGVgSEgjeujdImVds%3D%3AH9jQ%2B6AabhJDhnu6NeuIk4IaGjNEnTj2TxzXQ9Sg9lI%3D@3.115.244.62:40000#US-Hawaii-AWS"

	ss, err := DecodeSSURL(url)
	if err != nil {
		t.Fatalf("decode failed: %v", err)
	}

	assertEqualString(t, "Server", "3.115.244.62", ss.Server)
	assertEqualIntInterface(t, "Port", 40000, ss.Port)
	assertEqualString(t, "Cipher", "2022-blake3-aes-256-gcm", ss.Param.Cipher)
	assertEqualString(t, "Password", "n3RRAL3KF/zeWa1O722wd9UlNR+GVgSEgjeujdImVds=:H9jQ+6AabhJDhnu6NeuIk4IaGjNEnTj2TxzXQ9Sg9lI=", ss.Param.Password)
	assertEqualString(t, "Name", "US-Hawaii-AWS", ss.Name)
}
