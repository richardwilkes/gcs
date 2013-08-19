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

package com.trollworks.toolkit.widget;

import com.trollworks.toolkit.collections.TKRange;
import com.trollworks.toolkit.io.TKClipboard;
import com.trollworks.toolkit.text.TKDocument;
import com.trollworks.toolkit.text.TKDocumentListener;
import com.trollworks.toolkit.text.TKTextDrawing;
import com.trollworks.toolkit.utility.TKAlignment;
import com.trollworks.toolkit.utility.TKColor;
import com.trollworks.toolkit.utility.TKFont;
import com.trollworks.toolkit.utility.TKKeystroke;
import com.trollworks.toolkit.utility.TKRectUtils;
import com.trollworks.toolkit.utility.TKTimerTask;
import com.trollworks.toolkit.widget.border.TKBorder;
import com.trollworks.toolkit.widget.border.TKCompoundBorder;
import com.trollworks.toolkit.widget.border.TKEmptyBorder;
import com.trollworks.toolkit.widget.border.TKLineBorder;
import com.trollworks.toolkit.widget.menu.TKMenuItem;
import com.trollworks.toolkit.widget.menu.TKMenuTarget;
import com.trollworks.toolkit.widget.scroll.TKScrollBarOwner;
import com.trollworks.toolkit.window.TKBaseWindow;
import com.trollworks.toolkit.window.TKWindow;

import java.awt.AWTEvent;
import java.awt.Color;
import java.awt.Container;
import java.awt.Cursor;
import java.awt.Dimension;
import java.awt.Font;
import java.awt.Graphics2D;
import java.awt.Insets;
import java.awt.Point;
import java.awt.Rectangle;
import java.awt.Toolkit;
import java.awt.event.FocusEvent;
import java.awt.event.FocusListener;
import java.awt.event.KeyEvent;
import java.awt.event.MouseEvent;

/** A user-enterable text field. */
public class TKTextField extends TKPanel implements Runnable, TKMenuTarget, FocusListener {
	/** The standard text field border. */
	public static final TKBorder	BORDER				= new TKCompoundBorder(TKLineBorder.getSharedBorder(true), new TKEmptyBorder(2));
	private static final int		CURSOR_BLINK_DELAY	= 300;
	private TKDocument				mBackgroundImprintDocument;
	private KeyEvent				mReturnKeyEvent;
	private Color					mDisabledBackground;
	private Color					mDisabledForeground;
	private int						mHorizontalAlignment;
	private int						mMinimumWidth;
	private TKDocument				mDocument;
	private int						mAnchor;
	private Point					mAnchorPt;
	private boolean					mCursorVisible;
	private boolean					mCursorChangePending;
	private boolean					mInMouseDown;
	private boolean					mSingleLineOnly;
	private TKKeyEventFilter		mFilter;
	private Dimension				mPreferredDocumentSize;
	private boolean					mAutoScroll;
	private int						mXOffset;
	private int						mYOffset;
	private boolean					mRightBias;
	private TKTextFieldValidator	mValidator;
	private boolean					mShowInvalid;
	private boolean					mShowSelectionWhenNotFocused;
	private boolean					mNormalDisplay;

	/** Creates an empty, left-justified text field. */
	public TKTextField() {
		this(null, 0, TKAlignment.LEFT, true);
	}

	/**
	 * Creates an empty, left-justified text field.
	 * 
	 * @param singleLineOnly Whether to allow only a single line of text.
	 */
	public TKTextField(boolean singleLineOnly) {
		this(null, 0, TKAlignment.LEFT, singleLineOnly);
	}

	/**
	 * Creates an empty, left-justified text field with the minimum width.
	 * 
	 * @param minimumWidth The minimum width for this field.
	 */
	public TKTextField(int minimumWidth) {
		this(null, minimumWidth, TKAlignment.LEFT, true);
	}

