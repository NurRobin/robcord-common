package plugin

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
)

// SchemaVersion is the current manifest schema version.
const SchemaVersion = 1

// PluginType defines the primary archetype of a plugin.
type PluginType string

const (
	TypeVisualTile       PluginType = "visual_tile"
	TypeChatExtension    PluginType = "chat_extension"
	TypeBackgroundService PluginType = "background_service"
	TypeContextMenu      PluginType = "context_menu"
)

var validTypes = map[PluginType]bool{
	TypeVisualTile:        true,
	TypeChatExtension:     true,
	TypeBackgroundService: true,
	TypeContextMenu:       true,
}

// Manifest describes a plugin's identity, capabilities, and configuration.
// Parsed from plugin.json inside a .rcplugin archive.
type Manifest struct {
	// Identity
	ID      string     `json:"id"`
	Name    string     `json:"name"`
	Version string     `json:"version"`
	Author  string     `json:"author,omitempty"`
	Type    PluginType `json:"type"`

	// Schema version (default 1 if omitted)
	SchemaVersion int `json:"schema_version,omitempty"`

	// Capabilities — what the plugin needs access to
	Permissions []string `json:"permissions,omitempty"`

	// Event hooks the plugin subscribes to (background services)
	Hooks []string `json:"hooks,omitempty"`

	// Admin-configurable settings (rendered as a form in the UI)
	Settings []SettingDef `json:"settings,omitempty"`

	// Context menu entries (for context_menu and mixed-type plugins)
	ContextMenuEntries []ContextMenuEntry `json:"context_menu_entries,omitempty"`

	// Slash commands (for chat_extension plugins)
	Commands []CommandDef `json:"commands,omitempty"`

	// Voice tile config (for visual_tile plugins)
	Tile *TileDef `json:"tile,omitempty"`

	// Webhook URLs for external backend (alternative to WASM)
	WebhookURL string `json:"webhook_url,omitempty"`
}

// SettingDef describes a single admin-configurable setting.
type SettingDef struct {
	Key         string `json:"key"`
	Type        string `json:"type"` // "string", "number", "boolean", "select"
	Label       string `json:"label"`
	Description string `json:"description,omitempty"`
	Default     any    `json:"default,omitempty"`
	Required    bool   `json:"required,omitempty"`
	// Options for "select" type
	Options []SettingOption `json:"options,omitempty"`
}

// SettingOption is a choice for select-type settings.
type SettingOption struct {
	Value string `json:"value"`
	Label string `json:"label"`
}

// ContextMenuEntry registers an action in the right-click menu.
type ContextMenuEntry struct {
	ID        string              `json:"id"`
	Target    string              `json:"target"` // "member", "message", "channel"
	Label     string              `json:"label"`
	Icon      string              `json:"icon,omitempty"`
	Condition string              `json:"condition,omitempty"`
	Action    ContextMenuAction   `json:"action"`
}

// ContextMenuAction defines what happens when a context menu entry is clicked.
type ContextMenuAction struct {
	Type        string `json:"type"` // "open_url", "webhook"
	URLTemplate string `json:"url_template,omitempty"`
	WebhookURL  string `json:"webhook_url,omitempty"`
}

// CommandDef registers a slash command.
type CommandDef struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	// Parameters could be added later for autocomplete
}

// TileDef configures a visual_tile plugin's voice grid behavior.
type TileDef struct {
	NameplateIcon    string `json:"nameplate_icon,omitempty"`
	DefaultState     string `json:"default_state,omitempty"` // "grid" (default)
	SupportsFullscreen bool `json:"supports_fullscreen,omitempty"`
	MaxParticipants  int    `json:"max_participants,omitempty"`
	MinParticipants  int    `json:"min_participants,omitempty"`
}

// pluginIDRe validates plugin IDs: reverse-domain style (e.g. com.example.my-plugin).
var pluginIDRe = regexp.MustCompile(`^[a-z][a-z0-9]*(\.[a-z][a-z0-9]*)*(-[a-z0-9]+)*$`)

// semverRe validates a basic semver string (major.minor.patch).
var semverRe = regexp.MustCompile(`^\d+\.\d+\.\d+`)

