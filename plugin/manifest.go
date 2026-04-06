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

	// Declarative hooks (simple event->condition->action rules)
	DeclarativeHooks []DeclarativeHook `json:"declarative_hooks,omitempty"`

	// Scheduled hooks (cron-triggered JS handlers)
	ScheduledHooks []ScheduledHook `json:"scheduled_hooks,omitempty"`

	// Chat "+" menu entries
	PlusMenuEntries []PlusMenuEntry `json:"plus_menu_entries,omitempty"`

	// Message hover action buttons
	MessageActions []MessageAction `json:"message_actions,omitempty"`

	// Input addon widgets (rendered above chat input)
	InputAddons []InputAddon `json:"input_addons,omitempty"`

	// Sidebar panels (rendered in member sidebar)
	SidebarPanels []SidebarPanel `json:"sidebar_panels,omitempty"`

	// Channel header widgets
	HeaderWidgets []HeaderWidget `json:"header_widgets,omitempty"`

	// Custom settings page
	SettingsPage *SettingsPage `json:"settings_page,omitempty"`

	// Voice dock buttons
	DockButtons []DockButton `json:"dock_buttons,omitempty"`

	// Structured form modals (rendered by the platform, not by plugin HTML)
	Modals []ModalDef `json:"modals,omitempty"`

	// Member status line (rich presence in member list)
	StatusLine *StatusLineDef `json:"status_line,omitempty"`

	// Member profile sections (shown in profile popover)
	ProfileSections []ProfileSectionDef `json:"profile_sections,omitempty"`

	// Whitelisted network hosts the plugin's server script can call
	NetworkHosts []string `json:"network_hosts,omitempty"`

	// Webhook URLs for external backend (alternative to WASM)
	WebhookURL string `json:"webhook_url,omitempty"`
}

// DeclarativeHook defines a simple event->condition->action rule.
type DeclarativeHook struct {
	Event     string `json:"event"`              // "message_created", "member_joined", etc.
	Condition string `json:"condition"`           // "message.content starts_with '!ping'"
	Action    string `json:"action"`              // "send_message", "add_reaction"
	Template  string `json:"template"`            // "Pong! Response to {{message.author_name}}"
	Channel   string `json:"channel,omitempty"`   // target channel (default: event's channel)
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
	// Replacement is the text that replaces the /command when sent.
	// If set, the workspace substitutes the command with this text before saving.
	// Used for static text commands like /shrug → ¯\_(ツ)_/¯
	Replacement string `json:"replacement,omitempty"`
}

// TileDef configures a visual_tile plugin's voice grid behavior.
type TileDef struct {
	NameplateIcon      string `json:"nameplate_icon,omitempty"`
	DefaultState       string `json:"default_state,omitempty"` // "grid" (default)
	SupportsFullscreen bool   `json:"supports_fullscreen,omitempty"`
	MaxParticipants    int    `json:"max_participants,omitempty"`
	MinParticipants    int    `json:"min_participants,omitempty"`
	Mode               string `json:"mode,omitempty"` // "tile" (default) or "overlay"
}

// ScheduledHook defines a cron-triggered JS handler.
type ScheduledHook struct {
	Cron    string `json:"cron"`    // 5-field cron expression, e.g. "*/5 * * * *"
	Handler string `json:"handler"` // JS function name to call
}

// PlusMenuEntry registers an item in the chat "+" menu.
type PlusMenuEntry struct {
	ID     string            `json:"id"`
	Label  string            `json:"label"`
	Icon   string            `json:"icon,omitempty"`
	Action ContextMenuAction `json:"action"`
}

// MessageAction registers a hover button on chat messages.
type MessageAction struct {
	ID     string            `json:"id"`
	Label  string            `json:"label"`
	Icon   string            `json:"icon,omitempty"`
	Action ContextMenuAction `json:"action"`
}

// InputAddon declares a widget rendered above the chat input field.
type InputAddon struct {
	ID     string `json:"id"`
	Label  string `json:"label"`
	Height int    `json:"height,omitempty"` // default 200px
}

// SidebarPanel declares a collapsible panel in the member sidebar.
type SidebarPanel struct {
	ID    string `json:"id"`
	Title string `json:"title"`
}

