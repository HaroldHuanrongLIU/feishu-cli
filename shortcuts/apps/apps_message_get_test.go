// Copyright (c) 2026 Lark Technologies Pte. Ltd.
// SPDX-License-Identifier: MIT

package apps

import (
	"strings"
	"testing"

	"github.com/larksuite/cli/errs"
	"github.com/larksuite/cli/internal/httpmock"
)

const msgGetURL = "/open-apis/spark/v1/apps/app_x/sessions/sess_1/turns/turn_9/reply_message"

func TestAppsMessageGet_Success(t *testing.T) {
	factory, stdout, reg := newAppsExecuteFactory(t)
	reg.Register(&httpmock.Stub{
		Method: "GET",
		URL:    msgGetURL,
		Body: map[string]interface{}{
			"code": 0,
			"data": map[string]interface{}{
				"messages": []interface{}{
					map[string]interface{}{"message_id": "m1", "role": "assistant", "content": "hi"},
				},
				"next_cursor": 12,
				"has_more":    true,
			},
		},
	})
	if err := runAppsShortcut(t, AppsMessageGet,
		[]string{"+message-get", "--app-id", "app_x", "--session-id", "sess_1", "--turn-id", "turn_9", "--as", "user"},
		factory, stdout); err != nil {
		t.Fatalf("execute err=%v", err)
	}
	got := stdout.String()
	if !strings.Contains(got, `"message_id": "m1"`) || !strings.Contains(got, `"next_cursor": 12`) {
		t.Fatalf("stdout missing messages/next_cursor: %s", got)
	}
}

func TestAppsMessageGet_CursorOnlyWhenSet(t *testing.T) {
	factory, stdout, _ := newAppsExecuteFactory(t)
	if err := runAppsShortcut(t, AppsMessageGet,
		[]string{"+message-get", "--app-id", "app_x", "--session-id", "sess_1", "--turn-id", "turn_9", "--dry-run", "--as", "user"},
		factory, stdout); err != nil {
		t.Fatalf("dry-run err=%v", err)
	}
	if strings.Contains(stdout.String(), "cursor") {
		t.Fatalf("cursor must be absent when --cursor not set: %s", stdout.String())
	}

	factory2, stdout2, _ := newAppsExecuteFactory(t)
	if err := runAppsShortcut(t, AppsMessageGet,
		[]string{"+message-get", "--app-id", "app_x", "--session-id", "sess_1", "--turn-id", "turn_9", "--cursor", "5", "--dry-run", "--as", "user"},
		factory2, stdout2); err != nil {
		t.Fatalf("dry-run err=%v", err)
	}
	got := stdout2.String()
	if !strings.Contains(got, "cursor") || !strings.Contains(got, "5") {
		t.Fatalf("dry-run missing cursor=5: %s", got)
	}
}

func TestAppsMessageGet_RequiresIDs(t *testing.T) {
	cases := []struct {
		name string
		args []string
		want string
	}{
		{"no app-id", []string{"+message-get", "--app-id", "", "--session-id", "s", "--turn-id", "t", "--as", "user"}, "app-id"},
		{"no session-id", []string{"+message-get", "--app-id", "a", "--session-id", "", "--turn-id", "t", "--as", "user"}, "session-id"},
		{"no turn-id", []string{"+message-get", "--app-id", "a", "--session-id", "s", "--turn-id", "", "--as", "user"}, "turn-id"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			factory, stdout, _ := newAppsExecuteFactory(t)
			err := runAppsShortcut(t, AppsMessageGet, c.args, factory, stdout)
			if err == nil || !strings.Contains(err.Error(), c.want) {
				t.Fatalf("expected %q required error, got %v", c.want, err)
			}
		})
	}
}

func TestAppsMessageGet_APIErrorSurfacesHint(t *testing.T) {
	factory, stdout, reg := newAppsExecuteFactory(t)
	reg.Register(&httpmock.Stub{
		Method: "GET",
		URL:    msgGetURL,
		Body:   map[string]interface{}{"code": 1254043, "msg": "permission denied"},
	})
	err := runAppsShortcut(t, AppsMessageGet,
		[]string{"+message-get", "--app-id", "app_x", "--session-id", "sess_1", "--turn-id", "turn_9", "--as", "user"},
		factory, stdout)
	if err == nil {
		t.Fatalf("expected API error, got nil")
	}
	p, ok := errs.ProblemOf(err)
	if !ok {
		t.Fatalf("expected a typed Problem error, got %T: %v", err, err)
	}
	if !strings.Contains(p.Hint, "+session-get") {
		t.Fatalf("error should carry domain hint pointing at +session-get, got hint=%q (err=%v)", p.Hint, err)
	}
}
