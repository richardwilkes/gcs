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

import com.trollworks.toolkit.io.TKImage;
import com.trollworks.toolkit.io.TKPreferences;
import com.trollworks.toolkit.utility.TKAlignment;
import com.trollworks.toolkit.utility.TKFont;
import com.trollworks.toolkit.widget.TKLabel;
import com.trollworks.toolkit.widget.TKPanel;
import com.trollworks.toolkit.widget.TKTextField;
import com.trollworks.toolkit.widget.border.TKEmptyBorder;
import com.trollworks.toolkit.widget.button.TKButton;
import com.trollworks.toolkit.widget.button.TKCheckbox;
import com.trollworks.toolkit.widget.layout.TKCompassLayout;
import com.trollworks.toolkit.widget.layout.TKCompassPosition;

import java.awt.Component;
import java.awt.Dialog;
import java.awt.Dimension;
import java.awt.FlowLayout;
import java.awt.Frame;
import java.awt.Toolkit;
import java.awt.event.ActionEvent;
import java.awt.event.ActionListener;
import java.awt.image.BufferedImage;

/**
 * Provides a modal dialog with standard button options. Various static methods are provided to
 * create common dialog types.
 */
public class TKOptionDialog extends TKDialog implements ActionListener {
	/** A dialog with just an "OK" button. */
	public static final int		TYPE_OK				= 0;
	/** A dialog with both an "OK" and a "Cancel" button. */
	public static final int		TYPE_OK_CANCEL		= 1;
	/** A dialog with both a "Yes" and a "No" button. */
	public static final int		TYPE_YES_NO			= 2;
	/** A dialog with a "Yes", "No", and "Cancel" button. */
	public static final int		TYPE_YES_NO_CANCEL	= 3;
	/** The code returned when the "Yes" button is pressed. */
	public static final int		YES					= 1;
	/** The code returned when the "No" button is pressed. */
	public static final int		NO					= 3;
	/** The "OK" action command. */
	public static final String	CMD_OK				= "OptionDialog.OK";		//$NON-NLS-1$
	/** The "Cancel" action command. */
	public static final String	CMD_CANCEL			= "OptionDialog.Cancel";	//$NON-NLS-1$
	/** The "Yes" action command. */
	public static final String	CMD_YES				= "OptionDialog.Yes";		//$NON-NLS-1$
	/** The "No" action command. */
	public static final String	CMD_NO				= "OptionDialog.No";		//$NON-NLS-1$
	private static final String	MODULE				= "DontShowDialogAgain";	//$NON-NLS-1$
	private static final String	CMD_DONT_SHOW_AGAIN	= "DontShowAgain";			//$NON-NLS-1$
	private TKCheckbox			mCheckBox;
	private int					mType;
	private TKPanel				mButtonPanel;
	private TKButton			mOKButton;
	private TKButton			mNoButton;
	private String				mOKButtonTitle;
	private String				mCancelButtonTitle;
	private String				mNoButtonTitle;

	/**
	 * Creates a new dialog of the specified type.
	 * 
	 * @param title The title to use.
	 * @param type The type of dialog to create. One of the <code>TYPE_xxx</code> constants.
	 */
	public TKOptionDialog(String title, int type) {
		super(title, true);
		mType = type;
		setResizable(false);
	}

	/**
	 * Creates a new dialog of the specified type.
	 * 
	 * @param owner The owning window.
	 * @param title The title to use.
	 * @param type The type of dialog to create. One of the <code>TYPE_xxx</code> constants.
	 */
	public TKOptionDialog(Frame owner, String title, int type) {
		super(owner, title, true);
		mType = type;
		setResizable(false);
	}

	/**
	 * Creates a new dialog of the specified type.
	 * 
	 * @param owner The owning dialog.
	 * @param title The title to be used.
	 * @param type The type of dialog to create. One of the <code>TYPE_xxx</code> constants.
	 */
	public TKOptionDialog(Dialog owner, String title, int type) {
		super(owner, title, true);
		mType = type;
		setResizable(false);
	}