	/**
	 * Creates a left-justified text field with the specified text.
	 * 
	 * @param text The initial text for the field.
	 */
	public TKTextField(String text) {
		this(text, 0, TKAlignment.LEFT, true);
	}

	/**
	 * Creates a left-justified text field with the specified text.
	 * 
	 * @param text The initial text for the field.
	 * @param singleLineOnly Whether to allow only a single line of text.
	 */
	public TKTextField(String text, boolean singleLineOnly) {
		this(text, 0, TKAlignment.LEFT, singleLineOnly);
	}

	/**
	 * Creates a left-justified text field with the specified text and minimum width.
	 * 
	 * @param text The initial text for the field.
	 * @param minimumWidth The minimum width for this field.
	 */
	public TKTextField(String text, int minimumWidth) {
		this(text, minimumWidth, TKAlignment.LEFT);
	}

	/**
	 * Creates a text field with the specified minimum width and alignment.
	 * 
	 * @param minimumWidth The minimum width for this field.
	 * @param alignment The alignment of the text.
	 */
	public TKTextField(int minimumWidth, int alignment) {
		this(null, minimumWidth, alignment);
	}

	/**
	 * Creates a text field with the specified text, minimum width and alignment.
	 * 
	 * @param text The initial text for the field.
	 * @param minimumWidth The minimum width for this field.
	 * @param alignment The alignment of the text.
	 */
	public TKTextField(String text, int minimumWidth, int alignment) {
		this(text, minimumWidth, alignment, true);
	}

	/**
	 * Creates a text field with the specified text, minimum width and alignment.
	 * 
	 * @param text The initial text for the field.
	 * @param minimumWidth The minimum width for this field.
	 * @param alignment The alignment of the text.
	 * @param singleLineOnly Whether to allow only a single line of text.
	 */
	public TKTextField(String text, int minimumWidth, int alignment, boolean singleLineOnly) {
		super();
		mNormalDisplay = true;
		mShowSelectionWhenNotFocused = false;
		mAutoScroll = true;
		mCursorVisible = true;
		mDocument = new TKDocument();
		mBackgroundImprintDocument = new TKDocument();
		mReturnKeyEvent = new KeyEvent(this, KeyEvent.KEY_TYPED, 0, 0, 0, '\n');
		setCursor(Cursor.getPredefinedCursor(Cursor.TEXT_CURSOR));
		setOpaque(true);
		setBackground(TKColor.TEXT_BACKGROUND);
		setForeground(TKColor.TEXT);
		setDisabledBackground(TKColor.DISABLED_TEXT_BACKGROUND);
		setFontKey(TKFont.CONTROL_FONT_KEY);
		setMinimumWidth(minimumWidth);
		setHorizontalAlignment(alignment);
		setBorder(BORDER);
		setText(text);
		setFocusable(true);
		setSingleLineOnly(singleLineOnly);
		enableAWTEvents(AWTEvent.KEY_EVENT_MASK | AWTEvent.MOUSE_EVENT_MASK | AWTEvent.MOUSE_MOTION_EVENT_MASK);
		addFocusListener(this);
	}

	@Override public boolean flagForNoToolTip() {
		String tooltip = getToolTipText();

		return tooltip == null || tooltip.length() == 0;
	}

	/** Temporarily turns of the automatic selectAll() call made when gaining focus. */
	public void disableNextSelectAllForFocus() {
		mInMouseDown = true;
		mNormalDisplay = false;
	}

	/**
	 * Adds a document listener.
	 * 
	 * @param listener The listener to add.
	 */
	public void addDocumentListener(TKDocumentListener listener) {
		mDocument.addDocumentListener(listener);
	}

	/**
	 * Removes a document listener.
	 * 
	 * @param listener The listener to remove.
	 */
	public void removeDocumentListener(TKDocumentListener listener) {
		mDocument.removeDocumentListener(listener);
	}

