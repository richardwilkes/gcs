/* ***** BEGIN LICENSE BLOCK *****
 * Version: MPL 1.1
 *
 * The contents of this file are subject to the Mozilla Public License Version
 * 1.1 (the "License"); you may not use this file except in compliance with
 * the License. You may obtain a copy of the License at
 * http://www.mozilla.org/MPL/
 *
 * Software distributed under the License is distributed on an "AS IS" basis,
 * WITHOUT WARRANTY OF ANY KIND, either express or implied. See the License
 * for the specific language governing rights and limitations under the
 * License.
 *
 * The Original Code is GURPS Character Sheet.
 *
 * The Initial Developer of the Original Code is Richard A. Wilkes.
 * Portions created by the Initial Developer are Copyright (C) 1998-2002,
 * 2005-2007 the Initial Developer. All Rights Reserved.
 *
 * Contributor(s):
 *
 * ***** END LICENSE BLOCK ***** */

package com.trollworks.toolkit.window;

import com.trollworks.toolkit.undo.TKUndoManager;
import com.trollworks.toolkit.utility.TKGraphics;
import com.trollworks.toolkit.widget.TKPanel;
import com.trollworks.toolkit.widget.TKToolBar;
import com.trollworks.toolkit.widget.TKTextField;
import com.trollworks.toolkit.widget.button.TKButton;
import com.trollworks.toolkit.widget.layout.TKCompassLayout;
import com.trollworks.toolkit.widget.layout.TKWindowLayout;
import com.trollworks.toolkit.widget.menu.TKBaseMenu;
import com.trollworks.toolkit.widget.menu.TKMenuBar;

import java.awt.AWTEvent;
import java.awt.Component;
import java.awt.Dialog;
import java.awt.Dimension;
import java.awt.Frame;
import java.awt.Graphics;
import java.awt.Graphics2D;
import java.awt.Insets;
import java.awt.KeyboardFocusManager;
import java.awt.Rectangle;
import java.awt.Toolkit;
import java.awt.Window;
import java.awt.event.ComponentEvent;
import java.awt.event.ComponentListener;
import java.awt.event.KeyEvent;
import java.awt.event.MouseEvent;
import java.awt.event.WindowEvent;
import java.awt.event.WindowFocusListener;
import java.awt.event.WindowListener;
import java.util.HashMap;
import java.util.HashSet;

/** Provides a base OS-level dialog window. */
public class TKDialog extends Dialog implements TKBaseWindow, WindowListener, WindowFocusListener, ComponentListener {
	/** The result code when nothing has been set. */
	public static final int		NOT_SET	= 0;
	/** The result code for "OK". */
	public static final int		OK		= 1;
	/** The result code for "Cancel". */
	public static final int		CANCEL	= 2;
	private TKPanel				mContent;
	private TKRepaintManager	mRepaintManager;
	private TKUndoManager		mUndoManager;
	private TKUserInputManager	mUserInputManager;
	private boolean				mInForeground;
	private int					mResult;
	private TKButton			mCancelButton;
	private TKButton			mDefaultButton;
	private boolean				mIsClosed;

	/**
	 * Creates a new dialog window, centered on the main screen.
	 * 
	 * @param title The title of the dialog, or <code>null</code>.
	 * @param modal Whether the dialog is modal or not.
	 */
	public TKDialog(String title, boolean modal) {
		super(TKGraphics.getAnyFrame(), title, modal);
		initialize();
	}

	/**
	 * Creates a new dialog window, centered on the specified window.
	 * 
	 * @param owner The owning window.
	 * @param title The title of the dialog, or <code>null</code>.
	 * @param modal Whether the dialog is modal or not.
	 */
	public TKDialog(Frame owner, String title, boolean modal) {
		super(owner, title, modal);
		initialize();
	}

	/**
	 * Creates a new dialog window, centered on the specified dialog window.
	 * 
	 * @param owner The owning dialog window.
	 * @param title The title of the dialog, or <code>null</code>.
	 * @param modal Whether the dialog is modal or not.
	 */
	public TKDialog(Dialog owner, String title, boolean modal) {
		super(owner, title, modal);
		initialize();
	}

