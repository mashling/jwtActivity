package jwtActivity

import (
	"testing"
)

func TestJWT(t *testing.T) {
	settings := map[string]interface{}{
		"signingMethod": "HMAC",
		"key":           "qwertyuiopasdfghjklzxcvbnm789101",
		"aud":           "www.mashling.io",
		"iss":           "Mashling",
	}
	factory := Factory{}
	instance, err := factory.Make("jwtService", settings)
	if err != nil {
		t.Fatal(err)
	}
	err = instance.UpdateRequest(map[string]interface{}{
		"token": "eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9.eyJpc3MiOiJNYXNobGluZyIsImlhdCI6MTUzNjE4MTE4OSwiZXhwIjo0MTIzODYxMTg5LCJhdWQiOiJ3d3cubWFzaGxpbmcuaW8iLCJzdWIiOiJqcm9ja2V0QGV4YW1wbGUuY29tIn0.Zl4l68Z9VcuFXEFQt8kCH7fcaiMmRRGtrC28lSWvJWw",
	})
	if err != nil {
		t.Fatal(err)
	}
	err = instance.Execute()
	if err != nil {
		t.Fatal(err)
	}

	if !instance.(*JWT).Response.Valid {
		t.Fatal("JWT token should be valid")
	}
}
