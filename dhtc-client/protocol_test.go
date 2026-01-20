package dhtc_client

import (
	"crypto/rand"
	"net"
	"testing"
)

func TestVerifyToken(t *testing.T) {
	p := NewProtocol(":0", 100, ProtocolEventHandlers{})
	addr := net.ParseIP("127.0.0.1")
	token := p.CalculateToken(addr)

	if !p.VerifyToken(addr, token) {
		t.Error("VerifyToken failed for current token")
	}

	// Test with wrong address
	wrongAddr := net.ParseIP("127.0.0.2")
	if p.VerifyToken(wrongAddr, token) {
		t.Error("VerifyToken should fail for wrong address")
	}

	// Test with wrong token
	wrongToken := make([]byte, len(token))
	copy(wrongToken, token)
	wrongToken[0] ^= 0xFF
	if p.VerifyToken(addr, wrongToken) {
		t.Error("VerifyToken should fail for wrong token")
	}

	// Test with previous token
	p.tokenLock.Lock()
	copy(p.previousTokenSecret, p.currentTokenSecret)
	_, _ = rand.Read(p.currentTokenSecret)
	p.tokenLock.Unlock()

	if !p.VerifyToken(addr, token) {
		t.Error("VerifyToken failed for previous token")
	}

	newToken := p.CalculateToken(addr)
	if !p.VerifyToken(addr, newToken) {
		t.Error("VerifyToken failed for new current token")
	}
}