	private void adjustBoundsForAutoScrolling(Rectangle bounds) {
		if (mAutoScroll && (mPreferredDocumentSize.width > bounds.width || mPreferredDocumentSize.height > bounds.height)) {
			Rectangle selBounds = mDocument.getCharacterBounds(mDocument.getSelectionStart());
			Rectangle viewBounds = new Rectangle(bounds.x - mXOffset, bounds.y - mYOffset, bounds.width, bounds.height);
			int action = -1;
			int viewTmp;
			int selTmp;

			selBounds = TKRectUtils.union(selBounds, mDocument.getCharacterBounds(mDocument.getSelectionEnd()));
			selBounds.x += bounds.x;
			selBounds.y += bounds.y;

			viewTmp = viewBounds.x + viewBounds.width;
			selTmp = selBounds.x + selBounds.width;
			if (selBounds.width > viewBounds.width) {
				action = mRightBias ? 1 : 0;
			} else if (selBounds.x < viewBounds.x) {
				action = 0;
			} else if (selTmp > viewTmp) {
				action = 1;
			}
			if (action == 0) {
				mXOffset -= selBounds.x - viewBounds.x;
			} else if (action == 1) {
				mXOffset += viewTmp - selTmp;
			}
			if (mXOffset > 0) {
				mXOffset = 0;
			}
			bounds.x += mXOffset;

			if (selBounds.y < viewBounds.y) {
				mYOffset -= selBounds.y - viewBounds.y;
			} else if (selBounds.y + selBounds.height > viewBounds.y + viewBounds.height) {
				mYOffset += viewBounds.y + viewBounds.height - (selBounds.y + selBounds.height);
			}
			if (mYOffset > 0) {
				mYOffset = 0;
			}
			bounds.y += mYOffset;

			bounds.width = mPreferredDocumentSize.width;
			bounds.height = mPreferredDocumentSize.height;
		}
	}

	public void menusWillBeAdjusted() {
		notifyActionListeners();
	}

	public boolean adjustMenuItem(String command, TKMenuItem item) {
		boolean processed = true;
		int selStart = getSelectionStart();
		int selEnd = getSelectionEnd();

		if (TKWindow.CMD_CUT.equals(command)) {
			item.setEnabled(selStart != selEnd);
		} else if (TKWindow.CMD_COPY.equals(command)) {
			item.setEnabled(selStart != selEnd);
		} else if (TKWindow.CMD_PASTE.equals(command)) {
			item.setEnabled(TKClipboard.hasText());
		} else if (TKWindow.CMD_CLEAR.equals(command)) {
			item.setEnabled(selStart != selEnd);
		} else if (TKWindow.CMD_SELECT_ALL.equals(command)) {
			item.setEnabled(selStart != 0 || selEnd != getLength());
		} else {
			processed = false;
		}
		return processed;
	}

	public void menusWereAdjusted() {
		// Nothing to do...
	}

	public boolean obeyCommand(String command, TKMenuItem item) {
		boolean processed = true;

		if (TKWindow.CMD_CUT.equals(command)) {
			cut();
		} else if (TKWindow.CMD_COPY.equals(command)) {
			copy();
		} else if (TKWindow.CMD_PASTE.equals(command)) {
			paste();
		} else if (TKWindow.CMD_CLEAR.equals(command)) {
			clear();
		} else if (TKWindow.CMD_SELECT_ALL.equals(command)) {
			selectAll();
		} else {
			processed = false;
		}
		return processed;
	}

	/** Clears the current selection, if any, by deleting the selected text. */
	public void clear() {
		if (getSelectionStart() != getSelectionEnd()) {
			mDocument.deleteBackward();
			runValidator();
			mPreferredDocumentSize = mDocument.getPreferredSize(getFont());
			repaint();
		}
	}

	/**
	 * Copies the currently selected text to the clipboard. If there is no selection, the clipboard
	 * is not changed.
	 */
	public void copy() {
		if (getSelectionStart() != getSelectionEnd()) {
			TKClipboard.putString(getSelection());
		}
	}

