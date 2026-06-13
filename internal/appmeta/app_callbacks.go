// Copyright (c) 2026 Lark Technologies Pte. Ltd.
// SPDX-License-Identifier: MIT

package appmeta

import (
	"context"
	"encoding/json"
	"fmt"
)

// FetchSubscribedCallbacks returns the app's currently subscribed callback names
// from application/get. Returns (nil, nil) when callback_info is absent.
// Identity must be bot: the endpoint is app-level.
func FetchSubscribedCallbacks(ctx context.Context, client APIClient, appID string) ([]string, error) {
	path := fmt.Sprintf("/open-apis/application/v6/applications/%s?lang=zh_cn", appID)
	raw, err := client.CallAPI(ctx, "GET", path, nil)
	if err != nil {
		return nil, err
	}

	var envelope struct {
		Data struct {
			App struct {
				CallbackInfo *struct {
					SubscribedCallbacks []string `json:"subscribed_callbacks"`
				} `json:"callback_info"`
			} `json:"app"`
		} `json:"data"`
	}
	if err := json.Unmarshal(raw, &envelope); err != nil {
		return nil, fmt.Errorf("decode application response: %w", err)
	}
	if envelope.Data.App.CallbackInfo == nil {
		return nil, nil
	}
	return envelope.Data.App.CallbackInfo.SubscribedCallbacks, nil
}