	private void initialize() {
		mRepaintManager = new TKRepaintManager(this);
		mUndoManager = new TKUndoManager();
		mUserInputManager = new TKUserInputManager(this);

		TKPanel content = new TKPanel(new TKCompassLayout());
		content.setOpaque(true);
		setContent(content);
		setLayout(TKWindowLayout.getInstance());

		enableEvents(AWTEvent.KEY_EVENT_MASK | AWTEvent.MOUSE_EVENT_MASK | AWTEvent.COMPONENT_EVENT_MASK);
		addWindowListener(this);
		addWindowFocusListener(this);
		mIsClosed = false;
		Toolkit.getDefaultToolkit().setDynamicLayout(true);
	}

	/** @return The result. */
	public int getResult() {
		return mResult;
	}

	/** @param result The result. */
	public void setResult(int result) {
		mResult = result;
	}

	/** @see TKRepaintManager#forceRepaint() */
	public void forceRepaint() {
		mRepaintManager.forceRepaint();
	}

	public Rectangle getLocalBounds(boolean insets) {
		Rectangle bounds = getBounds();

		bounds.x = 0;
		bounds.y = 0;
		if (insets) {
			Insets theInsets = getInsets();

			bounds.x += theInsets.left;
			bounds.y += theInsets.top;
			bounds.width -= theInsets.left + theInsets.right;
			bounds.height -= theInsets.top + theInsets.bottom;
		}
		return bounds;
	}

	public boolean isPrinting() {
		return mRepaintManager.isPrinting();
	}

	@Override public Graphics getGraphics() {
		return TKGraphics.configureGraphics((Graphics2D) super.getGraphics());
	}

	public TKPanel getContent() {
		return mContent;
	}

	public void setContent(TKPanel panel) {
		if (mContent != null) {
			mContent.removeFromParent();
			mContent.removeComponentListener(this);
		}
		mContent = panel;
		mContent.addComponentListener(this);
		add(panel);
	}

	@Override public void toFront() {
		if (!isClosed()) {
			Component focus;

			super.toFront();
			if (!isActive() || !isFocused()) {
				focus = getMostRecentFocusOwner();
				if (focus != null) {
					focus.requestFocus();
				} else {
					requestFocus();
				}
			}
		}
	}

	@Override public void paint(Graphics graphics) {
		mRepaintManager.paint(graphics);
	}

	/**
	 * Paints the specified area within the dialog.
	 * 
	 * @param g2d The graphics context to use.
	 * @param clips The areas to draw.
	 */
	public void paint(Graphics2D g2d, Rectangle[] clips) {
		mRepaintManager.paint(g2d, clips);
	}

	/**
	 * Paints the specified area within the dialog without waiting for any lazy updates that may be
	 * pending.
	 * 
	 * @param bounds The area to draw.
	 */
	public void paintImmediately(Rectangle bounds) {
		mRepaintManager.paintImmediately(bounds);
	}

	/**
	 * Paints the specified area within the dialog without waiting for any lazy updates that may be
	 * pending.
	 * 
	 * @param bounds The area to draw.
	 */
	public void paintImmediately(Rectangle[] bounds) {
		mRepaintManager.paintImmediately(bounds);
	}

	@Override public void repaint(long maxDelay, int x, int y, int width, int height) {
		mRepaintManager.repaint(x, y, width, height);
	}

	/**
	 * Marks the specified area within the dialog as needing to be painted.
	 * 
	 * @param bounds The area to redraw.
	 */
	public void repaint(Rectangle bounds) {
		mRepaintManager.repaint(bounds);
	}

	/**
	 * Marks the specified area within the dialog as no longer needing to be painted.
	 * 
	 * @param bounds The area to redraw.
	 */
	public void stopRepaint(Rectangle bounds) {
		mRepaintManager.stopRepaint(bounds);
	}

	/**
	 * Marks the specified area within the dialog as no longer needing to be painted.
	 * 
	 * @param bounds The area to redraw.
	 */
	public void stopRepaint(Rectangle[] bounds) {
		mRepaintManager.stopRepaint(bounds);
	}

	@Override public void update(Graphics graphics) {
		mRepaintManager.paint(graphics);
	}

	/** Forces all areas that need repainting to be painted. */
	public void updateImmediately() {
		mRepaintManager.updateImmediately();
	}

