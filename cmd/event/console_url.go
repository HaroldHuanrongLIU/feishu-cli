// Copyright (c) 2026 Lark Technologies Pte. Ltd.
// SPDX-License-Identifier: MIT

package event

import (
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/larksuite/cli/internal/core"
	eventlib "github.com/larksuite/cli/internal/event"
)

// consoleScopeGrantURL builds the developer-console "apply & grant scopes" deep link; scopes are comma-joined without URL encoding.
func consoleScopeGrantURL(brand core.LarkBrand, appID string, scopes []string) string {
	host := core.ResolveEndpoints(brand).Open
	return fmt.Sprintf("%s/app/%s/auth?q=%s&op_from=openapi&token_type=tenant",
		host, appID, strings.Join(scopes, ","))
}

// consoleEventSubscriptionURL points at the app's event subscription console page.
func consoleEventSubscriptionURL(brand core.LarkBrand, appID string) string {
	host := core.ResolveEndpoints(brand).Open
	return fmt.Sprintf("%s/app/%s/event", host, appID)
}

// Landing-page contract for the scan-to-enable deep link. Centralized so the
// path/param can be corrected in one place once confirmed with the open platform.
// NOTE: inferred from the addons spec doc (BVsOdgbRNosaeJxZzDTcVwP6njb); verify before release.
const (
	addonsLandingPath   = "/page/launcher"
	addonsClientIDParam = "client_id"
)

// ManifestAddons mirrors the 5 public manifest sections the launcher page accepts.
// Encoded form: JSON -> gzip -> base64url(no padding).
type ManifestAddons struct {
	Scopes    *AddonsScopes    `json:"scopes,omitempty"`
	Events    *AddonsEvents    `json:"events,omitempty"`
	Callbacks *AddonsCallbacks `json:"callbacks,omitempty"`
}

type AddonsScopes struct {
	Tenant []string `json:"tenant"`
	User   []string `json:"user"`
}

type AddonsEvents struct {
	Items AddonsEventItems `json:"items"`
}

type AddonsEventItems struct {
	Tenant []string `json:"tenant"`
	User   []string `json:"user"`
}

type AddonsCallbacks struct {
	Items []string `json:"items"`
}

// encodeAddons: JSON -> gzip -> base64url(no padding). Matches the front-end decode chain.
func encodeAddons(a ManifestAddons) (string, error) {
	raw, err := json.Marshal(a)
	if err != nil {
		return "", err
	}
	var buf bytes.Buffer
	gw := gzip.NewWriter(&buf)
	if _, err := gw.Write(raw); err != nil {
		return "", err
	}
	if err := gw.Close(); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(buf.Bytes()), nil
}

// consoleAddonsURL builds the scan-to-enable deep link carrying incremental scopes/events/callbacks.
func consoleAddonsURL(brand core.LarkBrand, appID string, a ManifestAddons) (string, error) {
	encoded, err := encodeAddons(a)
	if err != nil {
		return "", err
	}
	host := core.ResolveEndpoints(brand).Open
	return fmt.Sprintf("%s%s?%s=%s&addons=%s", host, addonsLandingPath, addonsClientIDParam, appID, encoded), nil
}

// consoleLandingURL is the bare landing page (no addons) — fallback when encoding fails.
func consoleLandingURL(brand core.LarkBrand, appID string) string {
	host := core.ResolveEndpoints(brand).Open
	return fmt.Sprintf("%s%s?%s=%s", host, addonsLandingPath, addonsClientIDParam, appID)
}

// addonsHintURL returns the scan URL, degrading to the bare landing page on encode error.
func addonsHintURL(brand core.LarkBrand, appID string, a ManifestAddons) string {
	url, err := consoleAddonsURL(brand, appID, a)
	if err != nil {
		return consoleLandingURL(brand, appID)
	}
	return url
}

// missingScopeAddons routes missing scopes into the identity-appropriate section.
func missingScopeAddons(identity core.Identity, missing []string) ManifestAddons {
	s := &AddonsScopes{}
	if identity.IsBot() {
		s.Tenant = missing
	} else {
		s.User = missing
	}
	return ManifestAddons{Scopes: s}
}

// missingSubscriptionAddons routes missing events/callbacks into the right section.
func missingSubscriptionAddons(subType eventlib.SubscriptionType, identity core.Identity, missing []string) ManifestAddons {
	if subType == eventlib.SubTypeCallback {
		return ManifestAddons{Callbacks: &AddonsCallbacks{Items: missing}}
	}
	ev := &AddonsEvents{}
	if identity.IsBot() {
		ev.Items.Tenant = missing
	} else {
		ev.Items.User = missing
	}
	return ManifestAddons{Events: ev}
}