	public void actionPerformed(ActionEvent event) {
		String command = event.getActionCommand();

		if (CMD_OK.equals(command)) {
			setResult(OK);
		} else if (CMD_CANCEL.equals(command)) {
			setResult(CANCEL);
		} else if (CMD_YES.equals(command)) {
			setResult(YES);
		} else if (CMD_NO.equals(command)) {
			setResult(NO);
		}

		if (getResult() != NOT_SET) {
			attemptClose();
		}
	}

	private TKButton addButton(String command, String name) {
		TKButton button = new TKButton(name);

		button.setActionCommand(command);
		button.addActionListener(this);
		mButtonPanel.add(button);
		return button;
	}

	/** Resets all "Don't Show Again" dialogs so that they will show again. */
	public static void resetAllDialogs() {
		TKPreferences.getInstance().removePreferences(MODULE);
	}

	/**
	 * @return <code>true</code> if {@link #resetAllDialogs()} will do anything. Can be used to
	 *         enable/disable a "Reset All Dialogs" menu item.
	 */
	public static boolean isResetNeeded() {
		return TKPreferences.getInstance().hasPreferences(MODULE);
	}

	/**
	 * Provides a modal loop for servicing modal dialogs. This method returns once the dialog is no
	 * longer visible. Upon entry into this method, the dialog will be centered on the screen or its
	 * owning window and displayed on the screen.
	 * 
	 * @param icon The icon to use, if any.
	 * @param userContent The user content to display.
	 * @return The value set by calls to {@link #setResult(int)} during the run.
	 */
	public int doModal(BufferedImage icon, Component userContent) {
		return doModal(icon, userContent, null);
	}

	/**
	 * Provides a modal loop for servicing modal dialogs. This method returns once the dialog is no
	 * longer visible. Upon entry into this method, the dialog will be centered on the screen or its
	 * owning window and displayed on the screen.
	 * 
	 * @param icon The icon to use, if any.
	 * @param userContent The user content to display.
	 * @param dontShowTag The tag for tracking "don't show again" state, or <code>null</code> if
	 *            the dialog should always show.
	 * @return The value set by calls to {@link #setResult(int)} during the run.
	 */
	public int doModal(BufferedImage icon, Component userContent, String dontShowTag) {
		TKPanel content = getContent();
		TKPanel contentWrapper = new TKPanel(new TKCompassLayout());
		int returnValue;

		content.setBorder(new TKEmptyBorder(10, 10, 5, 10));
		if (icon != null) {
			TKLabel iconLabel = new TKLabel(icon);

			iconLabel.setVerticalAlignment(TKAlignment.TOP);
			content.add(iconLabel, TKCompassPosition.WEST);
			contentWrapper.setBorder(new TKEmptyBorder(0, 5, 0, 0));
		}
		contentWrapper.add(userContent, TKCompassPosition.CENTER);
		content.add(contentWrapper, TKCompassPosition.CENTER);

		if (dontShowTag != null) {
			TKPanel southWestPanel = new TKPanel(new TKCompassLayout());
			TKPanel southPanel = new TKPanel(new TKCompassLayout());

			mCheckBox = new TKCheckbox(Msgs.DONT_SHOW_AGAIN);
			mCheckBox.setActionCommand(CMD_DONT_SHOW_AGAIN);
			mCheckBox.addActionListener(this);

			southWestPanel.add(new TKPanel(), TKCompassPosition.CENTER);
			southWestPanel.add(mCheckBox, TKCompassPosition.SOUTH);
			southPanel.add(southWestPanel, TKCompassPosition.WEST);
			southPanel.add(getButtonPanel(), TKCompassPosition.EAST);
			content.add(southPanel, TKCompassPosition.SOUTH);
		} else {
			content.add(getButtonPanel(), TKCompassPosition.SOUTH);
		}

		adjustButtonSizes();

		returnValue = doModal();
		if (dontShowTag != null && mCheckBox.isChecked()) {
			TKPreferences.getInstance().setValue(MODULE, dontShowTag, true);
		}

		return returnValue;
	}