	public void processWindowKeyEvent(KeyEvent event) {
		if (event.getID() == KeyEvent.KEY_PRESSED) {
			Component focus = KeyboardFocusManager.getCurrentKeyboardFocusManager().getFocusOwner();

			if (!(focus instanceof TKPanel) || ((TKPanel) focus).isDefaultButtonProcessingAllowed()) {
				switch (event.getKeyCode()) {
					case KeyEvent.VK_ENTER:
						if (mDefaultButton != null && mDefaultButton.isEnabled()) {
							mDefaultButton.doClick();
							event.consume();
						}
						break;
					case KeyEvent.VK_ESCAPE:
						if (mCancelButton != null && mCancelButton.isEnabled()) {
							mCancelButton.doClick();
							event.consume();
						}
						break;
					default:
						if (!(focus instanceof TKTextField)) {
							HashMap<Character, TKButton> map = new HashMap<Character, TKButton>();
							HashSet<Character> exclude = new HashSet<Character>();
							TKButton button;

							collectButtons(getContent(), map, exclude);

							button = map.get(new Character(Character.toUpperCase(event.getKeyChar())));
							if (button != null && button.isEnabled()) {
								button.doClick();
								event.consume();
							}
						}
						break;
				}
			}
		}
	}

	private void collectButtons(TKPanel panel, HashMap<Character, TKButton> map, HashSet<Character> exclude) {
		if (panel instanceof TKButton) {
			buildKeyMappings((TKButton) panel, map, exclude);
		}

		if (panel.getComponentCount() > 0) {
			for (Component comp : panel.getComponents()) {
				if (comp instanceof TKPanel) {
					collectButtons((TKPanel) comp, map, exclude);
				}
			}
		}
	}

	private void buildKeyMappings(TKButton button, HashMap<Character, TKButton> map, HashSet<Character> exclude) {
		String title = button.getText().toUpperCase();

		if (title.length() > 0) {
			Character ch = new Character(title.charAt(0));

			if (!exclude.contains(ch)) {
				if (map.put(ch, button) != null) {
					map.remove(ch);
					exclude.add(ch);
				}
			}
		}
	}

	@Override protected void processMouseEvent(MouseEvent event) {
		mUserInputManager.processMouseEvent(event);
	}

	@Override protected void processMouseMotionEvent(MouseEvent event) {
		mUserInputManager.processMouseEvent(event);
	}

	public void processMouseEventSuper(MouseEvent event) {
		super.processMouseEvent(event);
	}

	public void processMouseMotionEventSuper(MouseEvent event) {
		super.processMouseMotionEvent(event);
	}

	/**
	 * Attempts to close the dialog. If the dialog does not want to be closed at this time, it will
	 * remain open and return <code>false</code>, otherwise it will close (the default
	 * implementation calls <code>dispose()</code>).
	 * 
	 * @return <code>true</code> if the dialog closed successfully, <code>false</code> if it
	 *         remains open.
	 */
	public boolean attemptClose() {
		dispose();
		return true;
	}

	public boolean isClosed() {
		return mIsClosed;
	}

	@Override public void dispose() {
		super.dispose();
		mIsClosed = true;
		TKWindow.notifyOfWindowClosed(this);
	}

	/**
	 * Moves the dialog to the standard dialog location.
	 * 
	 * @param window The window to move.
	 */
	public static void moveToStdLocation(Window window) {
		Window owner = window.getOwner();
		Rectangle bounds;
		Dimension size;

		if (owner != null && owner != TKGraphics.getHiddenFrame(false)) {
			bounds = owner.getBounds();
		} else {
			bounds = window.getGraphicsConfiguration().getBounds();
		}

		size = window.getSize();

		window.setLocation(bounds.x + (bounds.width - size.width) / 2, bounds.y + (bounds.height - size.height) / 3);
	}

	/**
	 * Provides a modal loop for servicing modal dialogs. This method returns once the dialog is no
	 * longer visible. Upon entry into this method, the dialog will be centered on the screen or its
	 * owning window and displayed on the screen.
	 * 
	 * @return Returns the value set by calls to <code>setResult()</code> during the run.
	 */
	public int doModal() {
		pack();
		moveToStdLocation(this);
		TKGraphics.forceOnScreen(this);
		setVisible(true);
		return getResult();
	}

	public boolean isInForeground() {
		return mInForeground;
	}

	private void setInForeground(boolean inForeground) {
		if (inForeground != mInForeground) {
			mInForeground = inForeground;
			foregroundStateChanged(inForeground);

			if (!mInForeground) {
				TKBaseMenu menu = mUserInputManager.getMenuInUse();

				if (menu != null) {
					menu.closeCompletely(false);
				}
			}

			Component focus = KeyboardFocusManager.getCurrentKeyboardFocusManager().getFocusOwner();
			if (focus instanceof TKPanel) {
				((TKPanel) focus).windowFocus(inForeground);
			}
		}
	}

