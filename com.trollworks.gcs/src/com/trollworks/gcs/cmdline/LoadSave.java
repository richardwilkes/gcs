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

package com.trollworks.gcs.cmdline;

import com.trollworks.gcs.advantage.AdvantageList;
import com.trollworks.gcs.character.GURPSCharacter;
import com.trollworks.gcs.datafile.DataFile;
import com.trollworks.gcs.equipment.EquipmentList;
import com.trollworks.gcs.modifier.AdvantageModifierList;
import com.trollworks.gcs.modifier.EquipmentModifierList;
import com.trollworks.gcs.notes.NoteList;
import com.trollworks.gcs.skill.SkillList;
import com.trollworks.gcs.spell.SpellList;
import com.trollworks.gcs.template.Template;
import com.trollworks.gcs.ui.UIUtilities;
import com.trollworks.gcs.utility.FileType;
import com.trollworks.gcs.utility.I18n;
import com.trollworks.gcs.utility.PathUtils;

import java.awt.EventQueue;
import java.io.IOException;
import java.nio.file.DirectoryStream;
import java.nio.file.Files;
import java.nio.file.Path;
import java.util.List;

public class LoadSave implements Runnable {
    private List<Path> mPaths;

    public static void process(List<Path> paths) {
        if (paths.isEmpty()) {
            System.err.println(I18n.Text("must specify one or more files or directories to process"));
            System.exit(1);
        }
        System.setProperty("java.awt.headless", Boolean.TRUE.toString());
        UIUtilities.initialize();
        try {
            // This is run on the event queue since much of the sheet logic assumes a UI
            // environment and would otherwise cause concurrent modification exceptions, as the
            // detection of whether it was safe to modify data would be inaccurate.
            EventQueue.invokeAndWait(new LoadSave(paths));
        } catch (Exception exception) {
            exception.printStackTrace(System.err);
            System.exit(1);
        }
    }

    private LoadSave(List<Path> paths) {
        mPaths = paths;
    }

    @Override
    public void run() {
        try {
            for (Path path : mPaths) {
                traverse(path);
            }
        } catch (IOException ioe) {
            ioe.printStackTrace();
            System.out.println(I18n.Text("  ** ERROR ENCOUNTERED **"));
        }
    }

    private void traverse(Path path) throws IOException {
        if (!shouldSkip(path)) {
            if (Files.isDirectory(path)) {
                try (DirectoryStream<Path> stream = Files.newDirectoryStream(path)) {
                    for (Path child : stream) {
                        traverse(child);
                    }
                }
            } else {
                String ext = PathUtils.getExtension(path.getFileName());
                if (FileType.SHEET.matchExtension(ext)) {
                    save(new GURPSCharacter(path), path);
                } else if (FileType.TEMPLATE.matchExtension(ext)) {
                    save(new Template(path), path);
                } else if (FileType.ADVANTAGE.matchExtension(ext)) {
                    loadSave(new AdvantageList(), path);
                } else if (FileType.ADVANTAGE_MODIFIER.matchExtension(ext)) {
                    loadSave(new AdvantageModifierList(), path);
                } else if (FileType.EQUIPMENT.matchExtension(ext)) {
                    loadSave(new EquipmentList(), path);
                } else if (FileType.EQUIPMENT_MODIFIER.matchExtension(ext)) {
                    loadSave(new EquipmentModifierList(), path);
                } else if (FileType.SKILL.matchExtension(ext)) {
                    loadSave(new SkillList(), path);
                } else if (FileType.SPELL.matchExtension(ext)) {
                    loadSave(new SpellList(), path);
                } else if (FileType.NOTE.matchExtension(ext)) {
                    loadSave(new NoteList(), path);
                }
            }
        }
    }

    private static void loadSave(DataFile data, Path path) throws IOException {
        data.load(path);
        save(data, path);
    }

    private static void save(DataFile data, Path path) {
        if (!data.save(path)) {
            System.out.println("failed to save " + path);
            System.exit(1);
        }
        System.out.println(path);
    }

    private boolean shouldSkip(Path path) {
        return path.getFileName().toString().startsWith(".");
    }
}
