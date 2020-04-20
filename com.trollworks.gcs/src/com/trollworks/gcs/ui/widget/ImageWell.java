/*
 * Copyright (c) 1998-2020 by Richard A. Wilkes. All rights reserved.
 *
 * This Source Code Form is subject to the terms of the Mozilla Public License, version 2.0.
 * If a copy of the MPL was not distributed with this file, You can obtain one at
 * http://mozilla.org/MPL/2.0/.
 *
 * This Source Code Form is "Incompatible With Secondary Licenses", as defined by the
 * Mozilla Public License, version 2.0.
 */

package com.trollworks.gcs.ui.widget;

import com.trollworks.gcs.io.Log;
import com.trollworks.gcs.ui.UIUtilities;
import com.trollworks.gcs.ui.border.LineBorder;
import com.trollworks.gcs.ui.image.Img;
import com.trollworks.gcs.utility.FileType;
import com.trollworks.gcs.utility.I18n;
import com.trollworks.gcs.utility.text.Text;

import java.awt.Color;
import java.awt.Dimension;
import java.awt.FlowLayout;
import java.awt.Graphics;
import java.awt.Rectangle;
import java.awt.datatransfer.DataFlavor;
import java.awt.dnd.DnDConstants;
import java.awt.dnd.DropTarget;
import java.awt.dnd.DropTargetDragEvent;
import java.awt.dnd.DropTargetDropEvent;
import java.awt.dnd.DropTargetEvent;
import java.awt.dnd.DropTargetListener;
import java.awt.event.MouseEvent;
import java.awt.event.MouseListener;
import java.io.File;
import java.io.IOException;
import java.util.List;
import javax.swing.JButton;
import javax.swing.JFileChooser;
import javax.swing.JPanel;

public class ImageWell extends JPanel implements DropTargetListener, MouseListener {
    private Getter mGetter;
    private Setter mSetter;

    public ImageWell(String tooltip, Getter getter, Setter setter) {
        mGetter = getter;
        mSetter = setter;
        setToolTipText(Text.wrapPlainTextForToolTip(tooltip));
        setBorder(new LineBorder());
        UIUtilities.setOnlySize(this, new Dimension(22, 22));
        setDropTarget(new DropTarget(this, DnDConstants.ACTION_COPY, this));
        addMouseListener(this);
    }

    @Override
    protected void paintComponent(Graphics g) {
        super.paintComponent(g);
        Rectangle bounds = UIUtilities.getLocalInsetBounds(this);
        Img       img    = mGetter.getWellImage();
        if (img != null) {
            g.drawImage(img, bounds.x, bounds.y, bounds.width, bounds.height, this);
        } else {
            g.setColor(Color.WHITE);
            g.fillRect(bounds.x, bounds.y, bounds.width, bounds.height);
            g.setColor(Color.LIGHT_GRAY);
            int xs     = bounds.width / 4;
            int ys     = bounds.height / 4;
            int offset = 0;
            for (int y = bounds.y; y < bounds.y + bounds.height; y += ys) {
                for (int x = bounds.x + offset; x < bounds.x + bounds.width; x += xs * 2) {
                    g.fillRect(x, y, xs, ys);
                }
                offset = offset == 0 ? xs : 0;
            }
        }
    }

    @Override
    public void dragEnter(DropTargetDragEvent dtde) {
        acceptOrRejectDrag(dtde);
    }

    @Override
    public void dragOver(DropTargetDragEvent dtde) {
        acceptOrRejectDrag(dtde);
    }

    @Override
    public void dropActionChanged(DropTargetDragEvent dtde) {
        acceptOrRejectDrag(dtde);
    }

    @Override
    public void dragExit(DropTargetEvent dte) {
        // Unused
    }

    @Override
    public void drop(DropTargetDropEvent dtde) {
        for (DataFlavor dataFlavor : dtde.getCurrentDataFlavors()) {
            if (dataFlavor.isFlavorJavaFileListType()) {
                try {
                    dtde.acceptDrop(DnDConstants.ACTION_COPY);
                    @SuppressWarnings("unchecked") List<File> transferData = (List<File>) dtde.getTransferable().getTransferData(dataFlavor);
                    for (File file : transferData) {
                        if (loadImageFile(file)) {
                            break;
                        }
                    }
                    dtde.dropComplete(true);
                    dtde.getDropTargetContext().getComponent().requestFocusInWindow();
                } catch (Exception exception) {
                    Log.error(exception);
                }
                return;
            }
        }
        dtde.dropComplete(false);
    }

    private static void acceptOrRejectDrag(DropTargetDragEvent dtde) {
        if (dtde.isDataFlavorSupported(DataFlavor.javaFileListFlavor)) {
            dtde.acceptDrag(DnDConstants.ACTION_COPY);
        } else {
            dtde.rejectDrag();
        }
    }

    private boolean loadImageFile(File file) {
        try {
            mSetter.setWellImage(Img.create(file));
            repaint();
            return true;
        } catch (IOException exception) {
            return false;
        }
    }

    @Override
    public void mouseClicked(MouseEvent event) {
        JPanel  panel  = new JPanel(new FlowLayout());
        JButton button = new JButton(I18n.Text("Clear Image"));
        button.addActionListener(action -> {
            mSetter.setWellImage(null);
            repaint();
            JFileChooser dialog = UIUtilities.getAncestorOfType(button, JFileChooser.class);
            if (dialog != null) {
                dialog.cancelSelection();
            }
        });
        panel.add(button);
        File file = StdFileDialog.showOpenDialog(null, I18n.Text("Select an image file"), panel, FileType.IMAGE_FILTERS);
        if (file != null) {
            loadImageFile(file);
        }
    }

    @Override
    public void mousePressed(MouseEvent event) {
        // Unused
    }

    @Override
    public void mouseReleased(MouseEvent event) {
        // Unused
    }

    @Override
    public void mouseEntered(MouseEvent event) {
        // Unused
    }

    @Override
    public void mouseExited(MouseEvent event) {
        // Unused
    }

    public interface Getter {
        Img getWellImage();
    }

    public interface Setter {
        void setWellImage(Img img);
    }
}
