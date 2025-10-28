package ui

import (
	"net/url"
	"testing"
)

const expectedCode = "7-some-code"

func TestParsingInvalidUri(t *testing.T) {
	uri := "not a valid code"

	parsed, err := parseCodeURI(uri)

	if parsed != nil || err == nil {
		t.Errorf("expected error, got %s, %v", parsed, err)
	}
}

func TestParsingWormholeWilliamUri(t *testing.T) {
	uri := "wormhole:ws://relay.magic-wormhole.io:4000/v1?code=" + expectedCode

	parsed, err := parseCodeURI(uri)

	if parsed.code != expectedCode || err != nil {
		t.Errorf("got %s, %v, expected %s", parsed.code, err, expectedCode)
	}
}

func TestParsingMagicWormholeUri(t *testing.T) {
	uri := "wormhole-transfer:" + expectedCode

	parsed, err := parseCodeURI(uri)

	if parsed.code != expectedCode || err != nil {
		t.Errorf("got %s, %v, expected %s", parsed.code, err, expectedCode)
	}
}

func TestParsingMagicWormholeUriPercentEncoded(t *testing.T) {
	unencodedCode := "8-éè-code"
	percentEncodedCode := url.QueryEscape(unencodedCode)
	if unencodedCode == percentEncodedCode { // sanity check
		t.Errorf("Percent encoded code should not be the same as unencoded")
	}
	uri := "wormhole-transfer:" + percentEncodedCode

	parsed, err := parseCodeURI(uri)

	if parsed.code != unencodedCode || err != nil {
		t.Errorf("got %s, %v expected %s", parsed.code, err, unencodedCode)
	}
}

func TestParsingMagicWormholeUriWithQueryParams(t *testing.T) {
	uri := "wormhole-transfer:" + expectedCode + "?version=0&rendezvous=ws%3A%2F%2Frelay.magic-wormhole.io%3A4000&role=follower"

	parsed, err := parseCodeURI(uri)

	if parsed.code != expectedCode || err != nil {
		t.Errorf("got %s, %v, expected %s", parsed.code, err, expectedCode)
	}
}
