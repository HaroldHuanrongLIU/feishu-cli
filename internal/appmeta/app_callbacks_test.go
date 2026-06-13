// Copyright (c) 2026 Lark Technologies Pte. Ltd.
// SPDX-License-Identifier: MIT

package appmeta

import (
	"context"
	"encoding/json"
	"testing"
)

type fakeCallbackClient struct {
	raw string
	err error
}

func (f fakeCallbackClient) CallAPI(_ context.Context, _, _ string, _ interface{}) (json.RawMessage, error) {
	if f.err != nil {
		return nil, f.err
	}
	return json.RawMessage(f.raw), nil
}

func TestFetchSubscribedCallbacks_ParsesList(t *testing.T) {
	raw := `{"code":0,"data":{"app":{"callback_info":{"callback_type":"websocket","subscribed_callbacks":["card.action.trigger","profile.view.get"]}}},"msg":"success"}`
	got, err := FetchSubscribedCallbacks(context.Background(), fakeCallbackClient{raw: raw}, "cli_x")
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	want := []string{"card.action.trigger", "profile.view.get"}
	if len(got) != len(want) {
		t.Fatalf("got %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("got[%d] = %q, want %q", i, got[i], want[i])
		}
	}
}

func TestFetchSubscribedCallbacks_NoCallbackInfo(t *testing.T) {
	raw := `{"code":0,"data":{"app":{"app_id":"cli_x"}},"msg":"success"}`
	got, err := FetchSubscribedCallbacks(context.Background(), fakeCallbackClient{raw: raw}, "cli_x")
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if got != nil {
		t.Errorf("got %v, want nil", got)
	}
}
