package io.sanford.wormhole_william.ui

import org.junit.Assert.assertEquals
import org.junit.Assert.assertNull
import org.junit.Test

class QRScannerTest {

    // Tests for existing wormhole: URI format
    // Format: wormhole:ws://relay.magic-wormhole.io:4000/v1?code=5-souvenir-scallion

    @Test
    fun `parseWormholeUri with valid wormhole URI returns code and relay`() {
        val uri = "wormhole:ws://relay.magic-wormhole.io:4000/v1?code=5-souvenir-scallion"
        val result = parseWormholeUri(uri)
        assertEquals("5-souvenir-scallion", result?.code)
        assertEquals("ws://relay.magic-wormhole.io:4000/v1", result?.rendezvousUrl)
    }

    @Test
    fun `parseWormholeUri with different relay returns code and relay`() {
        val uri = "wormhole:ws://custom.relay.com:8080/v2?code=123-test-code"
        val result = parseWormholeUri(uri)
        assertEquals("123-test-code", result?.code)
        assertEquals("ws://custom.relay.com:8080/v2", result?.rendezvousUrl)
    }

    @Test
    fun `parseWormholeUri with plain code containing dashes returns code with null relay`() {
        val code = "5-souvenir-scallion"
        val result = parseWormholeUri(code)
        assertEquals("5-souvenir-scallion", result?.code)
        assertNull(result?.rendezvousUrl)
    }

    @Test
    fun `parseWormholeUri with plain code with whitespace returns trimmed code`() {
        val code = "  5-souvenir-scallion  "
        val result = parseWormholeUri(code)
        assertEquals("5-souvenir-scallion", result?.code)
        assertNull(result?.rendezvousUrl)
    }

    @Test
    fun `parseWormholeUri with missing code parameter returns null`() {
        val uri = "wormhole:ws://relay.magic-wormhole.io:4000/v1"
        assertNull(parseWormholeUri(uri))
    }

    @Test
    fun `parseWormholeUri with empty code parameter returns null`() {
        val uri = "wormhole:ws://relay.magic-wormhole.io:4000/v1?code="
        assertNull(parseWormholeUri(uri))
    }

    @Test
    fun `parseWormholeUri with https URL returns null`() {
        val uri = "https://relay.example.com?code=123-456-789"
        assertNull(parseWormholeUri(uri))
    }

    @Test
    fun `parseWormholeUri with random string without dashes returns null`() {
        val input = "randomstringwithoutdashes"
        assertNull(parseWormholeUri(input))
    }

    @Test
    fun `parseWormholeUri with URL-like string without wormhole prefix returns null`() {
        val input = "ws://relay.example.com/v1"
        assertNull(parseWormholeUri(input))
    }

    @Test
    fun `parseWormholeUri with numeric code returns code`() {
        val code = "7-123-456"
        val result = parseWormholeUri(code)
        assertEquals("7-123-456", result?.code)
        assertNull(result?.rendezvousUrl)
    }

    // Tests for wormhole:// URI format (from Go tests)
    // Format: wormhole://relay.example.com?code=123-456-789

    @Test
    fun `parseWormholeUri with wormhole double-slash URI returns code and relay`() {
        val uri = "wormhole://relay.example.com?code=123-456-789"
        val result = parseWormholeUri(uri)
        assertEquals("123-456-789", result?.code)
        assertEquals("relay.example.com", result?.rendezvousUrl)
    }

    @Test
    fun `parseWormholeUri with wormhole double-slash and special chars in code`() {
        val uri = "wormhole://relay.example.com?code=123-abc_XYZ"
        val result = parseWormholeUri(uri)
        assertEquals("123-abc_XYZ", result?.code)
        assertEquals("relay.example.com", result?.rendezvousUrl)
    }

    @Test
    fun `parseWormholeUri with wormhole double-slash and additional parameters`() {
        val uri = "wormhole://relay.example.com?other=value&code=123-456&extra=param"
        val result = parseWormholeUri(uri)
        assertEquals("123-456", result?.code)
        assertEquals("relay.example.com", result?.rendezvousUrl)
    }

    @Test
    fun `parseWormholeUri with wormhole double-slash missing code returns null`() {
        val uri = "wormhole://relay.example.com"
        assertNull(parseWormholeUri(uri))
    }

    @Test
    fun `parseWormholeUri with wormhole double-slash empty code returns null`() {
        val uri = "wormhole://relay.example.com?code="
        assertNull(parseWormholeUri(uri))
    }

    @Test
    fun `parseWormholeUri with wormhole double-slash invalid URL returns null`() {
        val uri = "wormhole://relay with spaces.com?code=123"
        assertNull(parseWormholeUri(uri))
    }

    // Tests for wormhole-transfer: URI format (from Go tests)
    // Format: wormhole-transfer:4-hurricane-equipment?rendezvous=...

    @Test
    fun `parseWormholeUri with basic wormhole-transfer returns code with null relay`() {
        val uri = "wormhole-transfer:4-hurricane-equipment"
        val result = parseWormholeUri(uri)
        assertEquals("4-hurricane-equipment", result?.code)
        assertNull(result?.rendezvousUrl)
    }

    @Test
    fun `parseWormholeUri with wormhole-transfer and custom rendezvous`() {
        val uri = "wormhole-transfer:4-hurricane-equipment?rendezvous=ws%3A%2F%2Fcustom.relay.com%3A4000"
        val result = parseWormholeUri(uri)
        assertEquals("4-hurricane-equipment", result?.code)
        assertEquals("ws://custom.relay.com:4000", result?.rendezvousUrl)
    }

    @Test
    fun `parseWormholeUri with wormhole-transfer and percent-encoded code`() {
        val uri = "wormhole-transfer:4-hurricane%20equipment"
        val result = parseWormholeUri(uri)
        assertEquals("4-hurricane equipment", result?.code)
        assertNull(result?.rendezvousUrl)
    }

    @Test
    fun `parseWormholeUri with wormhole-transfer and multiple parameters`() {
        val uri = "wormhole-transfer:4-hurricane-equipment?rendezvous=ws%3A%2F%2Fcustom.relay.com%3A4000&role=leader&version=0"
        val result = parseWormholeUri(uri)
        assertEquals("4-hurricane-equipment", result?.code)
        assertEquals("ws://custom.relay.com:4000", result?.rendezvousUrl)
    }
}
