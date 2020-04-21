/*
 * Copyright ©1998-2020 by Richard A. Wilkes. All rights reserved.
 *
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, version 2.0. If a copy of the MPL was not distributed with
 * this file, You can obtain one at http://mozilla.org/MPL/2.0/.
 *
 * This Source Code Form is "Incompatible With Secondary Licenses", as
 * defined by the Mozilla Public License, version 2.0.
 */

package com.trollworks.gcs.ui.widget;

import com.trollworks.gcs.ui.Colors;

import java.awt.Color;
import java.awt.Component;
import java.util.List;
import javax.swing.JComboBox;
import javax.swing.ListCellRenderer;
import javax.swing.MutableComboBoxModel;
import javax.swing.event.ListDataListener;

public class PopupButton<T> extends JComboBox<T> {
    MutableComboBoxModel<T> mModel;
    private ListCellRenderer<? super T> mRenderer;

    @SuppressWarnings("unchecked")
    public PopupButton(List<T> items) {
        //noinspection SuspiciousArrayCast
        this((T[]) items.toArray());
    }

    public PopupButton(T[] items) {
        super(items);
        mModel = (MutableComboBoxModel<T>) getModel();
        mRenderer = getRenderer();
        setModel(new MutableComboBoxModel<>() {
            @Override
            public int getSize() {
                return mModel.getSize();
            }

            @Override
            public T getElementAt(int index) {
                return mModel.getElementAt(index);
            }

            @Override
            public void addListDataListener(ListDataListener listener) {
                mModel.addListDataListener(listener);
            }

            @Override
            public void removeListDataListener(ListDataListener listener) {
                mModel.removeListDataListener(listener);
            }

            @Override
            public void setSelectedItem(Object item) {
                if (item instanceof Enabled) {
                    if (!((Enabled) item).isEnabled()) {
                        return;
                    }
                }
                mModel.setSelectedItem(item);
            }

            @Override
            public Object getSelectedItem() {
                return mModel.getSelectedItem();
            }

            @Override
            public void addElement(T item) {
                mModel.addElement(item);
            }

            @Override
            public void insertElementAt(T item, int index) {
                mModel.insertElementAt(item, index);
            }

            @Override
            public void removeElement(Object obj) {
                mModel.removeElement(obj);
            }

            @Override
            public void removeElementAt(int index) {
                mModel.removeElementAt(index);
            }
        });
        setRenderer((list, value, index, isSelected, cellHasFocus) -> {
            Component comp = mRenderer.getListCellRendererComponent(list, value, index, isSelected, cellHasFocus);
            if (!(value instanceof Enabled && ((Enabled) value).isEnabled())) {
                Color background = list.getBackground();
                comp.setBackground(background);
                comp.setForeground(Colors.adjustBrightness(background, -0.2f));
            }
            return comp;
        });
    }
}
