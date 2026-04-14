package plugin

import (
	"testing"
)

func TestValidateMinimalManifest(t *testing.T) {
	m := &Manifest{
		ID:      "com.example.test",
		Name:    "Test Plugin",
		Version: "1.0.0",
		Type:    TypeBackgroundService,
	}
	if err := m.Validate(); err != nil {
		t.Fatalf("expected valid manifest, got: %v", err)
	}
}

func TestValidateVisualTileRequiresTileDef(t *testing.T) {
	m := &Manifest{
		ID:      "com.example.tile",
		Name:    "Tile Plugin",
		Version: "1.0.0",
		Type:    TypeVisualTile,
	}
	if err := m.Validate(); err == nil {
		t.Fatal("expected error for visual_tile without tile config")
	}

	m.Tile = &TileDef{SupportsFullscreen: true}
	if err := m.Validate(); err != nil {
		t.Fatalf("expected valid with tile config, got: %v", err)
	}
}

func TestValidateMissingID(t *testing.T) {
	m := &Manifest{Name: "Test", Version: "1.0.0", Type: TypeContextMenu}
	if err := m.Validate(); err == nil {
		t.Fatal("expected error for missing id")
	}
}

func TestValidateBadIDFormat(t *testing.T) {
	tests := []string{
		"UPPERCASE",
		"has spaces",
		"_leading",
		"",
	}
	for _, id := range tests {
		m := &Manifest{ID: id, Name: "Test", Version: "1.0.0", Type: TypeContextMenu}
		if err := m.Validate(); err == nil {
			t.Errorf("expected error for id %q", id)
		}
	}
}

func TestValidateGoodIDFormats(t *testing.T) {
	tests := []string{
		"com.example.test",
		"org.robcord.steam",
		"dev.plugin",
		"a.b.c",
	}
	for _, id := range tests {
		m := &Manifest{ID: id, Name: "Test", Version: "1.0.0", Type: TypeBackgroundService}
		if err := m.Validate(); err != nil {
			t.Errorf("expected valid for id %q, got: %v", id, err)
		}
	}
}

func TestValidateBadVersion(t *testing.T) {
	m := &Manifest{ID: "com.test.a", Name: "Test", Version: "notaversion", Type: TypeContextMenu}
	if err := m.Validate(); err == nil {
		t.Fatal("expected error for bad version")
	}
}

func TestValidateUnknownType(t *testing.T) {
	m := &Manifest{ID: "com.test.a", Name: "Test", Version: "1.0.0", Type: "invalid_type"}
	if err := m.Validate(); err == nil {
		t.Fatal("expected error for unknown type")
	}
}

func TestValidateUnknownPermission(t *testing.T) {
	m := &Manifest{
		ID:          "com.test.a",
		Name:        "Test",
		Version:     "1.0.0",
		Type:        TypeBackgroundService,
		Permissions: []string{"read:members", "hack:system"},
	}
	if err := m.Validate(); err == nil {
		t.Fatal("expected error for unknown permission")
	}
}

func TestValidateSettings(t *testing.T) {
	m := &Manifest{
		ID:      "com.test.a",
		Name:    "Test",
		Version: "1.0.0",
		Type:    TypeBackgroundService,
		Settings: []SettingDef{
			{Key: "api_key", Type: "string", Label: "API Key"},
			{Key: "max_items", Type: "number", Label: "Max Items", Default: 10},
		},
	}
	if err := m.Validate(); err != nil {
		t.Fatalf("expected valid settings, got: %v", err)
	}
}

func TestValidateSettingsMissingKey(t *testing.T) {
	m := &Manifest{
		ID:      "com.test.a",
		Name:    "Test",
		Version: "1.0.0",
		Type:    TypeBackgroundService,
		Settings: []SettingDef{
			{Key: "", Type: "string", Label: "Missing Key"},
		},
	}
	if err := m.Validate(); err == nil {
		t.Fatal("expected error for setting without key")
	}
}