	/**
	 * Copies the currently selected text to the clipboard and then deletes it from the field. If
	 * there is no selection, the clipboard is not changed.
	 */
	public void cut() {
		copy();
		clear();
	}

	@Override public Color getBackground() {
		return isEnabled() ? super.getBackground() : getDisabledBackground();
	}

	/** @return The disabled background color. */
	public Color getDisabledBackground() {
		return mDisabledBackground == null ? getBackground() : mDisabledBackground;
	}

	/** @return The disabled foreground color. */
	public Color getDisabledForeground() {
		return mDisabledForeground == null ? getForeground() : mDisabledForeground;
	}

	/** @return The underlying document object. */
	public TKDocument getDocument() {
		return mDocument;
	}

	/** @return The horizontal alignment of the field's contents. */
	public int getHorizontalAlignment() {
		return mHorizontalAlignment;
	}

	/** @return The key event filter for this field, if any. */
	public TKKeyEventFilter getKeyEventFilter() {
		return mFilter;
	}

	/** @return The length of the text in this field. */
	public int getLength() {
		return mDocument.getLength();
	}

	/** @return The maximum size of the component. */
	@Override protected Dimension getMaximumSizeSelf() {
		Dimension size = super.getMaximumSizeSelf();

		if (size.width < mMinimumWidth) {
			size.width = mMinimumWidth;
		}
		return size;
	}

	@Override protected Dimension getMinimumSizeSelf() {
		Insets insets = getInsets();
		Dimension size = super.getMinimumSizeSelf();
		int minHeight = TKTextDrawing.getPreferredSize(getFont(), null, "My").height; //$NON-NLS-1$

		if (size.width < mMinimumWidth) {
			size.width = mMinimumWidth;
		}
		if (size.height < minHeight) {
			size.height = minHeight;
		}
		size.width += insets.left + insets.right;
		size.height += insets.top + insets.bottom;
		return size;
	}

	/**
	 * @return The preferred height of this test field.
	 * @param wrapWidth The width to wrap at.
	 */
	public int getPreferredHeight(int wrapWidth) {
		Insets insets = getInsets();

		return insets.top + insets.bottom + (int) Math.ceil(Math.max(mDocument.getPreferredHeight(getFont(), wrapWidth - (insets.left + insets.right)), mBackgroundImprintDocument.getPreferredHeight(getFont(), wrapWidth - (insets.left + insets.right))));
	}

	private Dimension getLargestSize(Dimension one, Dimension two) {
		int width = one.width;
		int height = one.height;

		if (width < two.width) {
			width = two.width;
		}
		if (height < two.height) {
			height = two.height;
		}
		return new Dimension(width, height);
	}

	@Override protected Dimension getPreferredSizeSelf() {
		Font font = getFont();
		Insets insets = getInsets();
		int minHeight = TKTextDrawing.getPreferredSize(font, null, "My").height; //$NON-NLS-1$
		Dimension size = getLargestSize(mDocument.getPreferredSize(font), mBackgroundImprintDocument.getPreferredSize(font));
		int maxWidth = getMaximumSize().width - (insets.left + insets.right);

		if (size.width > maxWidth) {
			size.width = maxWidth;
			size.height = Math.max(mDocument.getPreferredHeight(font, maxWidth), mBackgroundImprintDocument.getPreferredHeight(font, maxWidth));
		}

		if (size.width < mMinimumWidth) {
			size.width = mMinimumWidth;
		}
		if (size.height < minHeight) {
			size.height = minHeight;
		}
		size.width += insets.left + insets.right;
		size.height += insets.top + insets.bottom;
		return size;
	}

	/** @return The currently selected range of text. */
	public String getSelection() {
		return mDocument.getSelection();
	}

