package io.sanford.wormhole_william.ui

import org.junit.Assert.assertEquals
import org.junit.Assert.assertNull
import org.junit.Test

class QRScannerTest {

    // Tests for existing wormhole: URI format
    // Format: wormhole:ws://relay.magic-wormhole.io:4000/v1?code=5-souvenir-scallion

    @Test
    fun `parseWormholeUri with valid wormhole URI returns code`() {
        val uri = "wormhole:ws://relay.magic-wormhole.io:4000/v1?code=5-souvenir-scallion"
        assertEquals("5-souvenir-scallion", parseWormholeUri(uri))
    }

    @Test
    fun `parseWormholeUri with different relay returns code`() {
        val uri = "wormhole:ws://custom.relay.com:8080/v2?code=123-test-code"
        assertEquals("123-test-code", parseWormholeUri(uri))
    }

    @Test
    fun `parseWormholeUri with plain code containing dashes returns code`() {
        val code = "5-souvenir-scallion"
        assertEquals("5-souvenir-scallion", parseWormholeUri(code))
    }

    @Test
    fun `parseWormholeUri with plain code with whitespace returns trimmed code`() {
        val code = "  5-souvenir-scallion  "
        assertEquals("5-souvenir-scallion", parseWormholeUri(code))
    }

    @Test
    fun `parseWormholeUri with missing code parameter returns null`() {
        val uri = "wormhole:ws://relay.magic-wormhole.io:4000/v1"
        assertNull(parseWormholeUri(uri))
    }

    @Test
    fun `parseWormholeUri with empty code parameter returns empty string`() {
        // Current behavior: returns empty string (trimmed)
        val uri = "wormhole:ws://relay.magic-wormhole.io:4000/v1?code="
        assertEquals("", parseWormholeUri(uri))
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
        assertEquals("7-123-456", parseWormholeUri(code))
    }
}