	/**
	 * Called when the dialog changes its foreground state.
	 * 
	 * @param inForeground <code>true</code> will be passed in when the dialog is in the
	 *            foreground. <code>false</code> will be passed in when the dialog is not in the
	 *            foreground.
	 */
	protected void foregroundStateChanged(@SuppressWarnings("unused") boolean inForeground) {
		// Nothing to do...
	}

	public void windowGainedFocus(WindowEvent event) {
		if (event.getWindow() == this) {
			setInForeground(true);
		}
	}

	public void windowLostFocus(WindowEvent event) {
		if (event.getWindow() == this) {
			setInForeground(false);
		}
	}

	public void windowActivated(WindowEvent event) {
		// Nothing to do...
	}

	public void windowClosed(WindowEvent event) {
		// Nothing to do...
	}

	public void windowClosing(WindowEvent event) {
		if (event.getWindow() == this) {
			attemptClose();
		}
	}

	public void windowDeactivated(WindowEvent event) {
		// Nothing to do...
	}

	public void windowDeiconified(WindowEvent event) {
		// Nothing to do...
	}

	public void windowIconified(WindowEvent event) {
		// Nothing to do...
	}

	public void windowOpened(WindowEvent event) {
		// Nothing to do...
	}

	/**
	 * Overrides <code>processComponentEvent</code> to enforce minimum/maximum sizes.
	 * <p>
	 * {@inheritDoc}
	 */
	@Override protected void processComponentEvent(ComponentEvent event) {
		if (event.getID() == ComponentEvent.COMPONENT_RESIZED) {
			enforceMinMaxSize();
		}
		super.processComponentEvent(event);
	}

	public void componentHidden(ComponentEvent event) {
		// Not used.
	}

	public void componentMoved(ComponentEvent event) {
		// Not used.
	}

	public void componentResized(ComponentEvent event) {
		enforceMinMaxSize();
	}

	public void componentShown(ComponentEvent event) {
		// Not used.
	}

	private void enforceMinMaxSize() {
		Dimension origSize = getSize();
		Dimension otherSize = getMinimumSize();
		int width = origSize.width;
		int height = origSize.height;

		if (width < otherSize.width) {
			width = otherSize.width;
		}
		if (height < otherSize.height) {
			height = otherSize.height;
		}
		otherSize = getMaximumSize();
		if (width > otherSize.width) {
			width = otherSize.width;
		}
		if (height > otherSize.height) {
			height = otherSize.height;
		}
		if (width != origSize.width || height != origSize.height) {
			setSize(width, height);
		}
	}

	public void forceFocusToAccept() {
		Component focus = KeyboardFocusManager.getCurrentKeyboardFocusManager().getFocusOwner();

		if (focus instanceof TKTextField) {
			((TKTextField) focus).notifyActionListeners();
		}
	}

	public TKRepaintManager getRepaintManager() {
		return mRepaintManager;
	}

	public TKUndoManager getUndoManager() {
		return mUndoManager;
	}

	public TKUserInputManager getUserInputManager() {
		return mUserInputManager;
	}

	public TKMenuBar getTKMenuBar() {
		return null;
	}

	public void setTKMenuBar(TKMenuBar menuBar) {
		// Nothing to do...
	}

	public TKToolBar getTKToolBar() {
		return null;
	}

	public void setTKToolBar(TKToolBar toolBar) {
		// Nothing to do...
	}

	/**
	 * @return The button that gets pressed if the user hits the escape key and no other component
	 *         consumes it. Returns <code>null</code> if there is no cancel button or default
	 *         button processing is turned off.
	 */
	public TKButton getCancelButton() {
		return mCancelButton;
	}

	/**
	 * @param button The button that gets pressed if the user hits the escape key and no other
	 *            component consumes it.
	 */
	public void setCancelButton(TKButton button) {
		mCancelButton = button;
	}

	/**
	 * @return The button that gets pressed if the user hits the Enter/Return key and no other
	 *         component consumes it. Returns <code>null</code> if there is no default button or
	 *         default button processing is turned off.
	 */
	public TKButton getDefaultButton() {
		return mDefaultButton;
	}

	/**
	 * @param button The button that gets pressed if the user hits the Enter/Return key and no other
	 *            component consumes it.
	 */
	public void setDefaultButton(TKButton button) {
		mDefaultButton = button;
		if (button != null) {
			button.repaint();
		}
	}
}
