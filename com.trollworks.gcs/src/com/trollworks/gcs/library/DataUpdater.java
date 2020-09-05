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

package com.trollworks.gcs.library;

import com.trollworks.gcs.advantage.AdvantageList;
import com.trollworks.gcs.character.GURPSCharacter;
import com.trollworks.gcs.datafile.DataFile;
import com.trollworks.gcs.datafile.Updatable;
import com.trollworks.gcs.equipment.EquipmentList;
import com.trollworks.gcs.modifier.AdvantageModifierList;
import com.trollworks.gcs.modifier.EquipmentModifierList;
import com.trollworks.gcs.notes.NoteList;
import com.trollworks.gcs.skill.SkillList;
import com.trollworks.gcs.spell.SpellList;
import com.trollworks.gcs.template.Template;
import com.trollworks.gcs.utility.FileType;
import com.trollworks.gcs.utility.PathUtils;

import java.io.IOException;
import java.nio.file.DirectoryStream;
import java.nio.file.Files;
import java.nio.file.Path;
import java.util.ArrayList;
import java.util.HashMap;
import java.util.List;
import java.util.Map;
import java.util.UUID;

public class DataUpdater {
    public Map<UUID, Updatable>  updatables;
    public Map<Path, List<Path>> ignoreMap;

    public DataUpdater() throws IOException {
        Map<UUID, Path> uuidToPathMap = new HashMap<>();
        updatables = new HashMap<>();
        ignoreMap = new HashMap<>();
        for (Library library : Library.LIBRARIES) {
            traverse(uuidToPathMap, library.getPath());
        }
    }

    private void traverse(Map<UUID, Path> uuidToPathMap, Path dir) throws IOException {
        if (!shouldSkip(dir)) {
            try (DirectoryStream<Path> stream = Files.newDirectoryStream(dir)) {
                for (Path path : stream) {
                    if (Files.isDirectory(path)) {
                        traverse(uuidToPathMap, path);
                    } else if (!shouldSkip(path)) {
                        String ext = PathUtils.getExtension(path.getFileName());
                        if (FileType.SHEET.matchExtension(ext)) {
                            add(uuidToPathMap, path, new GURPSCharacter(path));
                        } else if (FileType.TEMPLATE.matchExtension(ext)) {
                            add(uuidToPathMap, path, new Template(path));
                        } else if (FileType.ADVANTAGE.matchExtension(ext)) {
                            addList(uuidToPathMap, path, new AdvantageList());
                        } else if (FileType.ADVANTAGE_MODIFIER.matchExtension(ext)) {
                            addList(uuidToPathMap, path, new AdvantageModifierList());
                        } else if (FileType.EQUIPMENT.matchExtension(ext)) {
                            addList(uuidToPathMap, path, new EquipmentList());
                        } else if (FileType.EQUIPMENT_MODIFIER.matchExtension(ext)) {
                            addList(uuidToPathMap, path, new EquipmentModifierList());
                        } else if (FileType.SKILL.matchExtension(ext)) {
                            addList(uuidToPathMap, path, new SkillList());
                        } else if (FileType.SPELL.matchExtension(ext)) {
                            addList(uuidToPathMap, path, new SpellList());
                        } else if (FileType.NOTE.matchExtension(ext)) {
                            addList(uuidToPathMap, path, new NoteList());
                        }
                    }
                }
            }
        }
    }

    private void addList(Map<UUID, Path> uuidToPathMap, Path path, Updatable updatable) throws IOException {
        ((DataFile) updatable).load(path);
        add(uuidToPathMap, path, updatable);
    }

    private void add(Map<UUID, Path> uuidToPathMap, Path path, Updatable updatable) {
        UUID id = updatable.getID();
        if (uuidToPathMap.containsKey(id)) {
            ignoreMap.computeIfAbsent(uuidToPathMap.get(id), k -> new ArrayList<>()).add(path);
        } else {
            uuidToPathMap.put(id, path);
            updatables.put(updatable.getID(), updatable);
            updatable.getContainedUpdatables(updatables);
        }
    }

    private boolean shouldSkip(Path path) {
        return path.getFileName().toString().startsWith(".");
    }
}
