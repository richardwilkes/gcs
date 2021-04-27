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

package com.trollworks.gcs.attribute;

import com.trollworks.gcs.ui.layout.PrecisionLayout;
import com.trollworks.gcs.ui.layout.PrecisionLayoutData;
import com.trollworks.gcs.ui.widget.FontAwesomeButton;
import com.trollworks.gcs.ui.widget.Label;
import com.trollworks.gcs.utility.I18n;
import com.trollworks.gcs.utility.json.JsonMap;

import java.util.Map;
import javax.swing.JPanel;
import javax.swing.JScrollPane;
import javax.swing.ScrollPaneConstants;

public class AttributeEditor extends JPanel {
    public static final int BUTTON_SIZE = 14;
    private AttributeListPanel mListPanel;

    public AttributeEditor(Map<String, AttributeDef> attributes, Runnable adjustCallback) {
        super(new PrecisionLayout().setColumns(2).setMargins(0));
        setOpaque(false);
        add(new Label(I18n.Text("Attributes")));
        add(new FontAwesomeButton("\uf055", BUTTON_SIZE, I18n.Text("Add Attribute"), () -> mListPanel.addAttribute()));
        mListPanel = new AttributeListPanel(attributes, adjustCallback);
        JScrollPane scroller  = new JScrollPane(mListPanel, ScrollPaneConstants.VERTICAL_SCROLLBAR_AS_NEEDED, ScrollPaneConstants.HORIZONTAL_SCROLLBAR_AS_NEEDED);
        int         minHeight = new AttributePanel(null, new AttributeDef(new JsonMap(), 0), null).getPreferredSize().height + 8;
        add(scroller, new PrecisionLayoutData().setHorizontalSpan(2).setFillAlignment().setGrabSpace(true).setMinimumHeight(minHeight));
    }

    public void reset(Map<String, AttributeDef> attributes) {
        mListPanel.reset(attributes);
    }
}
