// Copyright (c) 2026 Lark Technologies Pte. Ltd.
// SPDX-License-Identifier: MIT

package apps

import (
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/larksuite/cli/internal/output"
	"github.com/larksuite/cli/internal/validate"
	"github.com/larksuite/cli/shortcuts/common"
)

const msgGetHint = "verify --app-id / --session-id / --turn-id are correct; get the latest turn_id via `lark-cli apps +session-get --app-id <app_id> --session-id <session_id>`"

var AppsMessageGet = common.Shortcut{
	Service:     appsService,
	Command:     "+message-get",
	Description: "Get the reply messages of a session turn (cursor pagination)",
	Risk:        "read",
	Tips: []string{
		"Example: lark-cli apps +message-get --app-id <app_id> --session-id <session_id> --turn-id <turn_id>",
		"Tip: turn_id comes from `+session-get` latest_turn.turn_id; page with --cursor <next_cursor>",
	},
	Scopes:    []string{"spark:app:read"},
	AuthTypes: []string{"user"},
	HasFormat: true,
	Flags: []common.Flag{
		// app-id / session-id / turn-id are intentionally NOT Required:true. The
		// framework maps Required:true to cobra's MarkFlagRequired, whose error is
		// plain-text exit-1, bypassing the structured envelope. Spec §6 mandates
		// exit-2 + a {"ok":false,"error":{...}} validation envelope, so the
		// emptiness checks live in Validate (output.ErrValidation -> exit 2).
		{Name: "app-id", Desc: "app ID"},
		{Name: "session-id", Desc: "session ID"},
		{Name: "turn-id", Desc: "turn ID (from +session-get latest_turn.turn_id)"},
		{Name: "cursor", Type: "int", Desc: "pagination cursor from previous response next_cursor (omit for first page)"},
	},
	Validate: func(ctx context.Context, rctx *common.RuntimeContext) error {
		if strings.TrimSpace(rctx.Str("app-id")) == "" {
			return output.ErrValidation("--app-id is required")
		}
		if strings.TrimSpace(rctx.Str("session-id")) == "" {
			return output.ErrValidation("--session-id is required")
		}
		if strings.TrimSpace(rctx.Str("turn-id")) == "" {
			return output.ErrValidation("--turn-id is required")
		}
		return nil
	},
	DryRun: func(ctx context.Context, rctx *common.RuntimeContext) *common.DryRunAPI {
		return common.NewDryRunAPI().
			GET(replyMessagePath(rctx.Str("app-id"), rctx.Str("session-id"), rctx.Str("turn-id"))).
			Desc("Get the reply messages of a session turn").
			Params(buildMessageGetParams(rctx))
	},
	Execute: func(ctx context.Context, rctx *common.RuntimeContext) error {
		data, err := rctx.CallAPITyped("GET",
			replyMessagePath(rctx.Str("app-id"), rctx.Str("session-id"), rctx.Str("turn-id")),
			buildMessageGetParams(rctx), nil)
		if err != nil {
			return withAppsHint(err, msgGetHint)
		}
		messages, _ := data["messages"].([]interface{})
		rctx.OutFormat(data, nil, func(w io.Writer) {
			rows := make([]map[string]interface{}, 0, len(messages))
			for _, item := range messages {
				m, ok := item.(map[string]interface{})
				if !ok {
					continue
				}
				rows = append(rows, map[string]interface{}{
					"message_id": m["message_id"],
					"role":       m["role"],
					"content":    m["content"],
				})
			}
			output.PrintTable(w, rows)
			fmt.Fprintf(w, "next_cursor: %v  has_more: %v\n", data["next_cursor"], data["has_more"])
		})
		return nil
	},
}

func replyMessagePath(appID, sessionID, turnID string) string {
	return fmt.Sprintf("%s/apps/%s/sessions/%s/turns/%s/reply_message",
		apiBasePath,
		validate.EncodePathSegment(strings.TrimSpace(appID)),
		validate.EncodePathSegment(strings.TrimSpace(sessionID)),
		validate.EncodePathSegment(strings.TrimSpace(turnID)))
}

func buildMessageGetParams(rctx *common.RuntimeContext) map[string]interface{} {
	params := map[string]interface{}{}
	if rctx.Cmd.Flags().Changed("cursor") {
		params["cursor"] = rctx.Int("cursor")
	}
	return params
}