	/** @return The current selection bounds. */
	public Rectangle getSelectionBounds() {
		Insets insets = getInsets();
		int start = mDocument.getSelectionStart();
		int end = mDocument.getSelectionStart();
		Rectangle selBounds = mDocument.getCharacterBounds(start);

		if (start == end) {
			selBounds.width = 1;
		} else {
			selBounds = TKRectUtils.union(selBounds, mDocument.getCharacterBounds(end));
		}
		selBounds.x += insets.left;
		selBounds.y += insets.top;
		selBounds.x += mXOffset;
		selBounds.y += mYOffset;
		return adjustViewBounds(selBounds);
	}

	/** @return The ending index for the current selection. */
	public int getSelectionEnd() {
		return mDocument.getSelectionEnd();
	}

	/** @return The starting index for the current selection. */
	public int getSelectionStart() {
		return mDocument.getSelectionStart();
	}

	/** @return <code>false</code>, unless a filter overrides it or single line only is on. */
	@Override public boolean isDefaultButtonProcessingAllowed() {
		if (isSingleLineOnly()) {
			return true;
		}
		if (mFilter != null) {
			return mFilter.filterKeyEvent(this, mReturnKeyEvent, false);
		}
		return false;
	}

	/** @return <code>true</code> if only single lines are allowed in this field. */
	public boolean isSingleLineOnly() {
		return mSingleLineOnly;
	}

	public void focusGained(FocusEvent event) {
		mDocument.setActive(true);
		if (!mInMouseDown) {
			selectAll();
			scrollIntoView();
		}
		repaintAndResetCursor();
	}

	public void focusLost(FocusEvent event) {
		mDocument.setActive(false);
		notifyActionListeners();
		repaintAndResetCursor();
	}

	@Override public void windowFocus(boolean gained) {
		super.windowFocus(gained);
		mDocument.setActive(gained);
		if (!gained) {
			notifyActionListeners();
		}
		repaintAndResetCursor();
	}

	private Rectangle adjustViewBounds(Rectangle bounds) {
		if (isSingleLineOnly()) {
			int minHeight = TKTextDrawing.getPreferredSize(getFont(), null, "My").height; //$NON-NLS-1$

			if (bounds.height > minHeight) {
				bounds.y += (bounds.height - minHeight) / 2;
			}
		}
		return bounds;
	}

	@Override protected void paintPanel(Graphics2D g2d, Rectangle[] clips) {
		TKBaseWindow window = getBaseWindow();
		Rectangle viewBounds = adjustViewBounds(getLocalInsetBounds());
		boolean showCursor = isEnabled() && mDocument.isActive() && getSelectionStart() == getSelectionEnd() && window.isInForeground();
		boolean drawCursor = false;
		boolean normalDisplay = mNormalDisplay;

		if (showCursor && window.getUserInputManager().getMenuInUse() == null) {
			drawCursor = isFocusOwner() ? mCursorVisible : false;
		}

		if (normalDisplay && isPrinting()) {
			normalDisplay = false;
		}

		g2d.setFont(getFont());
		adjustBoundsForAutoScrolling(viewBounds);
		if (mBackgroundImprintDocument.getLength() > 0 && getLength() == 0) {
			Rectangle imprintBounds;

			if (isSingleLineOnly()) {
				imprintBounds = new Rectangle(viewBounds);
				imprintBounds.width = TKPanel.MAX_SIZE;
				switch (getHorizontalAlignment()) {
					case TKAlignment.CENTER:
						imprintBounds.x -= (TKPanel.MAX_SIZE - viewBounds.width) / 2;
						break;
					case TKAlignment.RIGHT:
						imprintBounds.x -= TKPanel.MAX_SIZE - viewBounds.width;
						break;
				}
			} else {
				imprintBounds = viewBounds;
			}
			g2d.setColor(isEnabled() ? TKColor.DISABLED_TEXT_BACKGROUND : TKColor.CONTROL_HIGHLIGHT);
			mBackgroundImprintDocument.draw(g2d, imprintBounds, getHorizontalAlignment(), false, false, false, null);
		}
		g2d.setColor(isEnabled() ? getForeground() : getDisabledForeground());
		mDocument.draw(g2d, viewBounds, getHorizontalAlignment(), normalDisplay && drawCursor, normalDisplay && isEnabled() && (mShowSelectionWhenNotFocused || isFocusOwner()), mShowInvalid, TKColor.HIGHLIGHTED_TEXT);
		if (showCursor && !mCursorChangePending) {
			mCursorChangePending = true;
			TKTimerTask.schedule(this, CURSOR_BLINK_DELAY);
		}
		mNormalDisplay = true;
	}