	/** @return The component that contains the standard buttons for this dialog. */
	public TKPanel getButtonPanel() {
		if (mButtonPanel == null) {
			TKButton button;

			mButtonPanel = new TKPanel(new FlowLayout(FlowLayout.RIGHT));
			mButtonPanel.setBorder(new TKEmptyBorder(10, 0, 0, 0));

			switch (mType) {
				case TYPE_OK:
					mOKButton = addButton(CMD_OK, Msgs.OK_TITLE);
					break;
				case TYPE_OK_CANCEL:
					setCancelButton(addButton(CMD_CANCEL, Msgs.CANCEL_TITLE));
					mOKButton = addButton(CMD_OK, Msgs.OK_TITLE);
					break;
				case TYPE_YES_NO:
					setCancelButton(addButton(CMD_NO, Msgs.NO_TITLE));
					mOKButton = addButton(CMD_YES, Msgs.YES_TITLE);
					break;
				case TYPE_YES_NO_CANCEL:
					setCancelButton(addButton(CMD_CANCEL, Msgs.CANCEL_TITLE));
					mNoButton = addButton(CMD_NO, Msgs.NO_TITLE);
					mOKButton = addButton(CMD_YES, Msgs.YES_TITLE);
					break;
				default:
					mOKButton = addButton(CMD_OK, Msgs.OK_TITLE);
					break;
			}

			if (mOKButton != null && mOKButtonTitle != null) {
				mOKButton.setText(mOKButtonTitle);
			}
			button = getCancelButton();
			if (button != null && mCancelButtonTitle != null) {
				button.setText(mCancelButtonTitle);
			}
			if (mNoButton != null && mNoButtonTitle != null) {
				mNoButton.setText(mNoButtonTitle);
			}
			setDefaultButton(mOKButton);
		}
		return mButtonPanel;
	}

	private void adjustButtonSizes() {
		if (mButtonPanel != null) {
			int count = mButtonPanel.getComponentCount();
			Dimension size = new Dimension();
			int i;

			for (i = 0; i < count; i++) {
				Component comp = mButtonPanel.getComponent(i);

				if (comp instanceof TKButton) {
					Dimension cSize = comp.getPreferredSize();

					if (size.width < cSize.width) {
						size.width = cSize.width;
					}
					if (size.height < cSize.height) {
						size.height = cSize.height;
					}
				}
			}
			for (i = 0; i < count; i++) {
				Component comp = mButtonPanel.getComponent(i);

				if (comp instanceof TKButton) {
					((TKButton) comp).setMinimumSize(size);
					((TKButton) comp).setMaximumSize(size);
					((TKButton) comp).setPreferredSize(size);
				}
			}
		}
	}

	/** @return The Cancel button. */
	@Override public TKButton getCancelButton() {
		getButtonPanel();
		return super.getCancelButton();
	}

	/** @return The No button. */
	public TKButton getNoButton() {
		getButtonPanel();
		return mType == TYPE_YES_NO ? getCancelButton() : mNoButton;
	}

	/** @return The OK button. */
	public TKButton getOKButton() {
		getButtonPanel();
		return mOKButton;
	}

	/** @return The Yes button. */
	public TKButton getYesButton() {
		getButtonPanel();
		return mOKButton;
	}

	/**
	 * Sets the title of the Cancel button.
	 * 
	 * @param title The title to use. Pass in <code>null</code> to use the default.
	 */
	public void setCancelButtonTitle(String title) {
		TKButton button;

		mCancelButtonTitle = title;
		button = getCancelButton();
		if (button != null) {
			button.setText(title != null ? title : Msgs.CANCEL_TITLE);
		}
	}

	/**
	 * Sets the title of the No button.
	 * 
	 * @param title The title to use. Pass in <code>null</code> to use the default.
	 */
	public void setNoButtonTitle(String title) {
		TKButton button;

		mNoButtonTitle = title;
		button = getNoButton();
		if (button != null) {
			button.setText(title != null ? title : Msgs.NO_TITLE);
		}
	}

