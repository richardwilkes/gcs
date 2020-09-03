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

package com.trollworks.gcs.utility;

import com.trollworks.gcs.ui.RetinaIcon;
import com.trollworks.gcs.ui.image.Images;

import java.util.ArrayList;
import java.util.Arrays;
import java.util.List;
import javax.swing.filechooser.FileNameExtensionFilter;

/** Describes a file. */
public class FileType {
    public static final List<FileType>            OPENABLE           = new ArrayList<>();
    public static final List<FileType>            ALL_OPENABLE       = new ArrayList<>();
    public static final FileType                  SHEET              = new FileType(I18n.Text("Character Sheet"), Images.GCS_FILE, true, true, "gcs");
    public static final FileType                  TEMPLATE           = new FileType(I18n.Text("Character Template"), Images.GCT_FILE, true, true, "gct");
    public static final FileType                  ADVANTAGE          = new FileType(I18n.Text("Advantages Library"), Images.ADQ_FILE, true, true, "adq");
    public static final FileType                  ADVANTAGE_MODIFIER = new FileType(I18n.Text("Advantage Modifiers Library"), Images.ADM_FILE, true, true, "adm");
    public static final FileType                  EQUIPMENT          = new FileType(I18n.Text("Equipment Library"), Images.EQP_FILE, true, true, "eqp");
    public static final FileType                  EQUIPMENT_MODIFIER = new FileType(I18n.Text("Equipment Modifiers Library"), Images.EQM_FILE, true, true, "eqm");
    public static final FileType                  SKILL              = new FileType(I18n.Text("Skills Library"), Images.SKL_FILE, true, true, "skl");
    public static final FileType                  SPELL              = new FileType(I18n.Text("Spells Library"), Images.SPL_FILE, true, true, "spl");
    public static final FileType                  NOTE               = new FileType(I18n.Text("Notes Library"), Images.NOT_FILE, true, true, "not");
    public static final FileType                  PDF                = new FileType(I18n.Text("PDF Files"), Images.PDF_FILE, false, true, "pdf");
    public static final FileType                  PNG                = new FileType(I18n.Text("PNG Files"), Images.FILE, false, false, "png");
    public static final FileType                  JPEG               = new FileType(I18n.Text("JPEG Files"), Images.FILE, false, false, "jpg", "jpeg");
    public static final FileType                  GIF                = new FileType(I18n.Text("GIF Files"), Images.FILE, false, false, "gif");
    public static final FileNameExtensionFilter[] IMAGE_FILTERS      = createFileFilters(I18n.Text("Image Files"), PNG, JPEG, GIF);
    private             RetinaIcon                mIcon;
    private             String                    mPrimaryExtension;
    private             FileNameExtensionFilter   mFilter;

    public FileType(String description, RetinaIcon icon, boolean owned, boolean openable, String... extension) {
        mIcon = icon;
        mPrimaryExtension = extension[0];
        mFilter = new FileNameExtensionFilter(description, extension);
        if (openable) {
            if (owned) {
                OPENABLE.add(this);
            }
            ALL_OPENABLE.add(this);
        }
    }

    /** @return The icon representing the file. */
    public RetinaIcon getIcon() {
        return mIcon;
    }

    /** @return The primary extension of the file. */
    public String getExtension() {
        return mPrimaryExtension;
    }

    /**
     * @param extension The extension to check.
     * @return {@code true} if the extension matches one for this file type.
     */
    public boolean matchExtension(String extension) {
        for (String ext : mFilter.getExtensions()) {
            if (ext.equalsIgnoreCase(extension)) {
                return true;
            }
        }
        return false;
    }

    /** @return The file filter. */
    public FileNameExtensionFilter getFilter() {
        return mFilter;
    }

    /**
     * @param nameForAggregate The name to use for an aggregate filter that accepts all of the file
     *                         types in the list. Pass in {@code null} for no aggregate filter.
     * @param fileTypes        The set of file types to return filters for.
     * @return A set of file filters for the specified file types.
     */
    public static FileNameExtensionFilter[] createFileFilters(String nameForAggregate, FileType... fileTypes) {
        List<String>                  extensions      = new ArrayList<>();
        List<FileNameExtensionFilter> filters         = new ArrayList<>();
        boolean                       createAggregate = nameForAggregate != null && !nameForAggregate.isBlank();
        for (FileType fileType : fileTypes) {
            if (createAggregate) {
                extensions.addAll(Arrays.asList(fileType.getFilter().getExtensions()));
            }
            filters.add(fileType.getFilter());
        }
        if (createAggregate) {
            filters.add(0, new FileNameExtensionFilter(nameForAggregate, extensions.toArray(new String[0])));
        }
        return filters.toArray(new FileNameExtensionFilter[0]);
    }

    /**
     * @param name The name to return an icon for.
     * @return The icon for the specified file.
     */
    public static RetinaIcon getIconForFileName(String name) {
        String extension = PathUtils.getExtension(name);
        for (FileType one : ALL_OPENABLE) {
            if (one.matchExtension(extension)) {
                return one.getIcon();
            }
        }
        return Images.FILE;
    }
}
