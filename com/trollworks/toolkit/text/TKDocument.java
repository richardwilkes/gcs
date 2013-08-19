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

package com.trollworks.toolkit.text;

import com.trollworks.toolkit.collections.TKRange;
import com.trollworks.toolkit.utility.TKAlignment;
import com.trollworks.toolkit.utility.TKColor;
import com.trollworks.toolkit.widget.TKPanel;

import java.awt.Color;
import java.awt.Dimension;
import java.awt.Font;
import java.awt.Graphics2D;
import java.awt.Rectangle;
import java.awt.event.InputEvent;
import java.awt.event.KeyEvent;
import java.util.ArrayList;

/** Represents a collection of paragraphs. */
public class TKDocument {
	/** Characters that make up whitespace or close-to-whitespace. */
	public static final String				GRAYSPACE	= " \t\n()[]=+-{}<>,&:?;#\"/.\'\\*";	//$NON-NLS-1$
	private ArrayList<TKParagraph>			mParagraphs;
	private int								mSelectionStart;
	private int								mSelectionEnd;
	private boolean							mActive;
	private ArrayList<TKDocumentListener>	mDocumentListeners;

	/** Creates a new, empty document. */
	public TKDocument() {
		this(null);
	}

	/**
	 * Creates a new document from the specified text.
	 * 
	 * @param text The initial text.
	 */
	public TKDocument(String text) {
		mParagraphs = new ArrayList<TKParagraph>();
		setText(text);
	}

	/**
	 * Adds a document listener.
	 * 
	 * @param listener The listener to add.
	 */
	public void addDocumentListener(TKDocumentListener listener) {
		if (mDocumentListeners == null) {
			mDocumentListeners = new ArrayList<TKDocumentListener>(1);
		}
		if (!mDocumentListeners.contains(listener)) {
			mDocumentListeners.add(listener);
		}
	}

	/**
	 * Removes a document listener.
	 * 
	 * @param listener The listener to remove.
	 */
	public void removeDocumentListener(TKDocumentListener listener) {
		if (mDocumentListeners != null) {
			mDocumentListeners.remove(listener);
			if (mDocumentListeners.isEmpty()) {
				mDocumentListeners = null;
			}
		}
	}

	private void notifyDocumentListeners() {
		if (mDocumentListeners != null) {
			for (TKDocumentListener listener : mDocumentListeners) {
				listener.documentChanged(this);
			}
		}
	}

	/** Deletes the character just before the insertion cursor. */
	public void deleteBackward() {
		StringBuilder buffer = new StringBuilder(getText());

		buffer.append('\n');

		if (mSelectionStart != mSelectionEnd) {
			buffer.delete(mSelectionStart, mSelectionEnd);
		} else if (mSelectionStart > 0) {
			mSelectionStart--;
			buffer.delete(mSelectionStart, mSelectionStart + 1);
		}
		mSelectionEnd = mSelectionStart;
		setText(buffer.toString());
	}

	/** Deletes the character just after the insertion cursor. */
	public void deleteForward() {
		String text = getText();
		StringBuilder buffer = new StringBuilder(text);
		int length = text.length() + 1;

		buffer.append('\n');

		if (mSelectionStart != mSelectionEnd) {
			buffer.delete(mSelectionStart, mSelectionEnd);
		} else if (mSelectionStart < length) {
			buffer.delete(mSelectionStart, mSelectionStart + 1);
		}
		mSelectionEnd = mSelectionStart;
		setText(buffer.toString());
	}

	/**
	 * Draw the document. The desired font and foreground color should be applied to the graphics
	 * object before calling this method.
	 * 
	 * @param g2d The graphics object to use.
	 * @param bounds The bounding rectangle of the drawing area.
	 * @param alignment The horizontal alignment of the text.
	 * @param drawCursor Pass in <code>true</code> to draw the cursor.
	 * @param drawInvalidMarker Whether or not the "invalid" marker should be drawn.
	 * @param drawSelection Pass in <code>true</code> to draw the selected area with a highlight.
	 * @param selectedTextColor The color to use for selected text.
	 * @return The last vertical position used within the drawing area.
	 */
	public int draw(Graphics2D g2d, Rectangle bounds, int alignment, boolean drawCursor, boolean drawSelection, boolean drawInvalidMarker, Color selectedTextColor) {
		Rectangle clip = g2d.getClipBounds();
		int cY = clip.y;
		int bY = bounds.y;
		int bH = bounds.height;
		int minY = Math.max(cY, bY);
		int maxY = Math.min(cY + clip.height, bY + bH);
		int wrapWidth = bounds.width;
		Rectangle pBounds = new Rectangle(bounds.x, bY, wrapWidth, bH);
		int size = getParagraphCount();
		int length = 0;
		int height;
		int i;
		TKParagraph paragraph;

		for (i = 0; i < size && pBounds.y < maxY; i++) {
			paragraph = getParagraph(i);
			paragraph.prepareLayouts(g2d.getFont(), alignment, wrapWidth);
			height = paragraph.getHeight();
			if (pBounds.y + height >= minY) {
				break;
			}
			pBounds.y += height;
			length += 1 + paragraph.getLength();
		}

		pBounds.height = maxY - pBounds.y;

		for (; i < size && pBounds.height > 0; i++) {
			paragraph = getParagraph(i);
			pBounds.y = paragraph.draw(g2d, pBounds, alignment, drawCursor, drawSelection, isActive() ? TKColor.HIGHLIGHT : TKColor.INACTIVE_HIGHLIGHT, selectedTextColor, mSelectionStart - length, mSelectionEnd - length, drawInvalidMarker);
			pBounds.height = maxY - pBounds.y;
			length += 1 + paragraph.getLength();
		}
		return pBounds.y;
	}