	/**
	 * Sets the title of the OK button.
	 * 
	 * @param title The title to use. Pass in <code>null</code> to use the default.
	 */
	public void setOKButtonTitle(String title) {
		TKButton button;

		mOKButtonTitle = title;
		button = getOKButton();
		if (button != null) {
			button.setText(title != null ? title : Msgs.OK_TITLE);
		}
	}

	/**
	 * Sets the title of the Yes button.
	 * 
	 * @param title The title to use. Pass in <code>null</code> to use the default.
	 */
	public void setYesButtonTitle(String title) {
		TKButton button;

		mOKButtonTitle = title;
		button = getYesButton();
		if (button != null) {
			button.setText(title != null ? title : Msgs.YES_TITLE);
		}
	}

	/**
	 * Displays a standard confirmation dialog with the specified message and waits for the user to
	 * dismiss it.
	 * 
	 * @param message The message to display.
	 * @return The result.
	 */
	public static int confirm(String message) {
		return confirm(null, message, TYPE_YES_NO_CANCEL, null);
	}

	/**
	 * Displays a standard confirmation dialog with the specified message and waits for the user to
	 * dismiss it.
	 * 
	 * @param message The message to display.
	 * @param dontShowTag The tag for tracking "don't show again" state, or <code>null</code> if
	 *            the dialog should always show.
	 * @return The result.
	 */
	public static int confirm(String message, String dontShowTag) {
		return confirm(null, message, TYPE_YES_NO_CANCEL, dontShowTag);
	}

	/**
	 * Displays a confirmation dialog with the specified message and waits for the user to respond.
	 * 
	 * @param message The message to display.
	 * @param type The type of dialog to create. Pass in one of the <code>TYPE_xxx</code>
	 *            constants.
	 * @return The result.
	 */
	public static int confirm(String message, int type) {
		return confirm(null, message, type, null);
	}

	/**
	 * Displays a confirmation dialog with the specified message and waits for the user to respond.
	 * 
	 * @param message The message to display.
	 * @param type The type of dialog to create. Pass in one of the <code>TYPE_xxx</code>
	 *            constants.
	 * @param dontShowTag The tag for tracking "don't show again" state, or <code>null</code> if
	 *            the dialog should always show.
	 * @return The result.
	 */
	public static int confirm(String message, int type, String dontShowTag) {
		return confirm(null, message, type, dontShowTag);
	}

	/**
	 * Displays a standard confirmation dialog with the specified message and waits for the user to
	 * dismiss it.
	 * 
	 * @param owner The owning window/dialog.
	 * @param message The message to display.
	 * @return The result.
	 */
	public static int confirm(TKBaseWindow owner, String message) {
		return confirm(owner, message, TYPE_YES_NO_CANCEL, null);
	}

	/**
	 * Displays a confirmation dialog with the specified message and waits for the user to respond.
	 * 
	 * @param owner The owning window/dialog.
	 * @param message The message to display.
	 * @param type The type of dialog to create. Pass in one of the <code>TYPE_xxx</code>
	 *            constants.
	 * @return The result.
	 */
	public static int confirm(TKBaseWindow owner, String message, int type) {
		return confirm(owner, message, type, null);
	}

	/**
	 * Displays a standard confirmation dialog with the specified message and waits for the user to
	 * dismiss it.
	 * 
	 * @param owner The owning window/dialog.
	 * @param message The message to display.
	 * @param dontShowTag The tag for tracking "don't show again" state, or <code>null</code> if
	 *            the dialog should always show.
	 * @return The result.
	 */
	public static int confirm(TKBaseWindow owner, String message, String dontShowTag) {
		return confirm(owner, message, TYPE_YES_NO_CANCEL, dontShowTag);
	}

