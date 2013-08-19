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

package com.trollworks.toolkit.utility;

import java.awt.Event;
import java.awt.event.InputEvent;
import java.awt.event.KeyEvent;

/**
 * Represents a logical key stroke being typed on the keyboard, defined as a single key plus zero or
 * more modifier keys (shift, ctrl, alt, meta) or just the modifier keys themselves.
 */
public class TKKeystroke {
	private int	mKeyCode;
	private int	mModifiers;

	/**
	 * Create a new key stroke with the specified key code and the "command" modifier for the
	 * platform.
	 * 
	 * @param keyCode The key code.
	 */
	public TKKeystroke(int keyCode) {
		this(keyCode, getCommandMask());
	}

	/**
	 * Create a new key stroke with the specified key code and modifiers.
	 * 
	 * @param keyCode The key code.
	 * @param modifiers The key modifiers.
	 */
	public TKKeystroke(int keyCode, int modifiers) {
		mKeyCode = keyCode;
		mModifiers = modifiers;
	}

	/**
	 * Create a new key stroke based on the specified {@link KeyEvent}.
	 * 
	 * @param event The key event.
	 */
	public TKKeystroke(KeyEvent event) {
		mKeyCode = event.getKeyCode();
		mModifiers = event.getModifiers();
	}

	@Override public final boolean equals(Object obj) {
		if (obj == this) {
			return true;
		}
		if (obj instanceof TKKeystroke) {
			TKKeystroke other = (TKKeystroke) obj;
			int modifiers = mModifiers;
			int otherMods = other.mModifiers;

			return mKeyCode == other.mKeyCode && modifiers == otherMods;
		}
		return false;
	}

	/** @return The key code for this key stroke. */
	public final int getKeyCode() {
		return mKeyCode;
	}

	/** @return The modifiers for this key stroke. */
	public final int getModifiers() {
		return mModifiers;
	}

	@Override public final int hashCode() {
		return mModifiers << 17 | mKeyCode;
	}

	@Override public final String toString() {
		StringBuilder buffer = new StringBuilder();

		if ((mModifiers & InputEvent.META_MASK) != 0) {
			buffer.append(TKPlatform.isMacintosh() ? Msgs.COMMAND : Msgs.META);
		}
		if ((mModifiers & InputEvent.CTRL_MASK) != 0) {
			if (buffer.length() > 0) {
				buffer.append('+');
			}
			buffer.append(Msgs.CONTROL);
		}
		if ((mModifiers & InputEvent.ALT_MASK) != 0) {
			if (buffer.length() > 0) {
				buffer.append('+');
			}
			buffer.append(Msgs.ALT);
		}
		if ((mModifiers & InputEvent.SHIFT_MASK) != 0) {
			if (buffer.length() > 0) {
				buffer.append('+');
			}
			buffer.append(Msgs.SHIFT);
		}

		if (mKeyCode != 0) {
			if (buffer.length() > 0) {
				buffer.append('+');
			}
			switch (mKeyCode) {
				case '[':
					buffer.append('[');
					break;
				case ']':
					buffer.append(']');
					break;
				case ',':
					buffer.append(',');
					break;
				case '=':
					buffer.append('=');
					break;
				default:
					buffer.append(KeyEvent.getKeyText(mKeyCode));
					break;
			}
		}
		return buffer.toString();
	}

	/**
	 * @param event The input event.
	 * @return <code>true</code> if the platform's notion of a command key is down in the event.
	 */
	public static final boolean isCommandKeyDown(InputEvent event) {
		return TKPlatform.isMacintosh() ? event.isMetaDown() : event.isControlDown();
	}

	/** @return The appropriate event mask for the command key for this platform. */
	public static final int getCommandMask() {
		return TKPlatform.isMacintosh() ? Event.META_MASK : Event.CTRL_MASK;
	}

	/** @return The appropriate name fo the command key for this platform. */
	public static final String getCommandName() {
		return TKPlatform.isMacintosh() ? Msgs.COMMAND : Msgs.CONTROL;
	}
}