func TestValidateSelectSettingRequiresOptions(t *testing.T) {
	m := &Manifest{
		ID:      "com.test.a",
		Name:    "Test",
		Version: "1.0.0",
		Type:    TypeBackgroundService,
		Settings: []SettingDef{
			{Key: "mode", Type: "select", Label: "Mode"},
		},
	}
	if err := m.Validate(); err == nil {
		t.Fatal("expected error for select type without options")
	}
}

func TestValidateContextMenuEntry(t *testing.T) {
	m := &Manifest{
		ID:      "com.test.steam",
		Name:    "Steam",
		Version: "1.0.0",
		Type:    TypeContextMenu,
		ContextMenuEntries: []ContextMenuEntry{
			{
				ID:     "steam-profile",
				Target: "member",
				Label:  "Steam-Profil",
				Action: ContextMenuAction{Type: "open_url", URLTemplate: "https://steam/{{id}}"},
			},
		},
	}
	if err := m.Validate(); err != nil {
		t.Fatalf("expected valid context menu, got: %v", err)
	}
}

func TestValidateContextMenuBadTarget(t *testing.T) {
	m := &Manifest{
		ID:      "com.test.a",
		Name:    "Test",
		Version: "1.0.0",
		Type:    TypeContextMenu,
		ContextMenuEntries: []ContextMenuEntry{
			{ID: "x", Target: "invalid", Label: "X", Action: ContextMenuAction{Type: "open_url"}},
		},
	}
	if err := m.Validate(); err == nil {
		t.Fatal("expected error for bad context menu target")
	}
}

func TestParseManifest(t *testing.T) {
	raw := `{
		"id": "com.example.test",
		"name": "Test",
		"version": "1.0.0",
		"type": "background_service",
		"permissions": ["read:members"],
		"hooks": ["on:member_joined"]
	}`
	m, err := ParseManifest([]byte(raw))
	if err != nil {
		t.Fatalf("expected valid parse, got: %v", err)
	}
	if m.ID != "com.example.test" {
		t.Errorf("expected id com.example.test, got %s", m.ID)
	}
	if m.SchemaVersion != SchemaVersion {
		t.Errorf("expected schema_version %d, got %d", SchemaVersion, m.SchemaVersion)
	}
	if len(m.Permissions) != 1 || m.Permissions[0] != "read:members" {
		t.Errorf("unexpected permissions: %v", m.Permissions)
	}
}

func TestParseManifestInvalidJSON(t *testing.T) {
	_, err := ParseManifest([]byte("not json"))
	if err == nil {
		t.Fatal("expected error for invalid JSON")
	}
}

func TestNameTooLong(t *testing.T) {
	long := ""
	for i := 0; i < 65; i++ {
		long += "a"
	}
	m := &Manifest{ID: "com.test.a", Name: long, Version: "1.0.0", Type: TypeContextMenu}
	if err := m.Validate(); err == nil {
		t.Fatal("expected error for name > 64 chars")
	}
}

func TestValidateNetworkHostsFromSettings(t *testing.T) {
	m := &Manifest{
		ID: "com.test.a", Name: "Test", Version: "1.0.0", Type: TypeBackgroundService,
		Settings: []SettingDef{
			{Key: "server_url", Type: "string", Label: "Server URL"},
		},
		NetworkHostsFromSettings: []string{"server_url"},
	}
	if err := m.Validate(); err != nil {
		t.Fatalf("expected valid, got: %v", err)
	}
}

func TestValidateNetworkHostsFromSettings_BadRef(t *testing.T) {
	m := &Manifest{
		ID: "com.test.a", Name: "Test", Version: "1.0.0", Type: TypeBackgroundService,
		NetworkHostsFromSettings: []string{"nonexistent_key"},
	}
	if err := m.Validate(); err == nil {
		t.Fatal("expected error for network_hosts_from_settings referencing undefined setting")
	}
}

func TestValidateCustomPermissions(t *testing.T) {
	m := &Manifest{
		ID: "com.test.a", Name: "Test", Version: "1.0.0", Type: TypeBackgroundService,
		CustomPermissions: []CustomPermissionDef{
			{Key: "media:browse", Label: "Browse Media", Domain: "media"},
			{Key: "media:start_session", Label: "Start Session", Domain: "media"},
		},
	}
	if err := m.Validate(); err != nil {
		t.Fatalf("expected valid custom permissions, got: %v", err)
	}
}