	/**
	 * If there is text on the clipboard, inserts it into this field, replacing the current
	 * selection, if any.
	 */
	public void paste() {
		String data = TKClipboard.getString();

		if (data != null) {
			if (mFilter != null) {
				int length = data.length();
				StringBuilder buffer = new StringBuilder(length);
				boolean needBeep = true;

				for (int i = 0; i < length; i++) {
					char ch = data.charAt(i);

					if (!mFilter.filterKeyEvent(this, new KeyEvent(this, KeyEvent.KEY_TYPED, 0, 0, 0, ch), true)) {
						buffer.append(ch);
					} else if (needBeep) {
						Toolkit.getDefaultToolkit().beep();
						needBeep = false;
					}
				}
				if (buffer.length() == 0) {
					return;
				}
				data = buffer.toString();
			}

			if (isSingleLineOnly()) {
				int index = data.indexOf('\n');

				if (index != -1) {
					data = data.substring(0, index);
				}
			}
			mDocument.insert(data);
			runValidator();
			mPreferredDocumentSize = mDocument.getPreferredSize(getFont());
			repaint();
		}
	}

	@Override public void processKeyEvent(KeyEvent event) {
		super.processKeyEvent(event);
		if (!event.isConsumed() && isEnabled()) {
			int id = event.getID();
			char ch;

			if (id == KeyEvent.KEY_PRESSED && TKKeystroke.isCommandKeyDown(event)) {
				TKBaseWindow baseWindow = getBaseWindow();

				if (baseWindow != null && baseWindow.getTKMenuBar() == null) {
					switch (event.getKeyCode()) {
						case KeyEvent.VK_X:
							cut();
							scrollIntoView();
							event.consume();
							return;
						case KeyEvent.VK_C:
							copy();
							event.consume();
							return;
						case KeyEvent.VK_V:
							if (TKClipboard.hasText()) {
								paste();
								scrollIntoView();
								event.consume();
							}
							return;
						case KeyEvent.VK_A:
							selectAll();
							scrollIntoView();
							event.consume();
							return;
					}
				}
			}

			if (mFilter != null) {
				if (mFilter.filterKeyEvent(this, event, true)) {
					if (id == KeyEvent.KEY_TYPED && !TKKeystroke.isCommandKeyDown(event)) {
						Toolkit.getDefaultToolkit().beep();
					}
					return;
				}
			}

			ch = event.getKeyChar();
			if (isSingleLineOnly() && (ch == '\r' || ch == '\n') && (id == KeyEvent.KEY_TYPED || id == KeyEvent.KEY_PRESSED && event.getKeyCode() == 0)) {
				selectAll();
				scrollIntoView();
				event.consume();
				notifyActionListeners();
			} else {
				if (mDocument.respondToKeyEvent(event)) {
					mPreferredDocumentSize = mDocument.getPreferredSize(getFont());
					scrollIntoView();
					repaintAndResetCursor();
				}
				runValidator();
			}
		}
	}

