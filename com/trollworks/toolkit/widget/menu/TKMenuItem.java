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

package com.trollworks.toolkit.widget.menu;

import com.trollworks.toolkit.io.TKImage;
import com.trollworks.toolkit.text.TKTextDrawing;
import com.trollworks.toolkit.utility.TKAlignment;
import com.trollworks.toolkit.utility.TKColor;
import com.trollworks.toolkit.utility.TKFont;
import com.trollworks.toolkit.utility.TKKeystroke;

import java.awt.Color;
import java.awt.Font;
import java.awt.Graphics2D;
import java.awt.font.FontRenderContext;
import java.awt.font.LineMetrics;
import java.awt.image.BufferedImage;

/** A standard menu item. */
public class TKMenuItem {
	/** The left and right margins of the menu item. */
	public static final int	H_MARGIN	= 2;
	/** The gap between various parts of the menu item. */
	public static final int	GAP			= 4;
	private String			mTitle;
	private BufferedImage	mIcon;
	private String			mFontKey;
	private int				mIndent;
	private boolean			mMarked;
	private boolean			mAmbiguousMark;
	private TKKeystroke		mKeyStroke;
	private TKBaseMenu		mSubMenu;
	private boolean			mEnabled;
	private String			mCommand;
	private boolean			mFullDisplay;
	private int				mCachedDividingPoint;
	private int				mCachedWidth;
	private int				mCachedDividingPointUsed;
	private int				mCachedHeight;
	private Object			mUserObject;
	private int				mHierWidth;
	private int				mHierHeight;
	private int				mMarkWidth;
	private int				mMarkHeight;

	private TKMenuItem() {
		// Prevent creation without parameters
	}

	/**
	 * Creates a new menu item.
	 * 
	 * @param title The title to use.
	 */
	public TKMenuItem(String title) {
		this(title, null, null, null);
	}

	/**
	 * Creates a new menu item.
	 * 
	 * @param title The title to use.
	 * @param command The command to use.
	 */
	public TKMenuItem(String title, String command) {
		this(title, null, null, command);
	}

	/**
	 * Creates a new menu item.
	 * 
	 * @param title The title to use.
	 * @param command The command to use.
	 * @param fontKey The font to use.
	 */
	public TKMenuItem(String title, String command, String fontKey) {
		this(title, null, null, command);
		setFont(fontKey);
	}

	/**
	 * Creates a new menu item.
	 * 
	 * @param title The title to use.
	 * @param icon The icon to use.
	 */
	public TKMenuItem(String title, BufferedImage icon) {
		this(title, icon, null, null);
	}

	/**
	 * Creates a new menu item.
	 * 
	 * @param title The title to use.
	 * @param icon The icon to use.
	 * @param command The command to use.
	 */
	public TKMenuItem(String title, BufferedImage icon, String command) {
		this(title, icon, null, command);
	}

	/**
	 * Creates a new menu item.
	 * 
	 * @param title The title to use.
	 * @param keyStroke The key stroke to use.
	 * @param command The command to use.
	 */
	public TKMenuItem(String title, TKKeystroke keyStroke, String command) {
		this(title, null, keyStroke, command);
	}

	/**
	 * Creates a new menu item.
	 * 
	 * @param title The title to use.
	 * @param icon The icon to use.
	 * @param keyStroke The key stroke to use.
	 * @param command The command to use.
	 */
	public TKMenuItem(String title, BufferedImage icon, TKKeystroke keyStroke, String command) {
		super();

		BufferedImage aIcon = TKImage.getHierarchicalMenuArrowIcon();
		mHierWidth = aIcon.getWidth();
		mHierHeight = aIcon.getHeight();

		aIcon = TKImage.getCheckMarkIcon();
		mMarkWidth = aIcon.getWidth();
		mMarkHeight = aIcon.getHeight();

		setTitle(title);
		setIcon(icon);
		setKeyStroke(keyStroke);
		setCommand(command);
		setFullDisplay(true);
		setEnabled(true);
	}

