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

package com.trollworks.gcs.character.panels;

import com.trollworks.gcs.character.CharacterSheet;
import com.trollworks.gcs.character.Profile;
import com.trollworks.gcs.page.DropPanel;
import com.trollworks.gcs.ui.GraphicsUtilities;
import com.trollworks.gcs.ui.RetinaIcon;
import com.trollworks.gcs.ui.ThemeFont;
import com.trollworks.gcs.ui.border.TitledBorder;
import com.trollworks.gcs.ui.image.Images;
import com.trollworks.gcs.ui.image.Img;
import com.trollworks.gcs.ui.scale.Scale;
import com.trollworks.gcs.ui.widget.StdFileDialog;
import com.trollworks.gcs.ui.widget.WindowUtils;
import com.trollworks.gcs.utility.FileType;
import com.trollworks.gcs.utility.I18n;
import com.trollworks.gcs.utility.Log;
import com.trollworks.gcs.utility.PathUtils;
import com.trollworks.gcs.utility.text.Text;

import java.awt.Color;
import java.awt.Dimension;
import java.awt.Graphics;
import java.awt.Graphics2D;
import java.awt.Image;
import java.awt.Insets;
import java.awt.datatransfer.DataFlavor;
import java.awt.dnd.DnDConstants;
import java.awt.dnd.DropTarget;
import java.awt.dnd.DropTargetDragEvent;
import java.awt.dnd.DropTargetDropEvent;
import java.awt.dnd.DropTargetEvent;
import java.awt.dnd.DropTargetListener;
import java.awt.event.MouseAdapter;
import java.awt.event.MouseEvent;
import java.io.File;
import java.nio.file.Path;
import java.text.MessageFormat;
import java.util.List;

/** The character portrait. */
public class PortraitPanel extends DropPanel implements DropTargetListener {
    private CharacterSheet mSheet;

    /**
     * Creates a new character portrait.
     *
     * @param sheet The owning sheet.
     */
    public PortraitPanel(CharacterSheet sheet) {
        super(null, true);
        setBorder(new TitledBorder(ThemeFont.PAGE_LABEL_PRIMARY, I18n.text("Portrait")));
        mSheet = sheet;
        setToolTipText(Text.wrapPlainTextForToolTip(MessageFormat.format(I18n.text("""
                <html><body>
                <b>Double-click</b> to set a character portrait.<br>
                <br>
                The dimensions of the chosen picture should be in a ratio of<br>
                <b>3 pixels wide for every 4 pixels tall</b> to scale without distortion.<br>
                <br>
                Dimensions of <b>{0}x{1}</b> are ideal.
                </body></html>
                """), Integer.valueOf(Profile.PORTRAIT_WIDTH * 2), Integer.valueOf(Profile.PORTRAIT_HEIGHT * 2))));
        if (GraphicsUtilities.hasUserDisplay()) {
            setDropTarget(new DropTarget(this, DnDConstants.ACTION_COPY, this));
        }
        addMouseListener(new MouseAdapter() {
            @Override
            public void mouseClicked(MouseEvent event) {
                if (event.getClickCount() == 2) {
                    choosePortrait();
                }
            }
        });
    }

    /** Allows the user to choose a portrait for their character. */
    public void choosePortrait() {
        Path path = StdFileDialog.showOpenDialog(null, I18n.text("Select A Portrait"), FileType.IMAGE_FILTERS);
        if (path != null) {
            try {
                mSheet.getCharacter().getProfile().setPortrait(Img.create(path));
            } catch (Exception exception) {
                WindowUtils.showError(this, MessageFormat.format(I18n.text("Unable to load\n{0}."), path.normalize().toAbsolutePath()));
            }
        }
    }

    @Override
    protected void paintComponent(Graphics g) {
        Insets     insets = getInsets();
        Graphics2D gc     = GraphicsUtilities.prepare(g);
        gc.setColor(Color.WHITE);
        gc.fillRect(insets.left, insets.top, getWidth() - (insets.left + insets.right), getHeight() - (insets.top + insets.bottom));
        RetinaIcon portrait = mSheet.getCharacter().getProfile().getPortrait();
        if (portrait == null) {
            portrait = Images.DEFAULT_PORTRAIT;
        }
        portrait.paintIcon(this, gc, insets.left, insets.top);
    }

    @Override
    public Dimension getMinimumSize() {
        return getPreferredSize();
    }

    @Override
    public Dimension getPreferredSize() {
        Scale  scale  = Scale.get(this);
        Insets insets = getInsets();
        return new Dimension(insets.left + scale.scale(Profile.PORTRAIT_WIDTH) + insets.right, insets.top + scale.scale(Profile.PORTRAIT_HEIGHT) + insets.bottom);
    }

    @Override
    public Dimension getMaximumSize() {
        return getPreferredSize();
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
        if (dtde.isDataFlavorSupported(DataFlavor.imageFlavor)) {
            try {
                dtde.acceptDrop(DnDConstants.ACTION_COPY);
                Image img = (Image) dtde.getTransferable().getTransferData(DataFlavor.imageFlavor);
                try {
                    mSheet.getCharacter().getProfile().setPortrait(Img.create(img));
                } catch (Exception exception) {
                    WindowUtils.showError(this, I18n.text("Unable to load image."));
                }
                dtde.dropComplete(true);
                dtde.getDropTargetContext().getComponent().requestFocus();
            } catch (Exception exception) {
                Log.error(exception);
            }
            return;
        } else if (dtde.isDataFlavorSupported(DataFlavor.javaFileListFlavor)) {
            try {
                dtde.acceptDrop(DnDConstants.ACTION_COPY);
                @SuppressWarnings("unchecked") List<File> transferData = (List<File>) dtde.getTransferable().getTransferData(DataFlavor.javaFileListFlavor);
                for (File file : transferData) {
                    try {
                        mSheet.getCharacter().getProfile().setPortrait(Img.create(file));
                        break;
                    } catch (Exception exception) {
                        WindowUtils.showError(this, MessageFormat.format(I18n.text("Unable to load\n{0}."), PathUtils.getFullPath(file)));
                    }
                }
                dtde.dropComplete(true);
                dtde.getDropTargetContext().getComponent().requestFocus();
            } catch (Exception exception) {
                Log.error(exception);
            }
            return;
        }
        dtde.dropComplete(false);
    }

    private static void acceptOrRejectDrag(DropTargetDragEvent dtde) {
        if (dtde.isDataFlavorSupported(DataFlavor.imageFlavor) || dtde.isDataFlavorSupported(DataFlavor.javaFileListFlavor)) {
            dtde.acceptDrag(DnDConstants.ACTION_COPY);
        } else {
            dtde.rejectDrag();
        }
    }
}
