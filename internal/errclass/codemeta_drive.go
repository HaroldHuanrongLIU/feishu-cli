// Copyright (c) 2026 Lark Technologies Pte. Ltd.
// SPDX-License-Identifier: MIT

package errclass

import "github.com/larksuite/cli/errs"

// driveCodeMeta holds drive/docs-service Lark code → CodeMeta mappings.
// Only codes whose meaning is verifiable from repo evidence are registered;
// ambiguous codes fall back to CategoryAPI via BuildAPIError.
// BuildAPIError consumes this map via mergeCodeMeta + LookupCodeMeta.
var driveCodeMeta = map[int]CodeMeta{
	1061044: {Category: errs.CategoryAPI, Subtype: errs.SubtypeNotFound},          // parent folder does not exist (upload)
	1069302: {Category: errs.CategoryAPI, Subtype: errs.SubtypeInvalidParameters}, // comment endpoint "Invalid or missing parameters"

	// Commercial plan codes returned by the cloud-space explorer service
	// and passed through verbatim by document backends (observed via
	// slides_ai create — engine log "A2 create quota exceeded, code=90003087"
	// — and drive/v1/import_tasks, both as HTTP 200 with body code≠0).
	// Plan/billing limits: retrying can never succeed. 90003086/90003087 are
	// plan creation-count quotas (upgrade the plan or free quota); 90003088 is
	// not a quota at all — the tenant has not purchased/enabled the docs
	// module, so it maps to FailedPrecondition (admin must enable the module)
	// rather than QuotaExceeded, whose default hint suggests freeing quota.
	// Hint is set per-code because the SubtypeQuotaExceeded default
	// ("retry after the relevant quota resets") misleads here: plan quotas
	// never reset on their own. Keep the wording in sync with the legacy
	// path's legacyHints entries (internal/output/lark_errors.go).
	90003086: {Category: errs.CategoryAPI, Subtype: errs.SubtypeQuotaExceeded,
		Hint: "document creation quota of the current plan reached: upgrade the plan or delete documents you no longer need to free quota; retrying will not help"}, // premium plan creation count limit reached
	90003087: {Category: errs.CategoryAPI, Subtype: errs.SubtypeQuotaExceeded,
		Hint: "document creation quota of the current plan reached: upgrade the plan or delete documents you no longer need to free quota; retrying will not help"}, // A2 plan creation count limit reached
	90003088: {Category: errs.CategoryAPI, Subtype: errs.SubtypeFailedPrecondition,
		Hint: "the tenant has not purchased or enabled the docs module; ask the tenant admin to enable it before creating documents"}, // unbundle: tenant has not purchased / been granted the docs module
}

func init() { mergeCodeMeta(driveCodeMeta, "drive") }