	/**
	 * Draws this menu item into the specified bounds.
	 * 
	 * @param g2d The graphics context to use for drawing.
	 * @param x The starting horizontal location of the area that can be drawn into.
	 * @param y The starting vertical location of the area that can be drawn into.
	 * @param width The width of the area that can be drawn into.
	 * @param height The height of the area that can be drawn into.
	 * @param color If not <code>null</code>, the color to use when drawing text, otherwise, use
	 *            the defaults.
	 */
	public void draw(Graphics2D g2d, int x, int y, int width, int height, Color color) {
		int right = x + width;
		int origX = x;
		LineMetrics metrics;
		int fy;
		String title;
		FontRenderContext frc;

		x += H_MARGIN + mIndent;

		if (isFullDisplay()) {
			if (mMarked) {
				if (mAmbiguousMark) {
					int tmp = y + height / 2 - 1;

					g2d.setColor(color == null ? Color.black : color);
					g2d.drawLine(x + 1, tmp, x + mMarkWidth - 1, tmp);
					tmp++;
					g2d.drawLine(x + 1, tmp, x + mMarkWidth - 1, tmp);
				} else {
					g2d.drawImage(TKImage.getCheckMarkIcon(), x + 2, y + (height - mMarkHeight) / 2, null);
				}
			}
			x += mMarkWidth + GAP;
		} else {
			x += H_MARGIN;
		}

		if (mIcon != null) {
			g2d.drawImage(mIcon, x, y + (height - mIcon.getHeight()) / 2, null);
			x += mIcon.getWidth() + GAP;
		}

		if (mTitle != null && mTitle.length() > 0) {
			Font font = getFont();

			g2d.setFont(font);
			frc = g2d.getFontRenderContext();
			metrics = font.getLineMetrics(mTitle, frc);
			title = TKTextDrawing.truncateIfNecessary(font, frc, mTitle, (width < mCachedDividingPointUsed ? width : mCachedDividingPointUsed) + origX - x, TKAlignment.RIGHT);
			fy = y + (int) metrics.getAscent() + (height - (int) (metrics.getAscent() + metrics.getDescent())) / 2;

			if (!mEnabled) {
				g2d.setColor(Color.white);
				g2d.drawString(title, x + 1, fy + 1);
			}

			g2d.setColor(color == null ? getTitleColor() : color);
			g2d.drawString(title, x, fy);
		}

		if (isFullDisplay()) {
			if (mKeyStroke != null) {
				Font font = TKFont.lookup(TKFont.MENU_KEY_FONT_KEY);
				int fx;

				g2d.setFont(font);
				frc = g2d.getFontRenderContext();
				title = mKeyStroke.toString();
				metrics = font.getLineMetrics(title, frc);
				fx = right - (GAP + mHierWidth + H_MARGIN + TKTextDrawing.getWidth(font, frc, title));
				fy = y + (int) metrics.getAscent() + (height - (int) (metrics.getAscent() + metrics.getDescent())) / 2;

				if (!mEnabled) {
					g2d.setColor(Color.white);
					g2d.drawString(title, fx + 1, fy + 1);
				}

				g2d.setColor(color == null ? getKeyStrokeColor() : color);
				g2d.drawString(title, fx, fy);
			}

			if (mSubMenu != null) {
				BufferedImage img = TKImage.getHierarchicalMenuArrowIcon();

				if (!mEnabled) {
					img = TKImage.createDisabledImage(img);
				}
				g2d.drawImage(img, right - (mHierWidth + H_MARGIN), y + (height - mHierHeight) / 2, null);
			}
		}
	}

	/** @return The command for this menu item. */
	public String getCommand() {
		return mCommand == null ? "" : mCommand; //$NON-NLS-1$
	}

	/** @return The font for this menu item. */
	public Font getFont() {
		return TKFont.lookup(mFontKey == null ? TKFont.MENU_FONT_KEY : mFontKey);
	}

	/** @return The height of this menu item. */
	public int getHeight() {
		if (mCachedHeight == 0) {
			int height;

			if (mIcon != null) {
				mCachedHeight = mIcon.getHeight();
			}

			if (mTitle != null && mTitle.length() > 0) {
				height = TKTextDrawing.getPreferredSize(getFont(), null, mTitle).height;
				if (mCachedHeight < height) {
					mCachedHeight = height;
				}
			}

			if (isFullDisplay()) {
				if (isFullDisplay() && mMarked) {
					height = mMarkHeight;
					if (mCachedHeight < height) {
						mCachedHeight = height;
					}
				}

				if (mKeyStroke != null) {
					height = TKTextDrawing.getPreferredSize(TKFont.lookup(TKFont.MENU_KEY_FONT_KEY), null, mKeyStroke.toString()).height;
					if (mCachedHeight < height) {
						mCachedHeight = height;
					}
				}

				if (mSubMenu != null) {
					height = mHierHeight;
					if (mCachedHeight < height) {
						mCachedHeight = height;
					}
				}
			}

			mCachedHeight += 9;
		}
		return mCachedHeight;
	}

	/** @return The icon for this item. */
	public BufferedImage getIcon() {
		return mIcon;
	}

	/** @return The amount of indention to use. */
	public int getIndent() {
		return mIndent;
	}

	/** @return The key stroke for this item. */
	public TKKeystroke getKeyStroke() {
		return mKeyStroke;
	}

	/** @return The key stroke color. */
	public Color getKeyStrokeColor() {
		return isEnabled() ? TKColor.MENU_KEYSTROKE : TKColor.MENU_DISABLED_KEYSTROKE;
	}

	/** @return The mark area width. */
	public int getMarkWidth() {
		return mMarkWidth;
	}

	/** @return The preferred dividing point (the point between the title and key stroke title). */
	public int getPreferredDividingPoint() {
		if (mCachedDividingPoint == 0) {
			if (isFullDisplay()) {
				mCachedDividingPoint = mMarkWidth + GAP;
			} else {
				mCachedDividingPoint = H_MARGIN;
			}

			mCachedDividingPoint += H_MARGIN + mIndent;

			if (mIcon != null) {
				mCachedDividingPoint += mIcon.getWidth() + GAP;
			}

			if (mTitle != null && mTitle.length() > 0) {
				mCachedDividingPoint += TKTextDrawing.getWidth(getFont(), null, mTitle) + GAP;
			}
		}
		return mCachedDividingPoint;
	}