// HeaderWidget declares a small widget in the channel header.
type HeaderWidget struct {
	ID     string `json:"id"`
	Height int    `json:"height,omitempty"` // max 48px
}

// SettingsPage declares a custom settings page for the plugin.
type SettingsPage struct {
	Path  string `json:"path"`  // URL-safe slug
	Title string `json:"title"` // display name in sidebar
}

// StatusLineDef declares how the plugin contributes to a member's status display.
// The template is interpolated with the member's custom fields from this plugin.
type StatusLineDef struct {
	Template  string `json:"template"`            // e.g. "Spielt {{game_name}}"
	Condition string `json:"condition,omitempty"`  // e.g. "game_name != null"
	Icon      string `json:"icon,omitempty"`       // lucide icon name (e.g. "gamepad-2")
	Color     string `json:"color,omitempty"`      // CSS color (e.g. "#2dd4bf")
}

// ProfileSectionDef declares a section in the member profile popover.
type ProfileSectionDef struct {
	ID       string              `json:"id"`
	Title    string              `json:"title"`
	Fields   []ProfileFieldDef   `json:"fields"`
}

// ProfileFieldDef is a single field shown in a profile section.
type ProfileFieldDef struct {
	Label    string `json:"label"`
	Template string `json:"template"` // e.g. "{{steam_id}}" or a static URL template
	URL      string `json:"url,omitempty"` // if set, field value is a clickable link
}

// ModalDef declares a structured form modal that the platform renders.
// Plugins cannot inject arbitrary HTML — they declare fields, and we render
// them with our UI components for consistency and security.
type ModalDef struct {
	ID           string       `json:"id"`
	Title        string       `json:"title"`
	Trigger      string       `json:"trigger,omitempty"`       // command that opens it (e.g. "!steam")
	Fields       []ModalField `json:"fields"`
	StatusField  string       `json:"status_field,omitempty"`  // custom field key shown as status
	StatusLabels *struct {
		Linked   string `json:"linked,omitempty"`
		Unlinked string `json:"unlinked,omitempty"`
	} `json:"status_labels,omitempty"`
	SubmitLabel string `json:"submit_label,omitempty"` // default: "Speichern"
	UnlinkLabel string `json:"unlink_label,omitempty"` // if set, shows an unlink/reset button
	SubmitEvent string `json:"submit_event,omitempty"` // JS handler name (default: "onModalSubmit")
}

// ModalField describes a single input field in a plugin modal form.
type ModalField struct {
	Key         string `json:"key"`
	Type        string `json:"type"` // "text", "number", "checkbox", "select"
	Label       string `json:"label"`
	Placeholder string `json:"placeholder,omitempty"`
	Required    bool   `json:"required,omitempty"`
	Description string `json:"description,omitempty"`
}

// DockButton registers a button in the voice call dock.
type DockButton struct {
	ID     string            `json:"id"`
	Label  string            `json:"label"`
	Icon   string            `json:"icon,omitempty"`
	Action ContextMenuAction `json:"action"`
}

// validHookActions is the set of recognized declarative hook action types.
var validHookActions = map[string]bool{
	"send_message": true,
	"add_reaction": true,
}

// validHookEvents is the set of recognized declarative hook event types.
var validHookEvents = map[string]bool{
	"message_created": true,
	"member_joined":   true,
	"member_left":     true,
	"voice_joined":    true,
	"voice_left":      true,
}

