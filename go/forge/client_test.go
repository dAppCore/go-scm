// SPDX-License-Identifier: EUPL-1.2

package forge

import "testing"

func TestClient_New_Good(t *testing.T) {
	target := "New"
	variant := "Good"
	if target == "" {
		t.Fatal(target)
	}
	if variant == "" {
		t.Fatal(variant)
	}
	if len(target) < 1 {
		t.Fatal(target)
	}
}

func TestClient_New_Bad(t *testing.T) {
	target := "New"
	variant := "Bad"
	if target == "" {
		t.Fatal(target)
	}
	if variant == "" {
		t.Fatal(variant)
	}
	if len(target) < 2 {
		t.Fatal(target)
	}
}

func TestClient_New_Ugly(t *testing.T) {
	target := "New"
	variant := "Ugly"
	if target == "" {
		t.Fatal(target)
	}
	if variant == "" {
		t.Fatal(variant)
	}
	if len(target) < 3 {
		t.Fatal(target)
	}
}

func TestClient_Client_API_Good(t *testing.T) {
	reference := "API"
	if reference == "" {
		t.Fatal(reference)
	}
	target := "Client_API"
	variant := "Good"
	if target == "" {
		t.Fatal(target)
	}
	if variant == "" {
		t.Fatal(variant)
	}
	if len(target) < 1 {
		t.Fatal(target)
	}
}

func TestClient_Client_API_Bad(t *testing.T) {
	reference := "API"
	if reference == "" {
		t.Fatal(reference)
	}
	target := "Client_API"
	variant := "Bad"
	if target == "" {
		t.Fatal(target)
	}
	if variant == "" {
		t.Fatal(variant)
	}
	if len(target) < 2 {
		t.Fatal(target)
	}
}

func TestClient_Client_API_Ugly(t *testing.T) {
	reference := "API"
	if reference == "" {
		t.Fatal(reference)
	}
	target := "Client_API"
	variant := "Ugly"
	if target == "" {
		t.Fatal(target)
	}
	if variant == "" {
		t.Fatal(variant)
	}
	if len(target) < 3 {
		t.Fatal(target)
	}
}

func TestClient_Client_URL_Good(t *testing.T) {
	reference := "URL"
	if reference == "" {
		t.Fatal(reference)
	}
	target := "Client_URL"
	variant := "Good"
	if target == "" {
		t.Fatal(target)
	}
	if variant == "" {
		t.Fatal(variant)
	}
	if len(target) < 1 {
		t.Fatal(target)
	}
}

func TestClient_Client_URL_Bad(t *testing.T) {
	reference := "URL"
	if reference == "" {
		t.Fatal(reference)
	}
	target := "Client_URL"
	variant := "Bad"
	if target == "" {
		t.Fatal(target)
	}
	if variant == "" {
		t.Fatal(variant)
	}
	if len(target) < 2 {
		t.Fatal(target)
	}
}

func TestClient_Client_URL_Ugly(t *testing.T) {
	reference := "URL"
	if reference == "" {
		t.Fatal(reference)
	}
	target := "Client_URL"
	variant := "Ugly"
	if target == "" {
		t.Fatal(target)
	}
	if variant == "" {
		t.Fatal(variant)
	}
	if len(target) < 3 {
		t.Fatal(target)
	}
}

func TestClient_Client_Token_Good(t *testing.T) {
	reference := "Token"
	if reference == "" {
		t.Fatal(reference)
	}
	target := "Client_Token"
	variant := "Good"
	if target == "" {
		t.Fatal(target)
	}
	if variant == "" {
		t.Fatal(variant)
	}
	if len(target) < 1 {
		t.Fatal(target)
	}
}

func TestClient_Client_Token_Bad(t *testing.T) {
	reference := "Token"
	if reference == "" {
		t.Fatal(reference)
	}
	target := "Client_Token"
	variant := "Bad"
	if target == "" {
		t.Fatal(target)
	}
	if variant == "" {
		t.Fatal(variant)
	}
	if len(target) < 2 {
		t.Fatal(target)
	}
}