	@Override public void processMouseEventSelf(MouseEvent event) {
		int id = event.getID();
		Insets insets = getInsets();
		int x = event.getX() - insets.left;
		int y = event.getY() - insets.top;

		if (id != MouseEvent.MOUSE_MOVED && id != MouseEvent.MOUSE_DRAGGED) {
			if (mAutoScroll) {
				Rectangle oldViewBounds = getLocalInsetBounds();
				Rectangle viewBounds = new Rectangle(oldViewBounds);

				adjustBoundsForAutoScrolling(viewBounds);
				if (viewBounds.x != oldViewBounds.x) {
					x += oldViewBounds.x - viewBounds.x;
				}
				if (viewBounds.y != oldViewBounds.y) {
					y += oldViewBounds.y - viewBounds.y;
				}
			}

			mAnchorPt = new Point(event.getX(), event.getY());
		}

		switch (id) {
			case MouseEvent.MOUSE_CLICKED:
				int clickCount = event.getClickCount();
				TKRange range;

				if (clickCount == 2) {
					mInMouseDown = true;
					range = mDocument.getWordAt(x, y);
					mAnchor = range.getPosition();
					setSelection(mAnchor, range.getLastPosition() + 1);
					mInMouseDown = false;
				} else if (clickCount == 3) {
					mInMouseDown = true;
					range = mDocument.getLineAt(x, y);
					mAnchor = range.getPosition();
					setSelection(mAnchor, range.getLastPosition() + 1);
					mInMouseDown = false;
				}
				break;
			case MouseEvent.MOUSE_PRESSED:
				mInMouseDown = true;
				requestFocus();
				mAnchor = mDocument.getTextPosition(x, y);
				setSelection(mAnchor, mAnchor);
				mInMouseDown = false;
				break;
			case MouseEvent.MOUSE_DRAGGED:
				Container parent = getParent();
				Point pt = event.getPoint();
				int pos;

				mRightBias = pt.x > mAnchorPt.x;

				while (parent != null && !(parent instanceof TKScrollBarOwner)) {
					parent = parent.getParent();
				}
				if (parent != null) {
					((TKScrollBarOwner) parent).scrollPointIntoView(event, this, pt);
				}

				if (mAutoScroll) {
					Rectangle oldViewBounds = getLocalInsetBounds();
					Rectangle viewBounds = new Rectangle(oldViewBounds);

					adjustBoundsForAutoScrolling(viewBounds);
					if (viewBounds.x != oldViewBounds.x) {
						pt.x += oldViewBounds.x - viewBounds.x;
					}
					if (viewBounds.y != oldViewBounds.y) {
						pt.y += oldViewBounds.y - viewBounds.y;
					}
				}

				pos = mDocument.getTextPosition(pt.x - insets.left, pt.y - insets.top);

				if (pos > mAnchor) {
					setSelection(mAnchor, pos);
				} else {
					setSelection(pos, mAnchor);
				}
				break;
		}
	}

	/** Reset the text cursor and repaint. */
	protected void repaintAndResetCursor() {
		mCursorVisible = true;
		repaint();
	}

	public void run() {
		mCursorChangePending = false;
		mCursorVisible = !mCursorVisible;
		repaint(TKRectUtils.insetRectangle(getSelectionBounds(), -2, -2));
	}

	private void runValidator() {
		if (mValidator != null) {
			setShowInvalidMarker(!mValidator.isTextFieldValid(this));
		}
	}

	/** @param validator The current text field validator. */
	public void setValidator(TKTextFieldValidator validator) {
		mValidator = validator;
		if (mValidator != null) {
			runValidator();
		} else {
			setShowInvalidMarker(false);
		}
	}

	/** Scrolls the current selection into view. */
	public void scrollIntoView() {
		scrollRectIntoView(TKRectUtils.insetRectangle(getSelectionBounds(), -2, -2));
	}

	/** Selects all the text in the field. */
	public void selectAll() {
		setSelection(0, getLength());
	}

	/** @param enabled Whether auto-scrolling is on or off. */
	public void setAutoScrollEnabled(boolean enabled) {
		mAutoScroll = enabled;
	}

	/** @param color The disabled background color. */
	public void setDisabledBackground(Color color) {
		mDisabledBackground = color;
		repaint();
	}

	/** @param color The disabled foreground color. */
	public void setDisabledForeground(Color color) {
		mDisabledForeground = color;
		repaint();
	}