	/** @return The sub-menu attached to this item. */
	public TKBaseMenu getSubMenu() {
		return mSubMenu;
	}

	/** @return The title for this menu item. */
	public String getTitle() {
		return mTitle;
	}

	/** @return The title color. */
	public Color getTitleColor() {
		return isEnabled() ? Color.black : Color.gray;
	}

	/** @return The user object. */
	public Object getUserObject() {
		return mUserObject;
	}

	/**
	 * @param dividingPoint The dividing point between the menu title and its command key.
	 * @return The width of this menu based on the specified dividing point.
	 */
	public int getWidth(int dividingPoint) {
		if (mCachedWidth == 0 || mCachedDividingPointUsed != dividingPoint) {
			mCachedDividingPointUsed = dividingPoint;
			if (isFullDisplay()) {
				mCachedWidth = dividingPoint + GAP;
				if (mKeyStroke != null) {
					mCachedWidth += TKTextDrawing.getWidth(TKFont.lookup(TKFont.MENU_KEY_FONT_KEY), null, mKeyStroke.toString()) + GAP;
				}
				mCachedWidth += mHierWidth + H_MARGIN;
			} else {
				mCachedWidth = dividingPoint + H_MARGIN + H_MARGIN;
			}
		}
		return mCachedWidth;
	}

	/** Resets any cached information retained by this menu item. */
	public void invalidate() {
		mCachedDividingPoint = 0;
		mCachedWidth = 0;
		mCachedDividingPointUsed = 0;
		mCachedHeight = 0;
	}

	/** @return <code>true</code> if this menu item is enabled. */
	public boolean isEnabled() {
		return mEnabled;
	}

	/** @return <code>true</code> if this menu item is in full display mode. */
	public boolean isFullDisplay() {
		return mFullDisplay;
	}

	/** @return <code>true</code> if this menu item has a check next to it. */
	public boolean isMarked() {
		return mMarked;
	}

	/** @return <code>true</code> if this menu item has a dash next to it. */
	public boolean isAmbiguouslyMarked() {
		return mMarked && mAmbiguousMark;
	}

	/**
	 * Sets the command for this menu item.
	 * 
	 * @param command The command to set.
	 */
	public void setCommand(String command) {
		mCommand = command;
	}

	/**
	 * Sets the enabled state of this menu item.
	 * 
	 * @param enabled Pass in <code>true</code> if this item should be enabled.
	 */
	public void setEnabled(boolean enabled) {
		mEnabled = enabled;
	}

	/**
	 * Sets the font for this menu item.
	 * 
	 * @param fontKey The font to set.
	 */
	public void setFont(String fontKey) {
		if (fontKey == null ? mFontKey != null : !fontKey.equals(mFontKey)) {
			mFontKey = fontKey;
			invalidate();
		}
	}

	/**
	 * @param fullDisplay Whether the adornments (check mark, key stroke, hierarchical menu
	 *            indicator) should be used for this menu item.
	 */
	public void setFullDisplay(boolean fullDisplay) {
		if (mFullDisplay != fullDisplay) {
			mFullDisplay = fullDisplay;
			invalidate();
		}
	}

	/**
	 * Sets the icon for this menu item.
	 * 
	 * @param icon The icon to set.
	 */
	public void setIcon(BufferedImage icon) {
		if (mIcon != icon) {
			mIcon = icon;
			invalidate();
		}
	}

	/**
	 * @param amount The amount of indentation to use.
	 */
	public void setIndent(int amount) {
		if (amount != mIndent) {
			mIndent = amount;
			invalidate();
		}
	}

	/** @param keyStroke The key stroke for this menu item. */
	public void setKeyStroke(TKKeystroke keyStroke) {
		if (mKeyStroke == null || !mKeyStroke.equals(keyStroke)) {
			mKeyStroke = keyStroke;
			invalidate();
		}
	}

	/** @param marked Whether this menu item has a check mark or not. */
	public void setMarked(boolean marked) {
		setMarked(marked, false);
	}

	/**
	 * Sets whether this menu item has a check mark or not.
	 * 
	 * @param marked Pass in <code>true</code> to mark it, <code>false</code> to unmark it.
	 * @param ambiguous Pass in <code>true</code> to use the ambiguous check mark state. Only
	 *            meaningful if the menu is marked.
	 */
	public void setMarked(boolean marked, boolean ambiguous) {
		mMarked = marked;
		mAmbiguousMark = ambiguous;
	}

	/** @param menu The sub-menu for this menu item. */
	public void setSubMenu(TKBaseMenu menu) {
		if (mSubMenu != menu) {
			mSubMenu = menu;
			invalidate();
		}
	}

	/** @param title The title for this menu item. */
	public void setTitle(String title) {
		if (mTitle != title && (mTitle == null || mTitle != null && !mTitle.equals(title))) {
			mTitle = title;
			invalidate();
		}
	}

	/** @param obj The user object. */
	public void setUserObject(Object obj) {
		mUserObject = obj;
	}
}
