//go:build linux

package defaults

func GetDefaultConfig() string {
	return `targets:
  # Chrome
  - name: "Chrome Global Cache"
    path: "~/.cache/google-chrome/**"
    threshold_days: 30
    safety_level: 1
    type: "both"
  - name: "Chrome CacheStorage"
    path: "~/.config/google-chrome/**/CacheStorage/**"
    threshold_days: 30
    safety_level: 1
    type: "both"
  # Communication Tools
  - name: "Discord Cache"
    path: "~/.config/discord/Cache/**"
    threshold_days: 30
    safety_level: 1
    type: "both"
  # Development Tools
  - name: "VSCode CachedData"
    path: "~/.config/Code/CachedData/**"
    threshold_days: 30
    safety_level: 1
    type: "both"
  - name: "VSCode Service Worker Cache"
    path: "~/.config/Code/Service Worker/CacheStorage/**"
    threshold_days: 30
    safety_level: 1
    type: "both"
  # System/Build
  - name: "Go Build Cache"
    path: "~/.cache/go-build/**"
    threshold_days: 30
    safety_level: 1
    type: "both"
  - name: "npm/node-gyp"
    path: "~/.cache/node-gyp/**"
    threshold_days: 30
    safety_level: 1
    type: "both"
  # Commands
  - name: "PNPM Store Prune"
    command: "pnpm store prune"
    interval_days: 30
    safety_level: 1
  - name: "npm clean cache"
    command: "npm cache clean --force"
    interval_days: 30
    safety_level: 1
dry_run: true
ignore_patterns:
  - ".DS_Store"
  - ".thumbnails"
schedule: "0 0 0 * * *"
`
}
