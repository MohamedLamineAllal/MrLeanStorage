//go:build windows

package defaults

func GetDefaultConfig() string {
	return `targets:
  # Arc Browser
  - name: "Arc CacheStorage"
    path: "%LOCALAPPDATA%\\Arc\\User Data\\**\\CacheStorage\\**"
    threshold_days: 30
    safety_level: 1
    type: "both"
  - name: "Arc File System"
    path: "%LOCALAPPDATA%\\Arc\\User Data\\**\\File System\\**"
    threshold_days: 30
    safety_level: 1
    type: "both"
  - name: "Arc IndexedDB"
    path: "%LOCALAPPDATA%\\Arc\\User Data\\**\\IndexedDB\\**"
    threshold_days: 30
    safety_level: 1
    type: "both"
  # Google Chrome
  - name: "Chrome Global Cache"
    path: "%LOCALAPPDATA%\\Google\\Chrome\\User Data\\Default\\Cache\\**"
    threshold_days: 30
    safety_level: 1
    type: "both"
  - name: "Chrome CacheStorage"
    path: "%LOCALAPPDATA%\\Google\\Chrome\\User Data\\**\\CacheStorage\\**"
    threshold_days: 30
    safety_level: 1
    type: "both"
  - name: "Chrome File System"
    path: "%LOCALAPPDATA%\\Google\\Chrome\\User Data\\**\\File System\\**"
    threshold_days: 30
    safety_level: 1
    type: "both"
  - name: "Chrome IndexedDB"
    path: "%LOCALAPPDATA%\\Google\\Chrome\\User Data\\**\\IndexedDB\\**"
    threshold_days: 30
    safety_level: 1
    type: "both"
  # Communication Tools
  - name: "Discord Cache"
    path: "%APPDATA%\\discord\\Cache\\**"
    threshold_days: 30
    safety_level: 1
    type: "both"
  - name: "Slack CacheStorage"
    path: "%APPDATA%\\Slack\\Service Worker\\CacheStorage\\**"
    threshold_days: 30
    safety_level: 1
    type: "both"
  # Development Tools
  - name: "VSCode CachedData"
    path: "%APPDATA%\\Code\\CachedData\\**"
    threshold_days: 30
    safety_level: 1
    type: "both"
  - name: "VSCode Service Worker Cache"
    path: "%APPDATA%\\Code\\Service Worker\\CacheStorage\\**"
    threshold_days: 30
    safety_level: 1
    type: "both"
  - name: "VSCode CachedExtensionVSIXs"
    path: "%USERPROFILE%\\.vscode\\extensions\\.obsolete\\**"
    threshold_days: 30
    safety_level: 1
    type: "both"
  # System/Build
  - name: "Go Build Cache"
    path: "%USERPROFILE%\\AppData\\Local\\go-build\\**"
    threshold_days: 30
    safety_level: 1
    type: "both"
  - name: "npm/node-gyp"
    path: "%LOCALAPPDATA%\\node-gyp\\Cache\\**"
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
  - "thumbs.db"
  - "desktop.ini"
schedule: "0 0 0 * * *"
`
}