func TestValidateCustomPermissions_BadKey(t *testing.T) {
	tests := []CustomPermissionDef{
		{Key: "", Label: "X", Domain: "x"},
		{Key: "no-domain", Label: "X", Domain: "x"},
		{Key: "UPPER:case", Label: "X", Domain: "x"},
	}
	for _, cp := range tests {
		m := &Manifest{
			ID: "com.test.a", Name: "Test", Version: "1.0.0", Type: TypeBackgroundService,
			CustomPermissions: []CustomPermissionDef{cp},
		}
		if err := m.Validate(); err == nil {
			t.Errorf("expected error for custom permission key %q", cp.Key)
		}
	}
}

func TestValidateRoutes(t *testing.T) {
	m := &Manifest{
		ID: "com.test.a", Name: "Test", Version: "1.0.0", Type: TypeBackgroundService,
		CustomPermissions: []CustomPermissionDef{
			{Key: "media:browse", Label: "Browse", Domain: "media"},
		},
		Routes: []RouteDef{
			{Path: "libraries", Method: "GET", Handler: "onGetLibraries", Permission: "media:browse"},
			{Path: "items/{id}", Method: "GET", Handler: "onGetItem"},
		},
	}
	if err := m.Validate(); err != nil {
		t.Fatalf("expected valid routes, got: %v", err)
	}
}

func TestValidateRoutes_BadMethod(t *testing.T) {
	m := &Manifest{
		ID: "com.test.a", Name: "Test", Version: "1.0.0", Type: TypeBackgroundService,
		Routes: []RouteDef{
			{Path: "test", Method: "PATCH", Handler: "onTest"},
		},
	}
	if err := m.Validate(); err == nil {
		t.Fatal("expected error for invalid route method PATCH")
	}
}

func TestValidateRoutes_UndefinedPermission(t *testing.T) {
	m := &Manifest{
		ID: "com.test.a", Name: "Test", Version: "1.0.0", Type: TypeBackgroundService,
		Routes: []RouteDef{
			{Path: "test", Method: "GET", Handler: "onTest", Permission: "undefined:perm"},
		},
	}
	if err := m.Validate(); err == nil {
		t.Fatal("expected error for route referencing undefined custom permission")
	}
}

func TestParseManifestWithNewFields(t *testing.T) {
	raw := `{
		"id": "com.robcord.jellyfin",
		"name": "Jellyfin",
		"version": "1.0.0",
		"type": "visual_tile",
		"permissions": ["voice:tile", "network:http", "storage:secure", "routes:http"],
		"tile": {"nameplate_icon": "film", "supports_fullscreen": true},
		"settings": [
			{"key": "jellyfin_url", "type": "string", "label": "Server URL"}
		],
		"network_hosts_from_settings": ["jellyfin_url"],
		"custom_permissions": [
			{"key": "media:browse", "label": "Browse Media", "domain": "media"}
		],
		"routes": [
			{"path": "libraries", "method": "GET", "handler": "onGetLibraries", "permission": "media:browse"}
		]
	}`
	m, err := ParseManifest([]byte(raw))
	if err != nil {
		t.Fatalf("expected valid parse, got: %v", err)
	}
	if len(m.NetworkHostsFromSettings) != 1 || m.NetworkHostsFromSettings[0] != "jellyfin_url" {
		t.Errorf("unexpected network_hosts_from_settings: %v", m.NetworkHostsFromSettings)
	}
	if len(m.CustomPermissions) != 1 || m.CustomPermissions[0].Key != "media:browse" {
		t.Errorf("unexpected custom_permissions: %v", m.CustomPermissions)
	}
	if len(m.Routes) != 1 || m.Routes[0].Handler != "onGetLibraries" {
		t.Errorf("unexpected routes: %v", m.Routes)
	}
}
