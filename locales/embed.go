// SPDX-License-Identifier: EUPL-1.2

// Package locales embeds translation files for this module.
package locales

import "embed"

//
//go:embed *.json
var FS embed.FS