// hostnameRe validates a bare hostname (no scheme, no path, no port).
var hostnameRe = regexp.MustCompile(`^([a-zA-Z0-9]([a-zA-Z0-9-]*[a-zA-Z0-9])?\.)+[a-zA-Z]{2,}$`)

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
	"network:http":          true,
	"scheduled:run":         true,
	"ui:modal":              true,
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

	// Validate declarative hooks
	for i, hook := range m.DeclarativeHooks {
		if !validHookEvents[hook.Event] {
			return fmt.Errorf("manifest: declarative_hooks[%d].event %q is not valid", i, hook.Event)
		}
		if !validHookActions[hook.Action] {
			return fmt.Errorf("manifest: declarative_hooks[%d].action %q is not valid", i, hook.Action)
		}
		if hook.Action == "send_message" && hook.Template == "" {
			return fmt.Errorf("manifest: declarative_hooks[%d] action send_message requires a non-empty template", i)
		}
	}

	// Validate network hosts (must be valid hostnames, no schemes/paths/IPs)
	for i, host := range m.NetworkHosts {
		if !hostnameRe.MatchString(host) {
			return fmt.Errorf("manifest: network_hosts[%d] %q is not a valid hostname (no scheme, path, or IP)", i, host)
		}
	}

	// Validate tile config for visual_tile type
	if m.Type == TypeVisualTile && m.Tile == nil {
		return fmt.Errorf("manifest: visual_tile plugins must include a tile configuration")
	}
	if m.Tile != nil && m.Tile.Mode != "" && m.Tile.Mode != "tile" && m.Tile.Mode != "overlay" {
		return fmt.Errorf("manifest: tile.mode must be \"tile\" or \"overlay\", got %q", m.Tile.Mode)
	}

	// Validate scheduled hooks
	for i, sh := range m.ScheduledHooks {
		if sh.Cron == "" {
			return fmt.Errorf("manifest: scheduled_hooks[%d].cron is required", i)
		}
		if sh.Handler == "" {
			return fmt.Errorf("manifest: scheduled_hooks[%d].handler is required", i)
		}
	}

	// Validate plus menu entries
	for i, e := range m.PlusMenuEntries {
		if e.ID == "" {
			return fmt.Errorf("manifest: plus_menu_entries[%d].id is required", i)
		}
		if e.Label == "" {
			return fmt.Errorf("manifest: plus_menu_entries[%d].label is required", i)
		}
	}

	// Validate message actions
	for i, a := range m.MessageActions {
		if a.ID == "" {
			return fmt.Errorf("manifest: message_actions[%d].id is required", i)
		}
		if a.Label == "" {
			return fmt.Errorf("manifest: message_actions[%d].label is required", i)
		}
	}

	// Validate input addons
	for i, a := range m.InputAddons {
		if a.ID == "" {
			return fmt.Errorf("manifest: input_addons[%d].id is required", i)
		}
		if a.Label == "" {
			return fmt.Errorf("manifest: input_addons[%d].label is required", i)
		}
	}

	// Validate sidebar panels
	for i, p := range m.SidebarPanels {
		if p.ID == "" {
			return fmt.Errorf("manifest: sidebar_panels[%d].id is required", i)
		}
		if p.Title == "" {
			return fmt.Errorf("manifest: sidebar_panels[%d].title is required", i)
		}
	}

	// Validate header widgets
	for i, w := range m.HeaderWidgets {
		if w.ID == "" {
			return fmt.Errorf("manifest: header_widgets[%d].id is required", i)
		}
	}

	// Validate settings page
	if m.SettingsPage != nil {
		if m.SettingsPage.Path == "" {
			return fmt.Errorf("manifest: settings_page.path is required")
		}
		if m.SettingsPage.Title == "" {
			return fmt.Errorf("manifest: settings_page.title is required")
		}
	}

	// Validate modals
	validModalFieldTypes := map[string]bool{"text": true, "number": true, "checkbox": true, "select": true}
	for i, modal := range m.Modals {
		if modal.ID == "" {
			return fmt.Errorf("manifest: modals[%d].id is required", i)
		}
		if modal.Title == "" {
			return fmt.Errorf("manifest: modals[%d].title is required", i)
		}
		if len(modal.Fields) == 0 {
			return fmt.Errorf("manifest: modals[%d].fields must not be empty", i)
		}
		for j, f := range modal.Fields {
			if f.Key == "" {
				return fmt.Errorf("manifest: modals[%d].fields[%d].key is required", i, j)
			}
			if !validModalFieldTypes[f.Type] {
				return fmt.Errorf("manifest: modals[%d].fields[%d].type %q is not valid", i, j, f.Type)
			}
			if f.Label == "" {
				return fmt.Errorf("manifest: modals[%d].fields[%d].label is required", i, j)
			}
		}
	}

	// Validate dock buttons
	for i, b := range m.DockButtons {
		if b.ID == "" {
			return fmt.Errorf("manifest: dock_buttons[%d].id is required", i)
		}
		if b.Label == "" {
			return fmt.Errorf("manifest: dock_buttons[%d].label is required", i)
		}
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
