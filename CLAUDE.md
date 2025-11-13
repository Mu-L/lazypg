# Development Notes for AI Assistants

This document contains critical lessons learned and best practices for AI assistants working on this project.

## Critical: Bubble Tea / Lipgloss Width Calculation

**⚠️ ALWAYS USE `GetHorizontalFrameSize()` - DON'T CALCULATE MANUALLY!**

**⚠️⚠️⚠️ THE CORRECT WAY: Use lipgloss built-in methods for frame size calculation!**

### The Problem

When using `lipgloss` borders with `Width()` or `MaxWidth()`, the border and padding are rendered **outside** the content area, causing the total rendered width to exceed expectations and result in cut-off borders.

**THE FUNDAMENTAL MISTAKE:** Setting MaxWidth without considering the actual terminal width. Even if content fits, borders can still overflow!

### THE CORRECT METHOD: Use GetHorizontalFrameSize()

**✅ CORRECT APPROACH (Using lipgloss methods):**
```go
// Define your style FIRST
containerStyle := lipgloss.NewStyle().
    Border(lipgloss.RoundedBorder()).
    BorderForeground(lipgloss.Color("#cba6f7")).
    Padding(1, 2)

// Calculate content width using GetHorizontalFrameSize()
maxWidth := 80 // Terminal width
if windowWidth < maxWidth {
    maxWidth = windowWidth
}
contentWidth := maxWidth - containerStyle.GetHorizontalFrameSize()

// Now render with calculated content width
content := renderContent(contentWidth)
return containerStyle.Render(content)
```

**Why This Works:**
- `GetHorizontalFrameSize()` returns the EXACT width used by borders, padding, and margins
- No manual calculation needed - lipgloss knows its own sizes!
- Adapts automatically if you change border or padding styles

**❌ WRONG APPROACH (Manual Calculation):**
```go
// DON'T DO THIS - Prone to errors!
containerStyle := lipgloss.NewStyle().
    Border(lipgloss.RoundedBorder()).
    Padding(1, 2).
    MaxWidth(76)  // Manually calculated, easy to get wrong

// Problems:
// 1. Have to manually track border (2) + padding (4) + margin
// 2. Easy to forget to update when style changes
// 3. MaxWidth doesn't prevent overflow - still need to calculate!
```

**Key Insight:**
Using `MaxWidth()` alone does NOT solve the problem! You still need to calculate the correct content width by subtracting frame sizes.

### Safe Content Width Calculation

For a bordered container with padding that must fit within terminal width:

```go
// Target: Fit in 80-char terminal
// With Border(2) + Padding(4) = 6 extra chars
// Safe MaxWidth = 80 - 6 - safety_margin(4-10)
MaxWidth(70)  // Leaves 10-char safety margin for emojis, unicode, etc.
```

### Best Practices

1. **Use MaxWidth() instead of Width()**
   - `Width()` forces exact content width (can cause overflow)
   - `MaxWidth()` constrains maximum width (allows content to shrink)

2. **Conservative content width calculation**
   ```go
   safeContentWidth = targetWidth - padding - border - safetyMargin(8-10)
   ```

3. **Test with longest possible text**
   - Help text, titles, status messages
   - Include emojis (can be 2+ chars wide in terminal)
   - Account for Unicode characters

4. **Add width comments**
   ```go
   // Keep under 68 chars to fit MaxWidth(76) with Padding(1,2)
   helpText := "Short text here"
   ```

5. **Common pitfalls to avoid**
   - ❌ Setting `Width(80)` on bordered container
   - ❌ Not accounting for emoji width
   - ❌ Ignoring padding in calculations
   - ❌ Using exact measurements without safety margin

### Real Example from This Project

