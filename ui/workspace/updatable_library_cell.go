package workspace

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/richardwilkes/gcs/v5/model/library"
	"github.com/richardwilkes/toolbox/i18n"
	"github.com/richardwilkes/toolbox/xmath"
	"github.com/richardwilkes/unison"
)

type updatableLibraryCell struct {
	unison.Panel
	library           *library.Library
	release           library.Release
	title             *unison.Label
	button            *unison.Button
	updateError       error
	inButtonMouseDown bool
	inPanel           bool
	overButton        bool
}

func newUpdatableLibraryCell(lib *library.Library, title *unison.Label, rel *library.Release) *updatableLibraryCell {
	c := &updatableLibraryCell{
		library: lib,
		release: *rel,
		title:   title,
	}
	c.Self = &c
	c.SetLayout(&unison.FlexLayout{
		Columns:  2,
		HSpacing: unison.StdHSpacing,
	})

	c.AddChild(title)

	c.button = unison.NewButton()
	c.button.Text = fmt.Sprintf("Update to v%s", filterVersion(rel.Version))
	fd := c.button.Font.Descriptor()
	fd.Size = xmath.Round(fd.Size * 0.8)
	c.button.Font = fd.Font()
	c.button.ClickCallback = c.initiateUpdate
	c.AddChild(c.button)

	c.MouseDownCallback = c.mouseDown
	c.MouseDragCallback = c.mouseDrag
	c.MouseUpCallback = c.mouseUp
	c.MouseEnterCallback = c.mouseEnter
	c.MouseMoveCallback = c.mouseMove
	c.MouseExitCallback = c.mouseExit
	return c
}

func (c *updatableLibraryCell) initiateUpdate() {
	if w := FromWindowOrAny(c.Window()); w != nil {
		var list []unison.TabCloser
		w.DocumentDock.RootDockLayout().ForEachDockContainer(func(dc *unison.DockContainer) bool {
			p := c.library.PathOnDisk + "/"
			for _, one := range dc.Dockables() {
				if tc, ok := one.(unison.TabCloser); ok {
					var fbd FileBackedDockable
					if fbd, ok = one.(FileBackedDockable); ok {
						if strings.HasPrefix(fbd.BackingFilePath(), p) {
							list = append(list, tc)
						}
					}
				}
			}
			return false
		})
		for _, one := range list {
			if !one.MayAttemptClose() || !one.AttemptClose() {
				unison.WarningDialogWithMessage(i18n.Text("Update canceled"),
					i18n.Text("The library cannot be updated while\ndocuments from the library are open."))
				return
			}
		}

		var frame unison.Rect
		if focused := unison.ActiveWindow(); focused != nil {
			frame = focused.FrameRect()
		} else {
			frame = unison.PrimaryDisplay().Usable
		}
		wnd, err := unison.NewWindow(i18n.Text("Updating…"), unison.FloatingWindowOption(),
			unison.NotResizableWindowOption(), unison.UndecoratedWindowOption(), unison.TransientWindowOption())
		if err != nil {
			unison.ErrorDialogWithError(i18n.Text("Unable to update"), err)
			return
		}
		content := unison.NewPanel()
		content.SetBorder(unison.NewEmptyBorder(unison.NewUniformInsets(2 * unison.StdHSpacing)))
		content.SetLayout(&unison.FlexLayout{
			Columns:  1,
			VSpacing: unison.StdVSpacing,
		})
		label := unison.NewLabel()
		label.Text = fmt.Sprintf(i18n.Text("Updating %s to v%s…"), c.library.Title, filterVersion(c.release.Version))
		content.AddChild(label)
		progress := unison.NewProgressBar(0)
		progress.SetLayoutData(&unison.FlexLayoutData{
			MinSize: unison.Size{Width: 500},
			HAlign:  unison.FillAlignment,
			HGrab:   true,
		})
		content.AddChild(progress)
		wnd.SetContent(content)
		wnd.Pack()
		wndFrame := wnd.FrameRect()
		frame.Y += (frame.Height - wndFrame.Height) / 3
		frame.Height = wndFrame.Height
		frame.X += (frame.Width - wndFrame.Width) / 2
		frame.Width = wndFrame.Width
		frame.Align()
		wnd.SetFrameRect(frame)
		wnd.ToFront()
		c.updateError = nil
		go c.performUpdate(wnd)
		wnd.RunModal()
		if c.updateError != nil {
			unison.ErrorDialogWithError(i18n.Text("Unable to update"), err)
			return
		}
	}
}

func (c *updatableLibraryCell) performUpdate(wnd *unison.Window) {
	defer c.finishUpdate(wnd)
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	c.library.StopAllWatches()
	c.updateError = c.library.Download(ctx, &http.Client{}, c.release)
}

func (c *updatableLibraryCell) finishUpdate(wnd *unison.Window) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	c.library.CheckForAvailableUpgrade(ctx, &http.Client{})
	FromWindowOrAny(c.Window()).Navigator.eventuallyReload()
	wnd.StopModal(unison.ModalResponseOK)
}

func (c *updatableLibraryCell) updateForeground(fg unison.Ink) {
	c.title.OnBackgroundInk = fg
}

func (c *updatableLibraryCell) mouseDown(where unison.Point, btn, clickCount int, mod unison.Modifiers) bool {
	if !c.button.FrameRect().ContainsPoint(where) {
		return false
	}
	c.inButtonMouseDown = true
	return c.button.DefaultMouseDown(c.button.PointFromRoot(c.PointToRoot(where)), btn, clickCount, mod)
}

func (c *updatableLibraryCell) mouseDrag(where unison.Point, btn int, mod unison.Modifiers) bool {
	if !c.inButtonMouseDown {
		return false
	}
	return c.button.DefaultMouseDrag(c.button.PointFromRoot(c.PointToRoot(where)), btn, mod)
}

func (c *updatableLibraryCell) mouseUp(where unison.Point, btn int, mod unison.Modifiers) bool {
	if !c.inButtonMouseDown {
		return false
	}
	c.inButtonMouseDown = false
	return c.button.DefaultMouseUp(c.button.PointFromRoot(c.PointToRoot(where)), btn, mod)
}

func (c *updatableLibraryCell) mouseEnter(where unison.Point, mod unison.Modifiers) bool {
	c.inPanel = true
	if !c.button.FrameRect().ContainsPoint(where) {
		return false
	}
	c.overButton = true
	return c.button.DefaultMouseEnter(c.button.PointFromRoot(c.PointToRoot(where)), mod)
}

func (c *updatableLibraryCell) mouseMove(where unison.Point, mod unison.Modifiers) bool {
	if c.inPanel {
		over := c.button.FrameRect().ContainsPoint(where)
		if over != c.overButton {
			if over {
				c.overButton = true
				return c.button.DefaultMouseEnter(c.button.PointFromRoot(c.PointToRoot(where)), mod)
			}
			c.overButton = false
			return c.button.DefaultMouseExit()
		}
	}
	return false
}

func (c *updatableLibraryCell) mouseExit() bool {
	c.inPanel = false
	if !c.overButton {
		return false
	}
	c.overButton = false
	return c.button.DefaultMouseExit()
}