// validPermissions is the set of recognized plugin API scopes.
var validPermissions = map[string]bool{
	"read:members":          true,
	"read:channels":         true,
	"read:messages":         true,
	"write:messages":        true,
	"voice:tile":            true,
	"voice:overlay":         true,
	"storage:plugin":        true,
	"storage:custom_fields": true,
	"webhook:receive":       true,
	"messaging:participants": true,
	"member:kick":           true,
}

var validSettingTypes = map[string]bool{
	"string":  true,
	"number":  true,
	"boolean": true,
	"select":  true,
}

var validContextMenuTargets = map[string]bool{
	"member":  true,
	"message": true,
	"channel": true,
}

var validContextMenuActionTypes = map[string]bool{
	"open_url": true,
	"webhook":  true,
}

// Validate checks the manifest for correctness. Returns nil if valid.
func (m *Manifest) Validate() error {
	if m.ID == "" {
		return fmt.Errorf("manifest: id is required")
	}
	if !pluginIDRe.MatchString(m.ID) {
		return fmt.Errorf("manifest: id %q must be reverse-domain style (e.g. com.example.my-plugin)", m.ID)
	}
	if len(m.ID) > 128 {
		return fmt.Errorf("manifest: id must be <= 128 characters")
	}

	if m.Name == "" {
		return fmt.Errorf("manifest: name is required")
	}
	if len(m.Name) > 64 {
		return fmt.Errorf("manifest: name must be <= 64 characters")
	}

	if m.Version == "" {
		return fmt.Errorf("manifest: version is required")
	}
	if !semverRe.MatchString(m.Version) {
		return fmt.Errorf("manifest: version %q must be semver (e.g. 1.0.0)", m.Version)
	}

	if !validTypes[m.Type] {
		return fmt.Errorf("manifest: type %q is not valid (must be one of: %s)", m.Type, validTypeList())
	}

	// Validate permissions
	for _, p := range m.Permissions {
		if !validPermissions[p] {
			return fmt.Errorf("manifest: unknown permission %q", p)
		}
	}

	// Validate settings
	for i, s := range m.Settings {
		if s.Key == "" {
			return fmt.Errorf("manifest: settings[%d].key is required", i)
		}
		if !validSettingTypes[s.Type] {
			return fmt.Errorf("manifest: settings[%d].type %q is not valid", i, s.Type)
		}
		if s.Label == "" {
			return fmt.Errorf("manifest: settings[%d].label is required", i)
		}
		if s.Type == "select" && len(s.Options) == 0 {
			return fmt.Errorf("manifest: settings[%d] is select type but has no options", i)
		}
	}

	// Validate context menu entries
	for i, e := range m.ContextMenuEntries {
		if e.ID == "" {
			return fmt.Errorf("manifest: context_menu_entries[%d].id is required", i)
		}
		if !validContextMenuTargets[e.Target] {
			return fmt.Errorf("manifest: context_menu_entries[%d].target %q must be member, message, or channel", i, e.Target)
		}
		if e.Label == "" {
			return fmt.Errorf("manifest: context_menu_entries[%d].label is required", i)
		}
		if !validContextMenuActionTypes[e.Action.Type] {
			return fmt.Errorf("manifest: context_menu_entries[%d].action.type %q must be open_url or webhook", i, e.Action.Type)
		}
	}

	// Validate tile config for visual_tile type
	if m.Type == TypeVisualTile && m.Tile == nil {
		return fmt.Errorf("manifest: visual_tile plugins must include a tile configuration")
	}

	return nil
}

// ParseManifest parses and validates a plugin.json byte payload.
func ParseManifest(data []byte) (*Manifest, error) {
	var m Manifest
	if err := json.Unmarshal(data, &m); err != nil {
		return nil, fmt.Errorf("parse manifest: %w", err)
	}
	if m.SchemaVersion == 0 {
		m.SchemaVersion = SchemaVersion
	}
	if err := m.Validate(); err != nil {
		return nil, err
	}
	return &m, nil
}

func validTypeList() string {
	types := make([]string, 0, len(validTypes))
	for t := range validTypes {
		types = append(types, string(t))
	}
	return strings.Join(types, ", ")
}