	@Override public void setEnabled(boolean enabled) {
		super.setEnabled(enabled);
		setCursor(isEnabled() ? Cursor.getPredefinedCursor(Cursor.TEXT_CURSOR) : Cursor.getDefaultCursor());
	}

	/** @param font The dynamic font for this text field. */
	@Override public void setFontKey(String font) {
		super.setFontKey(font);
		mPreferredDocumentSize = mDocument.getPreferredSize(getFont());
	}

	/** @param alignment The horizontal alignment of the field's contents. */
	public void setHorizontalAlignment(int alignment) {
		if (alignment != mHorizontalAlignment) {
			mHorizontalAlignment = alignment;
			repaint();
		}
	}

	/** @param filter The key event filter for this field. */
	public void setKeyEventFilter(TKKeyEventFilter filter) {
		mFilter = filter;
	}

	/** @param width The minimum width for this field. */
	public void setMinimumWidth(int width) {
		mMinimumWidth = width;
		if (width > getWidth()) {
			revalidate();
		}
	}

	/**
	 * Select the specified text range.
	 * 
	 * @param startIndex The starting index, inclusive.
	 * @param endIndex The ending index, exclusive.
	 */
	public void setSelection(int startIndex, int endIndex) {
		if (startIndex != getSelectionStart() || endIndex != getSelectionEnd()) {
			mDocument.setSelection(startIndex, endIndex);
			repaintAndResetCursor();
		}
	}

	/** @param showInvalid Whether the invalid marker is on or off. */
	protected void setShowInvalidMarker(boolean showInvalid) {
		if (showInvalid != mShowInvalid) {
			mShowInvalid = showInvalid;
			repaint();
		}
	}

	/** @param singleLine Whether this field only allows single lines. */
	public void setSingleLineOnly(boolean singleLine) {
		if (mSingleLineOnly != singleLine) {
			mSingleLineOnly = singleLine;
			if (singleLine) {
				String text = getText();
				int index = text.indexOf('\n');

				if (index != -1) {
					setText(text.substring(0, index));
				}

				text = mBackgroundImprintDocument.getText();
				index = text.indexOf('\n');
				if (index != -1) {
					setImprint(text.substring(0, index));
				}
			}
		}
	}

	/** @return The text contained in this field. */
	public String getText() {
		return mDocument.getText();
	}

	/**
	 * @return The text contained in this field, or if the field is blank, the text in the
	 *         background imprint.
	 */
	public String getTextOrImprint() {
		String text = getText();

		if (text.length() == 0) {
			text = mBackgroundImprintDocument.getText();
		}
		return text;
	}

	/** @return The text in the background imprint. */
	public String getImprint() {
		return mBackgroundImprintDocument.getText();
	}

	/** @param text The text of the field. */
	public void setText(String text) {
		if (text == null) {
			text = ""; //$NON-NLS-1$
		}
		if (!text.equals(getText())) {
			int length;

			mDocument.setText(text);
			runValidator();
			mPreferredDocumentSize = mDocument.getPreferredSize(getFont());
			length = text.length();
			mXOffset = 0;
			mYOffset = 0;
			setSelection(length, length);
			repaint();
			notifyActionListeners();
		}
	}

	/** @param text The background imprint message. */
	public void setImprint(String text) {
		if (text == null) {
			text = ""; //$NON-NLS-1$
		}
		if (!mBackgroundImprintDocument.getText().equals(text)) {
			mBackgroundImprintDocument.setText(text);
			repaint();
		}
	}

	/** @param text The text to set the text and background imprint to. */
	public void setTextAndImprint(String text) {
		setText(text);
		setImprint(text);
	}

	/** @param show Whether the field should show the selection when not focused. */
	public void setShowSelectionWhenNotFocused(boolean show) {
		if (mShowSelectionWhenNotFocused != show) {
			mShowSelectionWhenNotFocused = show;
			repaint();
		}
	}
}
