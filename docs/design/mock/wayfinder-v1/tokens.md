# Wayfinder v1 design tokens

## Metadata

- Created: 2026-09-30
- Last updated: 2026-10-02
- Author: OpenAI Codex, Android Product Design Specialist
- File: `tokens.md`
- Status: Implementation handoff

Use semantic names in Compose. Hex values are reference values and may be adjusted only to pass measured contrast or platform dynamic-color integration.

## Color

| Role | Light | Dark | Use |
| --- | --- | --- | --- |
| `navContainer` | `#102333` | `#08141D` | Drawer, rail, compact bottom bar |
| `onNavContainer` | `#E2ECF2` | `#E6F1F7` | Navigation labels and icons |
| `navMuted` | `#89A7B9` | `#93ADBC` | Section labels, metadata |
| `navSelected` | `#D7F5FF` | `#123F50` | Selected navigation container |
| `onNavSelected` | `#073147` | `#D7F5FF` | Selected navigation content |
| `primary` | `#087EA4` | `#5CC9ED` | Primary action, links, focus accent |
| `onPrimary` | `#FFFFFF` | `#002A38` | Content on primary |
| `primaryContainer` | `#D7F5FF` | `#123F50` | Active scope and selection |
| `background` | `#F3F6F8` | `#0D171D` | App canvas |
| `surface` | `#FFFFFF` | `#142229` | Cards and panes |
| `surfaceMuted` | `#EEF3F6` | `#1C2D36` | Grouped controls and excerpts |
| `outline` | `#C9D5DC` | `#58717D` | Interactive control boundaries; chosen to retain at least 3:1 against dark surfaces |
| `textStrong` | `#142C3A` | `#E8F2F6` | Titles, answer lead |
| `text` | `#304A59` | `#C6D6DE` | Body text |
| `textMuted` | `#637788` | `#93A8B4` | Metadata and hints |
| `success` | `#23603B` | `#70D995` | Supported/ready status |
| `successContainer` | `#E1F5E8` | `#153B28` | Supported state container |
| `warning` | `#815500` | `#F2BE55` | Caution and partial state |
| `warningContainer` | `#FFF2CC` | `#4A3508` | Import/removal warning |
| `error` | `#B42318` | `#FFB4AB` | Error and destructive action |
| `errorContainer` | `#FDE8E7` | `#5A1A18` | Timeout/failure container |
| `focus` | `#00A7D6` | `#78D9F5` | 3 dp keyboard focus ring |

Status must never rely on color alone. Pair each state with an icon and text label.

### Theme application contract

- The app shell must explicitly paint `background` across the complete window, including behind navigation and system-inset regions. Never rely on the platform window's default color.
- Navigation paints `navContainer`; destination content paints `background`; cards, fields, menus, and panes paint `surface` or `surfaceMuted` before applying their matching content colors.
- Material semantic roles map from these tokens at theme construction. Components consume `MaterialTheme.colorScheme`; screen code must not switch colors by inspecting system dark mode or use light-only color literals.
- The target dark combinations are: `textStrong/background` 15.95:1, `text/background` 12.16:1, `textMuted/background` 7.34:1, `text/surface` 10.91:1, `textMuted/surface` 6.59:1, `onPrimary/primary` 7.95:1, `onNavSelected/navSelected` 9.93:1, and `outline/surface` 3.16:1. Re-measure if a token changes.
- Disabled content may use reduced emphasis, but must remain understandable and include an accessible disabled reason. Do not achieve disabled styling by applying opacity to an entire subtree when that makes text or boundaries unreadable.

## Typography

Default to the platform Roboto family; Atkinson Hyperlegible is an optional later product choice requiring packaging and rendering review.

| Token | Size / line | Weight | Use |
| --- | --- | --- | --- |
| `displaySmall` | 32 / 40 sp | 700 | Empty-state invitation on expanded layouts |
| `headlineLarge` | 28 / 36 sp | 700 | Screen title on tablet |
| `headlineMedium` | 24 / 32 sp | 700 | Answer lead, phone screen title |
| `titleLarge` | 20 / 28 sp | 700 | Pane and detail title |
| `titleMedium` | 16 / 24 sp | 600 | Cards and rows |
| `bodyLarge` | 16 / 26 sp | 400 | Answer prose |
| `bodyMedium` | 14 / 22 sp | 400 | Supporting copy and excerpts |
| `labelLarge` | 14 / 20 sp | 600 | Buttons and navigation |
| `labelMedium` | 12 / 16 sp | 600 | Status and metadata |
| `labelSmall` | 11 / 16 sp | 700 | Section eyebrow, sentence case |

Never force uppercase using locale-sensitive transforms. Eyebrow copy is authored explicitly and sparingly.

## Geometry and spacing

- Base spacing unit: 4 dp.
- Spacing scale: 4, 8, 12, 16, 20, 24, 32, 40, 48 dp.
- Compact horizontal margin: 16 dp; medium: 24 dp; expanded: 28-32 dp.
- Maximum readable answer measure: 720 dp; body prose target: 55-75 characters per line.
- Control radius: 8 dp; card radius: 10 dp; major pane radius: 14 dp; pill radius: 999 dp.
- Standard outline: 1 dp; active/focus outline: 2-3 dp.
- Minimum touch target: 48 x 48 dp; preferred primary action height: 48-52 dp.
- No decorative elevation. Use color and outlines; reserve 2-6 dp tonal elevation for transient menus, sheets, and dialogs.

## Motion

- State feedback: 100-150 ms.
- Pane/destination transition: 180-250 ms.
- Progress is functional, not theatrical; use determinate progress only when real progress exists.
- Do not stream text with decorative character animation.
- Respect system reduced-motion and animator-duration settings; preserve meaning without animation.

## Icons

Use Material Symbols/Icons with familiar metaphors: Ask/spark, Library/books, Session/playing-cards or campaign, Settings/gear, evidence/article, source folder, model/memory, Cancel/stop. Primary navigation always includes text labels. Do not use novelty fantasy glyphs as the sole label.