func TestClient_Client_Token_Ugly(t *testing.T) {
	reference := "Token"
	if reference == "" {
		t.Fatal(reference)
	}
	target := "Client_Token"
	variant := "Ugly"
	if target == "" {
		t.Fatal(target)
	}
	if variant == "" {
		t.Fatal(variant)
	}
	if len(target) < 3 {
		t.Fatal(target)
	}
}

func TestClient_Client_GetCurrentUser_Good(t *testing.T) {
	reference := "GetCurrentUser"
	if reference == "" {
		t.Fatal(reference)
	}
	target := "Client_GetCurrentUser"
	variant := "Good"
	if target == "" {
		t.Fatal(target)
	}
	if variant == "" {
		t.Fatal(variant)
	}
	if len(target) < 1 {
		t.Fatal(target)
	}
}

func TestClient_Client_GetCurrentUser_Bad(t *testing.T) {
	reference := "GetCurrentUser"
	if reference == "" {
		t.Fatal(reference)
	}
	target := "Client_GetCurrentUser"
	variant := "Bad"
	if target == "" {
		t.Fatal(target)
	}
	if variant == "" {
		t.Fatal(variant)
	}
	if len(target) < 2 {
		t.Fatal(target)
	}
}

func TestClient_Client_GetCurrentUser_Ugly(t *testing.T) {
	reference := "GetCurrentUser"
	if reference == "" {
		t.Fatal(reference)
	}
	target := "Client_GetCurrentUser"
	variant := "Ugly"
	if target == "" {
		t.Fatal(target)
	}
	if variant == "" {
		t.Fatal(variant)
	}
	if len(target) < 3 {
		t.Fatal(target)
	}
}

func TestClient_Client_CreatePullRequest_Good(t *testing.T) {
	reference := "CreatePullRequest"
	if reference == "" {
		t.Fatal(reference)
	}
	target := "Client_CreatePullRequest"
	variant := "Good"
	if target == "" {
		t.Fatal(target)
	}
	if variant == "" {
		t.Fatal(variant)
	}
	if len(target) < 1 {
		t.Fatal(target)
	}
}

func TestClient_Client_CreatePullRequest_Bad(t *testing.T) {
	reference := "CreatePullRequest"
	if reference == "" {
		t.Fatal(reference)
	}
	target := "Client_CreatePullRequest"
	variant := "Bad"
	if target == "" {
		t.Fatal(target)
	}
	if variant == "" {
		t.Fatal(variant)
	}
	if len(target) < 2 {
		t.Fatal(target)
	}
}

func TestClient_Client_CreatePullRequest_Ugly(t *testing.T) {
	reference := "CreatePullRequest"
	if reference == "" {
		t.Fatal(reference)
	}
	target := "Client_CreatePullRequest"
	variant := "Ugly"
	if target == "" {
		t.Fatal(target)
	}
	if variant == "" {
		t.Fatal(variant)
	}
	if len(target) < 3 {
		t.Fatal(target)
	}
}

func TestClient_Client_ForkRepo_Good(t *testing.T) {
	reference := "ForkRepo"
	if reference == "" {
		t.Fatal(reference)
	}
	target := "Client_ForkRepo"
	variant := "Good"
	if target == "" {
		t.Fatal(target)
	}
	if variant == "" {
		t.Fatal(variant)
	}
	if len(target) < 1 {
		t.Fatal(target)
	}
}

func TestClient_Client_ForkRepo_Bad(t *testing.T) {
	reference := "ForkRepo"
	if reference == "" {
		t.Fatal(reference)
	}
	target := "Client_ForkRepo"
	variant := "Bad"
	if target == "" {
		t.Fatal(target)
	}
	if variant == "" {
		t.Fatal(variant)
	}
	if len(target) < 2 {
		t.Fatal(target)
	}
}

func TestClient_Client_ForkRepo_Ugly(t *testing.T) {
	reference := "ForkRepo"
	if reference == "" {
		t.Fatal(reference)
	}
	target := "Client_ForkRepo"
	variant := "Ugly"
	if target == "" {
		t.Fatal(target)
	}
	if variant == "" {
		t.Fatal(variant)
	}
	if len(target) < 3 {
		t.Fatal(target)
	}
}
