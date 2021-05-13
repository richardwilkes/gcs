/*
 * Copyright ©1998-2021 by Richard A. Wilkes. All rights reserved.
 *
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, version 2.0. If a copy of the MPL was not distributed with
 * this file, You can obtain one at http://mozilla.org/MPL/2.0/.
 *
 * This Source Code Form is "Incompatible With Secondary Licenses", as
 * defined by the Mozilla Public License, version 2.0.
 */

package com.trollworks.gcs.ui.widget;

import com.trollworks.gcs.utility.text.Text;

import javax.swing.JTextArea;
import javax.swing.UIManager;
import javax.swing.event.DocumentListener;

public class MultiLineTextField extends JTextArea {
    public MultiLineTextField(String text, String tooltip, DocumentListener listener) {
        super(text);
        setToolTipText(Text.wrapPlainTextForToolTip(tooltip));
        setBorder(UIManager.getBorder("FormattedTextField.border"));
        setTabSize(2);
        setLineWrap(true);
        setWrapStyleWord(true);
        getDocument().addDocumentListener(listener);
    }
}