	/**
	 * @param characterPosition The position to retrieve a bounds for.
	 * @return The bounding rectangle, relative to the document, of the character at the specified
	 *         position within the text. An empty rectangle will be returned if this document has
	 *         never been drawn through a call to
	 *         {@link #draw(Graphics2D,Rectangle,int,boolean,boolean,boolean,Color)}.
	 */
	public Rectangle getCharacterBounds(int characterPosition) {
		int pCount = getParagraphCount();
		int length = 0;
		int y = 0;

		for (int i = 0; i < pCount; i++) {
			TKParagraph paragraph = getParagraph(i);
			int height = paragraph.getHeight();

			if (length + 1 + paragraph.getLength() > characterPosition) {
				Rectangle bounds = paragraph.getCharacterBounds(characterPosition - length);

				bounds.y += y;
				return bounds;
			}
			y += height;
			length += 1 + paragraph.getLength();
		}
		return new Rectangle();
	}

	/** @return The length of all text within the document. */
	public int getLength() {
		int size = getParagraphCount();
		int length = 0;

		for (int i = 0; i < size; i++) {
			length += 1 + getParagraph(i).getLength();
		}
		return length > 0 ? length - 1 : 0;
	}

	/**
	 * @param x The x-coordinate.
	 * @param y The y-coordinate.
	 * @return The range that represents the line at the specified position.
	 */
	public TKRange getLineAt(int x, int y) {
		String text = getText();
		int max = text.length() - 1;
		int start = getTextPosition(x, y);
		int end = start;

		while (start > 0 && text.charAt(start - 1) != '\n') {
			start--;
		}

		while (end < max && text.charAt(end + 1) != '\n') {
			end++;
		}
		return new TKRange(start, 1 + end - start);
	}

	/**
	 * @param index The index of the paragraph.
	 * @return The paragraph at the specified index.
	 */
	public TKParagraph getParagraph(int index) {
		if (index < 0 || index > getParagraphCount()) {
			throw new IndexOutOfBoundsException("Attempting to retrieve paragraph " + index + ", but there are only " + getParagraphCount() + " paragraphs in this document."); //$NON-NLS-1$ //$NON-NLS-2$ //$NON-NLS-3$
		}
		return mParagraphs.get(index);
	}

	/** @return The number of paragraphs contained in this document. */
	public int getParagraphCount() {
		return mParagraphs.size();
	}

	/**
	 * @param font The font to use.
	 * @param wrapWidth The width to wrap at.
	 * @return The preferred height of this document.
	 */
	public int getPreferredHeight(Font font, int wrapWidth) {
		int count = getParagraphCount();
		int height = 0;

		for (int i = 0; i < count; i++) {
			TKParagraph paragraph = getParagraph(i);

			paragraph.prepareLayouts(font, TKAlignment.LEFT, wrapWidth);
			height += paragraph.getHeight();
		}
		return height;
	}

	/**
	 * @param font The font to use.
	 * @return The preferred size of this document.
	 */
	public Dimension getPreferredSize(Font font) {
		int size = getParagraphCount();
		Dimension dSize = new Dimension();

		for (int i = 0; i < size; i++) {
			TKParagraph paragraph = getParagraph(i);
			Dimension pSize;

			paragraph.prepareLayouts(font, TKAlignment.LEFT, TKPanel.MAX_SIZE);
			pSize = paragraph.getSize();
			if (pSize.width > dSize.width) {
				dSize.width = pSize.width;
			}
			dSize.height += pSize.height;
		}
		return dSize;
	}

	/**
	 * @return The currently selected range of text, or an empty string if there isn't one.
	 */
	public String getSelection() {
		if (mSelectionStart != mSelectionEnd) {
			return getText().substring(mSelectionStart, mSelectionEnd);
		}
		return ""; //$NON-NLS-1$
	}

