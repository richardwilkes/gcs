# Menu Keys

Menu keys are **keyboard shortcuts** that let you perform actions in GCS more quickly without navigating menus.

- On **Windows** and **Linux**, shortcuts usually use the **`Ctrl`** key (for example, `Ctrl+S` to save).
- On **Mac**, the same shortcuts typically use the **`Command (⌘)`** key instead (for example, `⌘+S` to save).

GCS comes with several shortcuts pre-set for common actions (like copy, paste, or scaling the sheet). You can **add new
shortcuts** or **change existing ones** to match your preferences. Every menu command can be given a shortcut,
including the settings views and the **New** commands in the **Insert** menu.

To open the **Menu Keys** settings:

- Go to **Settings > Menu Keys.**

The commands are grouped by the menu they appear in, with submenus indented under their menu, in the same order as the
menu bar. Type in the **Filter** field at the top to show only the commands whose name or key matches what you type.

## Managing Settings

You can manage Menu Keys settings the same way as other settings:

- **Reset settings to defaults:** Select **Reset** ![reset](images/icons/icn-reset.svg).
- **Import or export settings:** On the **Menu Keys toolbar**, select **Menu** ![menu](images/icons/icn-menu.svg), then
  choose:
  - **Import…:** Load saved settings into GCS.
  - **Export…:** Save your current settings as a file.
  - Any `.keys` files found in the `Settings` folder of your libraries are listed at the bottom of the menu, ready to
    import.

## Changing a shortcut

To change a shortcut:

1. Go to **Settings > Menu Keys**.
2. Select the **shortcut box** next to the action you want to edit. It reads **None** when the action has no shortcut.
3. In the pop-up, press the keyboard combination you want to use (for example, `Ctrl+Shift+S` or `Cmd+Shift+S`). A
   modifier key on its own isn't captured, and `Escape`, `Return`, and `Tab` on their own cancel, set, or move the
   focus instead of being captured.
4. Select **Set** to save your new shortcut.

If the combination is already used by another action, the pop-up says so. Selecting **Set** removes it from the other
action and gives it to this one.

You can remove a shortcut by selecting **Clear** in the pop-up. To put a single action back to its default, select
**Reset** ![reset](images/icons/icn-reset.svg) beside it. This button is only enabled while the shortcut differs from its
default, so it also shows you which shortcuts you've changed.

> [!WARNING]
> Some shortcuts may **conflict with system-level hotkeys** (for example, `Cmd+Space` on macOS opens Spotlight). If a
> shortcut doesn’t seem to work in GCS, try assigning a different combination that isn’t already reserved by your
> operating system.

For other settings options, see [General Settings](General%20Settings).

---

[Return to Home](Home)