	/**
	 * Displays a confirmation dialog with the specified message and waits for the user to respond.
	 * 
	 * @param owner The owning window/dialog.
	 * @param message The message to display.
	 * @param type The type of dialog to create. Pass in one of the <code>TYPE_xxx</code>
	 *            constants.
	 * @param dontShowTag The tag for tracking "don't show again" state, or <code>null</code> if
	 *            the dialog should always show.
	 * @return The result.
	 */
	public static int confirm(TKBaseWindow owner, String message, int type, String dontShowTag) {
		return confirm(owner, message, type, dontShowTag, null);
	}

	/**
	 * Displays a confirmation dialog with the specified message and waits for the user to respond.
	 * 
	 * @param owner The owning window/dialog.
	 * @param message The message to display.
	 * @param type The type of dialog to create. Pass in one of the <code>TYPE_xxx</code>
	 *            constants.
	 * @param dontShowTag The tag for tracking "don't show again" state, or <code>null</code> if
	 *            the dialog should always show.
	 * @param buttonTitles An array containing the titles for the standard buttons. The title index
	 *            used will be equivalent to the return value of the dialog (if that button had been
	 *            pressed) minus 1. Pass in <code>null</code> to us the standards.
	 * @return The result.
	 */
	public static int confirm(TKBaseWindow owner, String message, int type, String dontShowTag, String[] buttonTitles) {
		return modal(owner, TKImage.getMessageIcon(), Msgs.CONFIRMATION_TITLE, type, message, dontShowTag, buttonTitles);
	}

	/**
	 * Displays a standard error dialog with the specified message and waits for the user to dismiss
	 * it.
	 * 
	 * @param message The message to display.
	 */
	public static void error(String message) {
		error(null, message, null);
	}

	/**
	 * Displays a standard error dialog with the specified message and waits for the user to dismiss
	 * it.
	 * 
	 * @param message The message to display.
	 * @param dontShowTag The tag for tracking "don't show again" state, or <code>null</code> if
	 *            the dialog should always show.
	 */
	public static void error(String message, String dontShowTag) {
		error(null, message, dontShowTag);
	}

	/**
	 * Displays a standard error dialog with the specified message and waits for the user to dismiss
	 * it.
	 * 
	 * @param owner The owning window/dialog.
	 * @param message The message to display.
	 */
	public static void error(TKBaseWindow owner, String message) {
		error(owner, message, null);
	}

	/**
	 * Displays a standard error dialog with the specified message and waits for the user to dismiss
	 * it.
	 * 
	 * @param owner The owning window/dialog.
	 * @param message The message to display.
	 * @param dontShowTag The tag for tracking "don't show again" state, or <code>null</code> if
	 *            the dialog should always show.
	 */
	public static void error(TKBaseWindow owner, String message, String dontShowTag) {
		error(owner, message, dontShowTag, null);
	}

	/**
	 * Displays a standard error dialog with the specified message and waits for the user to dismiss
	 * it.
	 * 
	 * @param owner The owning window/dialog.
	 * @param message The message to display.
	 * @param dontShowTag The tag for tracking "don't show again" state, or <code>null</code> if
	 *            the dialog should always show.
	 * @param okTitle The OK button title.
	 */
	public static void error(TKBaseWindow owner, String message, String dontShowTag, String okTitle) {
		Toolkit.getDefaultToolkit().beep();
		modal(owner, TKImage.getErrorIcon(), Msgs.ERROR_TITLE, TYPE_OK, message, dontShowTag, okTitle != null ? new String[] { okTitle, "", "" } : null); //$NON-NLS-1$ //$NON-NLS-2$
	}

	/**
	 * Displays a standard information dialog with the specified message and waits for the user to
	 * dismiss it.
	 * 
	 * @param message The message to display.
	 */
	public static void message(String message) {
		message(null, message, null);
	}

	/**
	 * Displays a standard information dialog with the specified message and waits for the user to
	 * dismiss it.
	 * 
	 * @param message The message to display.
	 * @param dontShowTag The tag for tracking "don't show again" state, or <code>null</code> if
	 *            the dialog should always show.
	 */
	public static void message(String message, String dontShowTag) {
		message(null, message, dontShowTag);
	}