	/** @return The ending index for the current selection. */
	public int getSelectionEnd() {
		return mSelectionEnd;
	}

	/** @return The starting index for the current selection. */
	public int getSelectionStart() {
		return mSelectionStart;
	}

	/**
	 * @return The dimensions of this document. This returns zero for width and height if the
	 *         document has never been drawn through a call to
	 *         {@link #draw(Graphics2D,Rectangle,int,boolean,boolean,boolean,Color)}.
	 */
	public Dimension getSize() {
		int length = getParagraphCount();
		int height = 0;
		int width = 0;

		for (int i = 0; i < length; i++) {
			TKParagraph paragraph = getParagraph(i);
			Dimension pSize = paragraph.getSize();

			if (pSize.width > width) {
				width = pSize.width;
			}
			height += pSize.height;
		}
		return new Dimension(width, height);
	}

	/** @return The text of this document. */
	public String getText() {
		int size = getParagraphCount();
		int length = getLength();
		StringBuilder buffer = new StringBuilder(length);

		for (int i = 0; i < size; i++) {
			buffer.append(getParagraph(i).getDataBuffer());
			buffer.append('\n');
		}

		length = buffer.length();
		if (length > 0) {
			buffer.setLength(length - 1);
		}

		return buffer.toString();
	}

	/**
	 * @param x The x-coordinate.
	 * @param y The y-coordinate.
	 * @return The position of the nearest character within the text at the specified location.
	 */
	public int getTextPosition(int x, int y) {
		int size = getParagraphCount();
		int length = 0;
		int height = 0;

		for (int i = 0; i < size; i++) {
			TKParagraph paragraph = getParagraph(i);
			int pHeight = paragraph.getHeight();

			if (height + pHeight > y) {
				return length + paragraph.getTextPosition(x, y - height);
			}
			height += pHeight;
			length += 1 + paragraph.getLength();
		}
		return length > 0 ? length - 1 : 0;
	}

	/**
	 * @param x The x-coordinate.
	 * @param y The y-coordinate.
	 * @return The range that represents the word at the specified position.
	 */
	public TKRange getWordAt(int x, int y) {
		String text = getText();
		int max = text.length() - 1;
		int start;
		int end;

		if (max == -1) {
			return new TKRange();
		}

		start = getTextPosition(x, y);
		if (start > max) {
			start = max;
		}
		end = start;

		if (TKDocument.GRAYSPACE.indexOf(text.charAt(start)) == -1) {
			while (start > 0 && TKDocument.GRAYSPACE.indexOf(text.charAt(start - 1)) == -1) {
				start--;
			}

			while (end < max && TKDocument.GRAYSPACE.indexOf(text.charAt(end + 1)) == -1) {
				end++;
			}
		}

		return new TKRange(start, 1 + end - start);
	}

	/**
	 * Insert a character at the current insertion point. If there is a selection, it will be
	 * deleted and then the character will be inserted.
	 * 
	 * @param ch The character to insert.
	 */
	public void insert(char ch) {
		if (ch == '\b') {
			deleteBackward();
		} else {
			StringBuilder buffer = new StringBuilder(getText());

			buffer.append('\n');

			if (mSelectionStart != mSelectionEnd) {
				buffer.delete(mSelectionStart, mSelectionEnd);
			}

			buffer.insert(mSelectionStart, ch);
			mSelectionEnd = ++mSelectionStart;
			setText(buffer.toString());
		}
	}

	/**
	 * Inserts a string at the current insertion point. If there is a selection, it will be deleted
	 * and then the string will be inserted.
	 * 
	 * @param data The string to insert.
	 */
	public void insert(String data) {
		StringBuilder buffer = new StringBuilder(getText());

		buffer.append('\n');

		if (mSelectionStart != mSelectionEnd) {
			buffer.delete(mSelectionStart, mSelectionEnd);
		}

		buffer.insert(mSelectionStart, data);
		mSelectionStart += data.length();
		mSelectionEnd = mSelectionStart;
		setText(buffer.toString());
	}

	/** @return <code>true</code> if this document is active. */
	public boolean isActive() {
		return mActive;
	}

