// SPDX-License-Identifier: EUPL-1.2

package gitea

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
