# Shamogu Chinese localization

- Baseline: official v1.6.0, commit 0c1e2d4a4e50ff4cc10cd922b4e82004b9659e5a.
- All translation and Chinese copy changes are owned by the user-requested Luna subagent.
- Preserve format arguments, style markup, keybindings, serialized identifiers and game rules.
- SDL map: 80x24 cells at 48x90 pixels. Large text panels: 48x24 cells at 80x90 pixels. Both render at 3840x2160.
- Preserve all 175 upstream embedded images. Source uppercase filenames are renamed to avoid case-insensitive filesystem collisions.
- Keep pixel scaling integral, CJK and Latin ink sizes aligned, border geometry continuous, and automatic Steam keyboard disabled before SDL initialization.
- Long menu descriptions must remain scrollable. Test 80-to-48-to-80 rendering and replay forward/backward before release.
- Vendor contains necessary local renderer, Unicode wrapping, replay and keyboard patches. Do not regenerate it without preserving those changes.
- Never claim exhaustive random runtime branch coverage from static string audits.
