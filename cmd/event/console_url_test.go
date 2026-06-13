// Copyright (c) 2026 Lark Technologies Pte. Ltd.
// SPDX-License-Identifier: MIT

package event

import (
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"encoding/json"
	"io"
	"strings"
	"testing"

	"github.com/larksuite/cli/internal/core"
	eventlib "github.com/larksuite/cli/internal/event"
)

func decodeAddons(t *testing.T, encoded string) ManifestAddons {
	t.Helper()
	gz, err := base64.RawURLEncoding.DecodeString(encoded)
	if err != nil {
		t.Fatalf("base64url decode: %v", err)
	}
	zr, err := gzip.NewReader(bytes.NewReader(gz))
	if err != nil {
		t.Fatalf("gzip reader: %v", err)
	}
	raw, err := io.ReadAll(zr)
	if err != nil {
		t.Fatalf("gunzip: %v", err)
	}
	var a ManifestAddons
	if err := json.Unmarshal(raw, &a); err != nil {
		t.Fatalf("json: %v", err)
	}
	return a
}

func TestEncodeAddons_RoundTrip(t *testing.T) {
	in := ManifestAddons{Scopes: &AddonsScopes{Tenant: []string{"im:message"}}}
	encoded, err := encodeAddons(in)
	if err != nil {
		t.Fatalf("encode: %v", err)
	}
	for _, r := range encoded {
		if !(r == '-' || r == '_' || (r >= '0' && r <= '9') || (r >= 'A' && r <= 'Z') || (r >= 'a' && r <= 'z')) {
			t.Fatalf("encoded contains non-base64url char %q in %q", r, encoded)
		}
	}
	out := decodeAddons(t, encoded)
	if out.Scopes == nil || len(out.Scopes.Tenant) != 1 || out.Scopes.Tenant[0] != "im:message" {
		t.Errorf("roundtrip mismatch: %+v", out)
	}
}

func TestConsoleAddonsURL_FormatAndBrandHost(t *testing.T) {
	url, err := consoleAddonsURL(core.BrandFeishu, "cli_x", ManifestAddons{Callbacks: &AddonsCallbacks{Items: []string{"card.action.trigger"}}})
	if err != nil {
		t.Fatalf("url: %v", err)
	}
	host := core.ResolveEndpoints(core.BrandFeishu).Open
	prefix := host + "/page/launcher?client_id=cli_x&addons="
	if !strings.HasPrefix(url, prefix) {
		t.Errorf("url = %q, want prefix %q", url, prefix)
	}
	out := decodeAddons(t, strings.TrimPrefix(url, prefix))
	if out.Callbacks == nil || len(out.Callbacks.Items) != 1 || out.Callbacks.Items[0] != "card.action.trigger" {
		t.Errorf("decoded callbacks mismatch: %+v", out)
	}
}

func TestMissingScopeAddons_ByIdentity(t *testing.T) {
	bot := missingScopeAddons(core.AsBot, []string{"im:message"})
	if bot.Scopes == nil || len(bot.Scopes.Tenant) != 1 || len(bot.Scopes.User) != 0 {
		t.Errorf("bot scopes = %+v, want tenant-only", bot.Scopes)
	}
	user := missingScopeAddons(core.AsUser, []string{"im:message"})
	if user.Scopes == nil || len(user.Scopes.User) != 1 || len(user.Scopes.Tenant) != 0 {
		t.Errorf("user scopes = %+v, want user-only", user.Scopes)
	}
}

func TestMissingSubscriptionAddons_EventVsCallback(t *testing.T) {
	ev := missingSubscriptionAddons(eventlib.SubTypeEvent, core.AsBot, []string{"im.message.receive_v1"})
	if ev.Events == nil || len(ev.Events.Items.Tenant) != 1 {
		t.Errorf("event addons = %+v, want events.items.tenant", ev.Events)
	}
	cb := missingSubscriptionAddons(eventlib.SubTypeCallback, core.AsBot, []string{"card.action.trigger"})
	if cb.Callbacks == nil || len(cb.Callbacks.Items) != 1 || cb.Events != nil {
		t.Errorf("callback addons = %+v, want callbacks.items only", cb)
	}
}