	/**
	 * Responds to the specified key event, consuming it if possible.
	 * 
	 * @param event The key event to process.
	 * @return <code>true</code> if the document should be redrawn.
	 */
	public boolean respondToKeyEvent(KeyEvent event) {
		boolean needRepaint = false;

		if ((event.getModifiers() & (InputEvent.ALT_MASK | InputEvent.CTRL_MASK | InputEvent.META_MASK)) == 0) {
			needRepaint = true;

			switch (event.getID()) {
				case KeyEvent.KEY_PRESSED:
					int code = event.getKeyCode();
					boolean shift = event.isShiftDown();
					Rectangle bounds;
					int pos;

					if (mSelectionStart != mSelectionEnd && !shift) {
						if (code == KeyEvent.VK_LEFT || code == KeyEvent.VK_UP || code == KeyEvent.VK_HOME) {
							mSelectionEnd = mSelectionStart;
							event.consume();
							return true;
						} else if (code == KeyEvent.VK_RIGHT || code == KeyEvent.VK_DOWN || code == KeyEvent.VK_END) {
							mSelectionStart = mSelectionEnd;
							event.consume();
							return true;
						}
					}

					if (code == KeyEvent.VK_LEFT) {
						if (mSelectionStart == 0) {
							needRepaint = false;
						} else {
							mSelectionStart--;
							if (!shift) {
								mSelectionEnd = mSelectionStart;
							}
						}
					} else if (code == KeyEvent.VK_RIGHT) {
						if (mSelectionEnd > getLength() - 1) {
							needRepaint = false;
						} else {
							mSelectionEnd++;
							if (!shift) {
								mSelectionStart = mSelectionEnd;
							}
						}
					} else if (code == KeyEvent.VK_UP) {
						if (mSelectionStart == 0) {
							needRepaint = false;
						} else {
							bounds = getCharacterBounds(mSelectionStart);
							pos = getTextPosition(bounds.x, bounds.y - 2);
							mSelectionStart = pos != mSelectionStart ? pos : 0;
							if (!shift) {
								mSelectionEnd = mSelectionStart;
							}
						}
					} else if (code == KeyEvent.VK_DOWN) {
						if (mSelectionEnd > getLength() - 1) {
							needRepaint = false;
						} else {
							bounds = getCharacterBounds(mSelectionStart);
							mSelectionEnd = getTextPosition(bounds.x, bounds.y + bounds.height + 2);
							if (!shift) {
								mSelectionStart = mSelectionEnd;
							}
						}
					} else if (code == KeyEvent.VK_HOME) {
						bounds = getCharacterBounds(mSelectionStart);
						mSelectionStart = getTextPosition(0, bounds.y + 2);
						if (!shift) {
							mSelectionEnd = mSelectionStart;
						}
					} else if (code == KeyEvent.VK_END) {
						bounds = getCharacterBounds(mSelectionStart);
						mSelectionEnd = getTextPosition(Integer.MAX_VALUE, bounds.y + 2);
						if (!shift) {
							mSelectionStart = mSelectionEnd;
						}
					} else if (code == KeyEvent.VK_DELETE) {
						deleteForward();
					} else {
						return false;
					}
					break;
				case KeyEvent.KEY_TYPED:
					char ch = event.getKeyChar();

					if (ch != '\t' && ch != KeyEvent.VK_DELETE && (ch >= ' ' || ch == '\r' || ch == '\n' || ch == '\b')) {
						insert(ch);
					} else if (ch != KeyEvent.VK_DELETE) {
						return false;
					}
					break;
				default:
					return false;
			}

			event.consume();
		}
		return needRepaint;
	}

	/** @param active Whether this document is active. */
	public void setActive(boolean active) {
		mActive = active;
	}

	/**
	 * Select the specified text range.
	 * 
	 * @param startIndex The starting index, inclusive.
	 * @param endIndex The ending index, exclusive.
	 */
	public void setSelection(int startIndex, int endIndex) {
		int length = getLength();

		if (startIndex < 0) {
			startIndex = 0;
		}
		if (startIndex > length) {
			startIndex = length;
		}
		if (endIndex < startIndex) {
			endIndex = startIndex;
		}
		if (endIndex > length) {
			endIndex = length;
		}
		mSelectionStart = startIndex;
		mSelectionEnd = endIndex;
	}

	/**
	 * Replaces the contents of this document with the specified text.
	 * 
	 * @param text The text to use for replacement.
	 */
	public void setText(String text) {
		mParagraphs.clear();
		if (text != null) {
			int length = text.length() - 1;
			int beginAt = 0;

			for (int i = 0; i <= length; i++) {
				char ch = text.charAt(i);

				if (i == length || ch == '\n') {
					TKParagraph paragraph;

					if (i == length && ch != '\n') {
						i++;
					}
					if (beginAt == i) {
						paragraph = new TKParagraph();
					} else {
						paragraph = new TKParagraph(text.substring(beginAt, i));
					}
					mParagraphs.add(paragraph);
					beginAt = i + 1;
				}
			}
		}
		if (getParagraphCount() < 1) {
			mParagraphs.add(new TKParagraph());
		}
		notifyDocumentListeners();
	}
}