	/**
	 * Displays a standard information dialog with the specified message and waits for the user to
	 * dismiss it.
	 * 
	 * @param owner The owning window/dialog.
	 * @param message The message to display.
	 */
	public static void message(TKBaseWindow owner, String message) {
		message(owner, message, null);
	}

	/**
	 * Displays a standard information dialog with the specified message and waits for the user to
	 * dismiss it.
	 * 
	 * @param owner The owning window/dialog.
	 * @param message The message to display.
	 * @param dontShowTag The tag for tracking "don't show again" state, or <code>null</code> if
	 *            the dialog should always show.
	 */
	public static void message(TKBaseWindow owner, String message, String dontShowTag) {
		message(owner, message, dontShowTag, null);
	}

	/**
	 * Displays a standard information dialog with the specified message and waits for the user to
	 * dismiss it.
	 * 
	 * @param owner The owning window/dialog.
	 * @param message The message to display.
	 * @param dontShowTag The tag for tracking "don't show again" state, or <code>null</code> if
	 *            the dialog should always show.
	 * @param buttonTitles An array containing the titles for the standard buttons. The title index
	 *            used will be equivalent to the return value of the dialog (if that button had been
	 *            pressed) minus 1. Pass in <code>null</code> to us the standards.
	 */
	public static void message(TKBaseWindow owner, String message, String dontShowTag, String[] buttonTitles) {
		modal(owner, TKImage.getMessageIcon(), Msgs.MESSAGE_TITLE, TYPE_OK, message, dontShowTag, buttonTitles);
	}

	/**
	 * Displays a standard modal dialog.
	 * 
	 * @param icon The icon to display, or <code>null</code>.
	 * @param title The title for the dialog.
	 * @param type The type of dialog to create. One of the <code>TYPE_xxx</code> constants.
	 * @param data If this parameter is a {@link TKPanel}, it will be inserted into the dialog,
	 *            otherwise, it will be made into a string and displayed in a {@link TKLabel}.
	 * @return The result code of the dialog.
	 */
	public static int modal(BufferedImage icon, String title, int type, Object data) {
		return modal(null, icon, title, type, data);
	}

	/**
	 * Displays a standard modal dialog.
	 * 
	 * @param owner The owning window/dialog.
	 * @param icon The icon to display, or <code>null</code>.
	 * @param title The title for the dialog.
	 * @param type The type of dialog to create. One of the <code>TYPE_xxx</code> constants.
	 * @param data If this parameter is a {@link TKPanel}, it will be inserted into the
	 *            dialog, otherwise, it will be made into a string and displayed in a
	 *            {@link TKLabel}.
	 * @return The result code of the dialog.
	 */
	public static int modal(TKBaseWindow owner, BufferedImage icon, String title, int type, Object data) {
		return modal(owner, icon, title, type, data, null);
	}

	/**
	 * Displays a standard modal dialog.
	 * 
	 * @param owner The owning window/dialog.
	 * @param icon The icon to display, or <code>null</code>.
	 * @param title The title for the dialog.
	 * @param type The type of dialog to create. One of the <code>TYPE_xxx</code> constants.
	 * @param data If this parameter is a {@link TKPanel}, it will be inserted into the
	 *            dialog, otherwise, it will be made into a string and displayed in a
	 *            {@link TKLabel}.
	 * @param dontShowTag The tag for tracking "don't show again" state, or <code>null</code> if
	 *            the dialog should always show.
	 * @return The result code of the dialog.
	 */
	public static int modal(TKBaseWindow owner, BufferedImage icon, String title, int type, Object data, String dontShowTag) {
		return modal(owner, icon, title, type, data, dontShowTag, null);
	}