**Before (Border Cut Off):**
```go
containerStyle := lipgloss.NewStyle().
    Border(lipgloss.RoundedBorder()).
    Padding(1, 2).
    MaxWidth(76)

// Help text: 78 chars
sections = append(sections, "↑↓: Select  │  Tab: Switch  │  Enter: Connect  │  m: Manual  │  Esc: Cancel")
// Result: Right border cut off (76 + 4 + 2 = 82 chars total)
```

**After (Fixed):**
```go
containerStyle := lipgloss.NewStyle().
    Border(lipgloss.RoundedBorder()).
    Padding(1, 2).
    MaxWidth(76)

// Help text: 62 chars (14 char safety margin)
sections = append(sections, "↑↓: Navigate │ Tab: Switch │ Enter: Connect │ m: Manual")
// Result: Fits perfectly (76 + 4 + 2 = 82 → reduced to 72 effective width)
```

### Nested Borders - Using GetHorizontalFrameSize() for Each Layer

**✅ CORRECT WAY (Using GetHorizontalFrameSize for nested components):**
```go
// Outer container
containerStyle := lipgloss.NewStyle().
    Border(lipgloss.RoundedBorder()).
    Padding(1, 2)

maxWidth := 80 // Terminal width
contentWidth := maxWidth - containerStyle.GetHorizontalFrameSize()
// contentWidth is now the available space for inner content

// Inner search box (nested)
searchBoxStyle := lipgloss.NewStyle().
    Border(lipgloss.RoundedBorder()).
    Padding(0, 1)

// Calculate search input width from parent's content width
searchInputWidth := contentWidth - searchBoxStyle.GetHorizontalFrameSize() - 4 // 4 for emoji + margins
searchInput.Width = searchInputWidth

// Render
searchBox := searchBoxStyle.Render("🔍 " + searchInput.View())
content := titleText + "\n" + searchBox + "\n" + moreContent
return containerStyle.Render(content)
```

**Why This Works:**
1. Each layer uses `GetHorizontalFrameSize()` to know its own frame width
2. Parent calculates content width and passes it to children
3. Children subtract their own frame size from parent's content width
4. No manual calculations - all sizes come from lipgloss itself!

**Real Example from This Project:**
```go
// In View() method:
containerStyle := lipgloss.NewStyle().Border(...).Padding(1, 2)
contentWidth := maxWidth - containerStyle.GetHorizontalFrameSize()
content := c.renderDiscoveryMode(contentWidth) // Pass width down
return containerStyle.Render(content)

// In renderDiscoveryMode(contentWidth int):
searchBoxStyle := lipgloss.NewStyle().Border(...).Padding(0, 1)
searchInputWidth := contentWidth - searchBoxStyle.GetHorizontalFrameSize() - 4
c.searchInput.Width = searchInputWidth
return searchBoxStyle.Render("🔍 " + c.searchInput.View())
```

**The Critical Lesson:**
1. **Define styles WITHOUT MaxWidth** - let content determine size
2. **Use GetHorizontalFrameSize()** for each styled component
3. **Pass content width down** through render methods
4. **Each child subtracts its frame size** from parent's content width

### Quick Reference Table

| Terminal Width | Border | Padding | Safe MaxWidth | Safe Content |
|----------------|--------|---------|---------------|--------------|
| 80             | 2      | 4       | 70            | 66           |
| 100            | 2      | 4       | 90            | 86           |
| 120            | 2      | 4       | 110           | 106          |

**Formula:** Safe Content = Terminal Width - Border(2) - Padding(4) - Safety Margin(4-10)

---

## Other Best Practices

### Code Organization

- Keep UI components in `internal/ui/components/`
- Keep business logic in `internal/` packages
- Use models package for shared data structures

### Error Handling

- Always log errors but don't crash on non-critical failures
- Provide user-friendly error messages in UI overlays
- Use `log.Printf("Warning: ...")` for non-fatal errors

### Testing

- Test UI changes at 80-char terminal width (most common)
- Test with actual PostgreSQL instances
- Verify password storage works across restarts

---

*Last updated: 2025-01-11*
