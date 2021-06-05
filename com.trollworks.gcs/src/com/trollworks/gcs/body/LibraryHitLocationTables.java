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

package com.trollworks.gcs.body;

import com.trollworks.gcs.library.Library;
import com.trollworks.gcs.preferences.Preferences;
import com.trollworks.gcs.utility.Dice;
import com.trollworks.gcs.utility.I18n;
import com.trollworks.gcs.utility.Log;

import java.io.IOException;
import java.nio.file.DirectoryStream;
import java.nio.file.Files;
import java.nio.file.Path;
import java.util.ArrayList;
import java.util.Collections;
import java.util.List;

public final class LibraryHitLocationTables {
    private static HitLocationTable       mHumanoid;
    private        String                 mLibraryName;
    private        List<HitLocationTable> mTables;

    private LibraryHitLocationTables(Library library) {
        mLibraryName = library.getPath().getFileName().toString();
        mTables = new ArrayList<>();
    }

    @Override
    public String toString() {
        return mLibraryName;
    }

    public List<HitLocationTable> getTables() {
        return mTables;
    }

    public static synchronized List<LibraryHitLocationTables> get() {
        List<LibraryHitLocationTables> all = new ArrayList<>();
        mHumanoid = null;
        Preferences.getInstance(); // Just to ensure the libraries list is initialized
        for (Library lib : Library.LIBRARIES) {
            Path dir = lib.getPath().resolve("Hit Locations");
            if (Files.isDirectory(dir)) {
                LibraryHitLocationTables libTables = new LibraryHitLocationTables(lib);
                List<HitLocationTable>   tables    = libTables.getTables();
                // IMPORTANT: On Windows, calling any of the older methods to list the contents of a
                // directory results in leaving state around that prevents future move & delete
                // operations. Only use this style of access for directory listings to avoid that.
                try (DirectoryStream<Path> stream = Files.newDirectoryStream(dir)) {
                    for (Path path : stream) {
                        try {
                            HitLocationTable table = new HitLocationTable(path);
                            tables.add(table);
                            if (lib == Library.MASTER && "humanoid".equals(table.getID())) {
                                mHumanoid = table;
                            }
                        } catch (IOException ioe) {
                            Log.error("unable to load " + path, ioe);
                        }
                    }
                } catch (IOException exception) {
                    Log.error(exception);
                }
                if (lib == Library.MASTER && mHumanoid == null) {
                    mHumanoid = createHumanoidTable();
                    tables.add(mHumanoid);
                }
                if (!tables.isEmpty()) {
                    Collections.sort(tables);
                    all.add(libTables);
                }
            }
        }
        return all;
    }

    public static synchronized HitLocationTable getHumanoid() {
        if (mHumanoid == null) {
            mHumanoid = createHumanoidTable();
        }
        return mHumanoid;
    }

    private static HitLocationTable createHumanoidTable() {
        HitLocationTable table = new HitLocationTable("humanoid", I18n.Text("Humanoid"), new Dice(3));
        table.addLocation(new HitLocation("eye", I18n.Text("Eyes"), 0, -9, 0, I18n.Text("An attack that misses by 1 hits the torso instead. Only impaling (imp), piercing (pi-, pi, pi+, pi++), and tight-beam burning (burn) attacks can target the eye – and only from the front or sides. Injury over HP÷10 blinds the eye. Otherwise, treat as skull, but without the extra DR!")));
        table.addLocation(new HitLocation("skull", I18n.Text("Skull"), 2, -7, 2, I18n.Text("An attack that misses by 1 hits the torso instead. Wounding modifier is x4. Knockdown rolls are at -10. Critical hits use the Critical Head Blow Table (B556). Exception: These special effects do not apply to toxic (tox) damage.")));
        table.addLocation(new HitLocation("face", I18n.Text("Face"), 1, -5, 0, I18n.Text("An attack that misses by 1 hits the torso instead. Jaw, cheeks, nose, ears, etc. If the target has an open-faced helmet, ignore its DR. Knockdown rolls are at -5. Critical hits use the Critical Head Blow Table (B556). Corrosion (cor) damage gets a x1½ wounding modifier, and if it inflicts a major wound, it also blinds one eye (both eyes on damage over full HP). Random attacks from behind hit the skull instead.")));
        table.addLocation(new HitLocation("leg", I18n.Text("Leg"), I18n.Text("Right Leg"), 2, -2, 0, getLimbDescription()));
        table.addLocation(new HitLocation("arm", I18n.Text("Arm"), I18n.Text("Right Arm"), 1, -2, 0, getArmDescription()));
        table.addLocation(new HitLocation("torso", I18n.Text("Torso"), 2, 0, 0, ""));
        table.addLocation(new HitLocation("groin", I18n.Text("Groin"), 1, -3, 0, I18n.Text("An attack that misses by 1 hits the torso instead. Human males and the males of similar species suffer double shock from crushing (cr) damage, and get -5 to knockdown rolls. Otherwise, treat as a torso hit.")));
        table.addLocation(new HitLocation("arm", I18n.Text("Arm"), I18n.Text("Left Arm"), 1, -2, 0, getArmDescription()));
        table.addLocation(new HitLocation("leg", I18n.Text("Leg"), I18n.Text("Left Leg"), 2, -2, 0, getLimbDescription()));
        table.addLocation(new HitLocation("hand", I18n.Text("Hand"), 1, -4, 0, String.format(I18n.Text("If holding a shield, double the penalty to hit: -8 for shield hand instead of -4. %s"), getExtremityDescription())));
        table.addLocation(new HitLocation("foot", I18n.Text("Foot"), 1, -4, 0, getExtremityDescription()));
        table.addLocation(new HitLocation("neck", I18n.Text("Neck"), 2, -5, 0, I18n.Text("An attack that misses by 1 hits the torso instead. Neck and throat. Increase the wounding multiplier of crushing (cr) and corrosion (cor) attacks to x1½, and that of cutting (cut) damage to x2. At the GM’s option, anyone killed by a cutting (cut) blow to the neck is decapitated!")));
        table.addLocation(new HitLocation("vitals", I18n.Text("Vitals"), 0, -3, 0, I18n.Text("An attack that misses by 1 hits the torso instead. Heart, lungs, kidneys, etc. Increase the wounding modifier for an impaling (imp) or any piercing (pi-, pi, pi+, pi++) attack to x3. Increase the wounding modifier for a tight-beam burning (burn) attack to x2. Other attacks cannot target the vitals.")));
        table.update();
        return table;
    }

    private static String getLimbDescription() {
        return I18n.Text("Reduce the wounding multiplier of large piercing (pi+), huge piercing (pi++), and impaling (imp) damage to x1. Any major wound (loss of over ½ HP from one blow) cripples the limb. Damage beyond that threshold is lost.");
    }

    private static String getArmDescription() {
        return String.format(I18n.Text("%s If holding a shield, double the penalty to hit: -4 for shield arm instead of -2."), getLimbDescription());
    }

    private static String getExtremityDescription() {
        return I18n.Text("Reduce the wounding multiplier of large piercing (pi+), huge piercing (pi++), and impaling (imp) damage to x1. Any major wound (loss of over ⅓ HP from one blow) cripples the extremity. Damage beyond that threshold is lost.");
    }
}
