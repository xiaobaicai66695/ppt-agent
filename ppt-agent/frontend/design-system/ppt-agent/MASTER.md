# PPT Agent Frontend Design System

Updated: 2026-08-03

## Product Direction

PPT Agent is a production workbench, not a marketing site. The interface should help users move quickly through four actions: create, compose, monitor, and deliver. Real template and slide previews provide the visual interest; application chrome stays quiet and predictable.

Reference patterns are synthesized from mature creation and productivity products:

- Prompt-first entry for a short path from intent to creation.
- Stable resource, page, and property regions in editing workflows.
- Compact navigation and progressive disclosure for dense operational data.
- Real presentation previews instead of decorative placeholder art.

Do not copy external branding, protected assets, or pixel-level layouts.

## Visual Tokens

The source of truth is `src/App.vue`.

| Role | Token | Value |
|---|---|---|
| Canvas | `--canvas` | `#f2f4f4` |
| Primary surface | `--surface` | `#ffffff` |
| Muted surface | `--surface-muted` | `#f6f7f7` |
| Navigation | `--nav-surface` | `#191c1e` |
| Primary text | `--text` | `#171a1c` |
| Secondary text | `--text-secondary` | `#50585e` |
| Border | `--border` | `#dce1e2` |
| Primary action | `--action-ink` | `#075e57` |
| Action highlight | `--action` | `#57d5c7` |
| Information | `--info` | `#2f6fed` |
| Warning accent | `--accent-coral` | `#d9654a` |

Use semantic tokens rather than raw colors in page-level CSS. Purple AI gradients, ambient blobs, glassmorphism, beige monochrome themes, and decorative bokeh are outside this system.

## Geometry And Density

- Radius: 3, 4, 6, or 8px. Use a circle only for avatars, status dots, and compact counters.
- Spacing follows a 4px base with common steps of 8, 12, 16, 24, and 32px.
- Cards are only for repeated items, modals, and framed tools. Do not place cards inside cards.
- Shadows are reserved for drawers, modals, and interactive previews. Static sections rely on borders and surface contrast.
- Main navigation is 216px on desktop and a drawer at 1024px or below.
- Fixed-format slide media always uses a stable 16:9 aspect ratio.

## Typography

- Font stack: Inter, system UI, PingFang SC, Microsoft YaHei, sans-serif.
- Page titles: 15-22px depending on available space.
- Panel titles: 12-15px with tighter vertical rhythm.
- Body text: 13-14px; compact metadata: 10-12px.
- Letter spacing is always `0`.
- Long task names and filenames truncate or wrap within their own regions; they never resize the surrounding layout.

## Components

### Navigation

- Authenticated pages use `AppShell`.
- Desktop primary navigation remains fixed; task-specific navigation is a separate local panel.
- At narrow widths, drawers use a 50-60% black scrim and do not resize main content.

### Buttons And Icons

- Use `lucide-vue-next` for product icons.
- Icon-only actions require `aria-label` and `title` where the action is not obvious.
- Important touch targets are at least 44 by 44px on phone layouts.
- Press states may change color, opacity, or a small transform without shifting adjacent content.

### Forms

- Inputs have visible labels; placeholders are examples, not labels.
- Errors stay near the affected region and include a retry path where recovery is possible.
- Disabled and loading states remain visually distinct and preserve layout dimensions.

### Status And Progress

- User progress and ready slide previews always precede raw logs and runtime diagnostics.
- Preview states are distinct: generating, thumbnail pending, thumbnail failed, file missing, and ready.
- Runtime metadata and timelines live in collapsed diagnostics unless a warning needs a compact summary.

### Tables

- Admin tables use sticky headers, restrained row hover, compact badges, and internal horizontal scrolling.
- Filtering applies to the active data workspace without changing backend contracts.

## Page Rules

- Home: real creation input in the first viewport; templates and recent work support the action.
- Auth: visible field labels, segmented auth mode, complete loading/error feedback, and real preview media.
- Compose: resource panel, slide track, and property panel on desktop; dismissible sheets on narrow screens.
- Dashboard: task context, progress, ready previews, activity, then diagnostics.
- Admin: compact metrics, filters, tabs, tables, and inspectable detail dialogs.

## Responsive And Accessibility

Verify at 375, 768, 1024, and 1440px:

- No page-level horizontal overflow.
- Navigation and property drawers fully enter and leave the viewport.
- Sticky or fixed controls do not hide scroll content.
- Keyboard focus is visible and follows the visual order.
- Images have useful alt text; icon controls have accessible names.
- Motion respects `prefers-reduced-motion`.
- Text contrast targets WCAG AA: 4.5:1 for body text and 3:1 for large text or essential glyphs.