	/**
	 * Displays a standard modal dialog.
	 * 
	 * @param owner The owning window/dialog.
	 * @param icon The icon to display, or <code>null</code>.
	 * @param title The title for the dialog.
	 * @param type The type of dialog to create. One of the <code>TYPE_xxx</code> constants.
	 * @param data If this parameter is a {@link TKPanel}, it will be inserted into the
	 *            dialog, otherwise, it will be made into a string and displayed in a
	 *            {@link TKLabel}.
	 * @param dontShowTag The tag for tracking "don't show again" state, or <code>null</code> if
	 *            the dialog should always show.
	 * @param buttonTitles An array containing the titles for the standard buttons. The title index
	 *            used will be equivalent to the return value of the dialog (if that button had been
	 *            pressed) minus 1. Pass in <code>null</code> to us the standards.
	 * @return The result code of the dialog.
	 */
	public static int modal(TKBaseWindow owner, BufferedImage icon, String title, int type, Object data, String dontShowTag, String[] buttonTitles) {
		if (dontShowTag == null || !TKPreferences.getInstance().getBooleanValue(MODULE, dontShowTag)) {
			TKOptionDialog dialog;
			TKPanel panel;

			if (owner == null) {
				dialog = new TKOptionDialog(title, type);
			} else if (owner instanceof TKWindow) {
				dialog = new TKOptionDialog((TKWindow) owner, title, type);
			} else {
				dialog = new TKOptionDialog((TKDialog) owner, title, type);
			}

			if (buttonTitles != null) {
				dialog.setOKButtonTitle(buttonTitles[OK - 1]);
				dialog.setCancelButtonTitle(buttonTitles[CANCEL - 1]);
				dialog.setNoButtonTitle(buttonTitles[NO - 1]);
			}

			if (data instanceof TKPanel) {
				panel = (TKPanel) data;
			} else {
				panel = new TKLabel(data.toString(), null, TKAlignment.LEFT, true, TKFont.CONTROL_FONT_KEY);
			}

			return dialog.doModal(icon, panel, dontShowTag);
		}
		return OK;
	}

	/**
	 * Displays a standard input dialog with the specified message and waits for the user to
	 * respond.
	 * 
	 * @param message The message to display.
	 * @return The response, or <code>null</code>.
	 */
	public static String response(String message) {
		return response((TKBaseWindow) null, message);
	}

	/**
	 * Displays a standard input dialog with the specified message and waits for the user to
	 * respond.
	 * 
	 * @param message The message to display.
	 * @param defaultAnswer The response to start with.
	 * @return The response, or <code>null</code>.
	 */
	public static String response(String message, String defaultAnswer) {
		return response(null, message, defaultAnswer);
	}

	/**
	 * Displays a standard input dialog with the specified message and waits for the user to
	 * respond.
	 * 
	 * @param message The message to display.
	 * @param defaultAnswer The response to start with.
	 * @param size The size to start with.
	 * @return The response, or <code>null</code>.
	 */
	public static String response(String message, String defaultAnswer, int size) {
		return response(null, message, defaultAnswer, size);
	}

	/**
	 * Displays a standard input dialog with the specified message and waits for the user to
	 * respond.
	 * 
	 * @param owner The owning window/dialog.
	 * @param message The message to display.
	 * @return The response, or <code>null</code>.
	 */
	public static String response(TKBaseWindow owner, String message) {
		return response(owner, message, null, 100);
	}

	/**
	 * Displays a standard input dialog with the specified message and waits for the user to
	 * respond.
	 * 
	 * @param owner The owning window/dialog.
	 * @param message The message to display.
	 * @param defaultAnswer The response to start with.
	 * @return The response, or <code>null</code>.
	 */
	public static String response(TKBaseWindow owner, String message, String defaultAnswer) {
		return response(owner, message, defaultAnswer, 100);
	}

	/**
	 * Displays a standard input dialog with the specified message and waits for the user to
	 * respond.
	 * 
	 * @param owner The owning window/dialog.
	 * @param message The message to display.
	 * @param defaultAnswer The response to start with.
	 * @param size The size to start with.
	 * @return The response, or <code>null</code>.
	 */
	public static String response(TKBaseWindow owner, String message, String defaultAnswer, int size) {
		return response(owner, message, defaultAnswer, size, null);
	}

