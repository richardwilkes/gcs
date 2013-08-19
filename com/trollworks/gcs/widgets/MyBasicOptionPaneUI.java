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

package com.trollworks.gcs.widgets;

import java.awt.Dimension;
import java.awt.Graphics;
import java.awt.LayoutManager;

import javax.accessibility.Accessible;
import javax.swing.JComponent;
import javax.swing.JOptionPane;
import javax.swing.plaf.OptionPaneUI;
import javax.swing.plaf.basic.BasicOptionPaneUI;

/** A wrapper around the {@link BasicOptionPaneUI} that respects component minimum sizes. */
public class MyBasicOptionPaneUI extends BasicOptionPaneUI {
	private OptionPaneUI	mOriginal;

	/**
	 * Creates a new {@link MyBasicOptionPaneUI}.
	 * 
	 * @param original The original {@link OptionPaneUI}.
	 */
	public MyBasicOptionPaneUI(OptionPaneUI original) {
		mOriginal = original;
	}

	@Override public Dimension getMinimumSize(JComponent c) {
		if ((JOptionPane) c == optionPane) {
			Dimension ourMin = getMinimumOptionPaneSize();
			LayoutManager lm = c.getLayout();
			if (lm != null) {
				Dimension lmSize = lm.minimumLayoutSize(c);
				if (ourMin != null) {
					return new Dimension(Math.max(lmSize.width, ourMin.width), Math.max(lmSize.height, ourMin.height));
				}
				return lmSize;
			}
			return ourMin;
		}
		return null;
	}

	@Override public boolean containsCustomComponents(JOptionPane op) {
		return mOriginal.containsCustomComponents(op);
	}

	@Override public void selectInitialValue(JOptionPane op) {
		mOriginal.selectInitialValue(op);
	}

	@Override public boolean contains(JComponent c, int x, int y) {
		return mOriginal.contains(c, x, y);
	}

	@Override public Accessible getAccessibleChild(JComponent c, int i) {
		return mOriginal.getAccessibleChild(c, i);
	}

	@Override public int getAccessibleChildrenCount(JComponent c) {
		return mOriginal.getAccessibleChildrenCount(c);
	}

	@Override public Dimension getMaximumSize(JComponent c) {
		return mOriginal.getMaximumSize(c);
	}

	@Override public Dimension getPreferredSize(JComponent c) {
		return mOriginal.getPreferredSize(c);
	}

	@Override public void installUI(JComponent c) {
		mOriginal.installUI(c);
	}

	@Override public void paint(Graphics g, JComponent c) {
		mOriginal.paint(g, c);
	}

	@Override public void uninstallUI(JComponent c) {
		mOriginal.uninstallUI(c);
	}

	@Override public void update(Graphics g, JComponent c) {
		mOriginal.update(g, c);
	}
}