	/**
	 * Displays a standard input dialog with the specified message and waits for the user to
	 * respond.
	 * 
	 * @param owner The owning window/dialog.
	 * @param message The message to display.
	 * @param defaultAnswer The response to start with.
	 * @param size The size to start with.
	 * @param buttonTitles An array containing the titles for the standard buttons. The title index
	 *            used will be equivalent to the return value of the dialog (if that button had been
	 *            pressed) minus 1. Pass in <code>null</code> to us the standards.
	 * @return The response, or <code>null</code>.
	 */
	public static String response(TKBaseWindow owner, String message, String defaultAnswer, int size, String[] buttonTitles) {
		TKPanel panel = new TKPanel(new FlowLayout());
		TKTextField field;
		int result;

		if (defaultAnswer == null) {
			defaultAnswer = ""; //$NON-NLS-1$
		}

		if (size < 100) {
			size = 100;
		}

		field = new TKTextField(defaultAnswer, size);
		field.selectAll();

		panel.add(new TKLabel(message, null, TKAlignment.LEFT, true, TKFont.CONTROL_FONT_KEY));
		panel.add(field);

		result = modal(owner, TKImage.getMessageIcon(), Msgs.RESPONSE_TITLE, TYPE_OK_CANCEL, panel, null, buttonTitles);
		return result == OK ? field.getText() : null;
	}

	/**
	 * Displays a standard warning dialog with the specified message and waits for the user to
	 * dismiss it.
	 * 
	 * @param message The message to display.
	 */
	public static void warn(String message) {
		warn(null, message, null);
	}

	/**
	 * Displays a standard warning dialog with the specified message and waits for the user to
	 * dismiss it.
	 * 
	 * @param message The message to display.
	 * @param dontShowTag The tag for tracking "don't show again" state, or <code>null</code> if
	 *            the dialog should always show.
	 */
	public static void warn(String message, String dontShowTag) {
		warn(null, message, dontShowTag);
	}

	/**
	 * Displays a standard warning dialog with the specified message and waits for the user to
	 * dismiss it.
	 * 
	 * @param owner The owning window/dialog.
	 * @param message The message to display.
	 */
	public static void warn(TKBaseWindow owner, String message) {
		warn(owner, message, null);
	}

	/**
	 * Displays a standard warning dialog with the specified message and waits for the user to
	 * dismiss it.
	 * 
	 * @param owner The owning window/dialog.
	 * @param message The message to display.
	 * @param dontShowTag The tag for tracking "don't show again" state, or <code>null</code> if
	 *            the dialog should always show.
	 */
	public static void warn(TKBaseWindow owner, String message, String dontShowTag) {
		warn(owner, message, TYPE_OK, dontShowTag);
	}

	/**
	 * Displays a standard warning dialog with the specified message and waits for the user to
	 * dismiss it.
	 * 
	 * @param owner The owning window/dialog.
	 * @param message The message to display.
	 * @param type The type of dialog to create. One of the <code>TYPE_xxx</code> constants.
	 * @param dontShowTag The tag for tracking "don't show again" state, or <code>null</code> if
	 *            the dialog should always show.
	 * @return The result.
	 */
	public static int warn(TKBaseWindow owner, String message, int type, String dontShowTag) {
		return warn(owner, message, type, dontShowTag, null);
	}

	/**
	 * Displays a standard warning dialog with the specified message and waits for the user to
	 * dismiss it.
	 * 
	 * @param owner The owning window/dialog.
	 * @param message The message to display.
	 * @param type The type of dialog to create. One of the <code>TYPE_xxx</code> constants.
	 * @param dontShowTag The tag for tracking "don't show again" state, or <code>null</code> if
	 *            the dialog should always show.
	 * @param buttonTitles An array containing the titles for the standard buttons. The title index
	 *            used will be equivalent to the return value of the dialog (if that button had been
	 *            pressed) minus 1. Pass in <code>null</code> to us the standards.
	 * @return The result.
	 */
	public static int warn(TKBaseWindow owner, String message, int type, String dontShowTag, String[] buttonTitles) {
		return modal(owner, TKImage.getWarningIcon(), Msgs.WARNING_TITLE, type, message, dontShowTag, buttonTitles);
	}
}
