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

package com.trollworks.gcs.menu.file;

import com.trollworks.gcs.character.CharacterSheet;
import com.trollworks.gcs.character.GURPSCharacter;
import com.trollworks.gcs.character.SheetDockable;
import com.trollworks.gcs.character.TextTemplate;
import com.trollworks.gcs.menu.Command;
import com.trollworks.gcs.settings.QuickExport;
import com.trollworks.gcs.settings.Settings;
import com.trollworks.gcs.ui.widget.MessageType;
import com.trollworks.gcs.ui.widget.StdDialog;
import com.trollworks.gcs.utility.I18n;
import com.trollworks.gcs.utility.Log;
import com.trollworks.gcs.utility.SaveType;
import com.trollworks.gcs.utility.json.JsonWriter;

import java.awt.KeyboardFocusManager;
import java.awt.event.ActionEvent;
import java.awt.event.KeyEvent;
import java.io.ByteArrayOutputStream;
import java.io.DataOutputStream;
import java.io.File;
import java.io.FileNotFoundException;
import java.io.IOException;
import java.io.InputStream;
import java.io.OutputStreamWriter;
import java.io.PrintWriter;
import java.net.URL;
import java.net.URLConnection;
import java.nio.charset.StandardCharsets;
import java.nio.file.Files;
import java.util.Scanner;
import java.util.UUID;
import java.util.regex.Pattern;
import javax.imageio.ImageIO;

public final class ExportToGURPSCalculatorCommand extends Command {
    public static final  String                         BASE_GURPS_CALCULATOR_URL = "http://www.gurpscalculator.com";
    public static final  String                         GURPS_CALCULATOR_URL      = BASE_GURPS_CALCULATOR_URL + "/Character/ImportGCS";
    public static final  ExportToGURPSCalculatorCommand INSTANCE                  = new ExportToGURPSCalculatorCommand();
    private static final Pattern                        UUID_PATTERN              = Pattern.compile("[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[1-5][0-9a-fA-F]{3}-[89ab][0-9a-fA-F]{3}-[0-9a-fA-F]{12}");

    private ExportToGURPSCalculatorCommand() {
        super(I18n.text("GURPS Calculator…"), "ExportToGURPSCalculator", KeyEvent.VK_L);
    }

    @Override
    public void adjust() {
        setEnabled(Command.getTarget(SheetDockable.class) != null);
    }

    @Override
    public void actionPerformed(ActionEvent event) {
        performExport(getTarget(SheetDockable.class));
    }

    public static void performExport(SheetDockable dockable) {
        if (dockable != null) {
            CharacterSheet sheet     = dockable.getSheet();
            GURPSCharacter character = sheet.getCharacter();
            String         key       = Settings.getInstance().getGURPSCalculatorKey();
            try {
                if ("true".equals(get(String.format("api/GetCharacterExists/%s/%s", character.getID(), key)))) {
                    StdDialog dialog = StdDialog.prepareToShowMessage(KeyboardFocusManager.getCurrentKeyboardFocusManager().getPermanentFocusOwner(),
                            I18n.text("Character already exists"), MessageType.WARNING,
                            I18n.text("""
                                    This character already exists in GURPS Calculator.
                                    Would you like to replace it?

                                    If you choose 'Create New', you should save your
                                    character afterwards."""));
                    dialog.addButton(I18n.text("Replace"), (btn) -> {
                        export(dockable);
                        dialog.setVisible(false);
                    });
                    dialog.addButton(I18n.text("Create New"), (btn) -> {
                        character.generateNewID();
                        character.setModified(true);
                        export(dockable);
                        dialog.setVisible(false);
                    });
                    dialog.addCancelButton();
                    dialog.presentToUser();
                } else {
                    export(dockable);
                }
            } catch (Exception exception) {
                Log.error(exception);
                showResult(false);
            }
        }
    }

    private static void export(SheetDockable dockable) {
        CharacterSheet sheet = dockable.getSheet();
        try {
            File templateFile = File.createTempFile("gcalcTemplate", ".html");
            try {
                try (PrintWriter out = new PrintWriter(templateFile, StandardCharsets.UTF_8)) {
                    out.print(get("api/GetOutputTemplate"));
                }
                File outputFile = File.createTempFile("gcalcOutput", ".html");
                try {
                    if (new TextTemplate(sheet).export(outputFile.toPath(), templateFile.toPath())) {
                        String result = null;
                        try (Scanner scanner = new Scanner(outputFile, StandardCharsets.UTF_8)) {
                            result = scanner.useDelimiter("\\A").next();
                        } catch (FileNotFoundException exception) {
                            Log.error(exception);
                        }
                        GURPSCharacter character = sheet.getCharacter();
                        UUID           id        = character.getID();
                        String         key       = Settings.getInstance().getGURPSCalculatorKey();
                        String         path      = String.format("api/SaveCharacter/%s/%s", id, key);
                        result = post(path, result);
                        if (!result.isEmpty()) {
                            throw new IOException("Bad response from the web server for template write");
                        }
                        File image = File.createTempFile("gcalcImage", ".png");
                        try {
                            ImageIO.write(character.getProfile().getPortrait().getRetina(), "png", image);
                            path = String.format("api/SaveCharacterImage/%s/%s", id, key);
                            result = post(path, Files.readAllBytes(image.toPath()));
                            if (!result.isEmpty()) {
                                throw new IOException("Bad response from the web server for image write");
                            }
                        } finally {
                            image.delete();
                        }
                        try (ByteArrayOutputStream out = new ByteArrayOutputStream()) {
                            try (JsonWriter w = new JsonWriter(new OutputStreamWriter(out, StandardCharsets.UTF_8), "\t")) {
                                character.save(w, SaveType.NORMAL, false);
                            }
                            path = String.format("api/SaveCharacterRawFileGCS/%s/%s", id, key);
                            result = post(path, out.toByteArray());
                            if (!result.isEmpty()) {
                                throw new IOException("Bad response from the web server for GCS file write");
                            }
                        }
                        dockable.recordQuickExport(new QuickExport());
                        showResult(true);
                    } else {
                        showResult(false);
                    }
                } finally {
                    outputFile.delete();
                }
            } finally {
                templateFile.delete();
            }
        } catch (Exception exception) {
            Log.error(exception);
            showResult(false);
        }
    }

    private static void showResult(boolean success) {
        if (success) {
            StdDialog.showMessage(Command.getFocusOwner(), I18n.text("Success"), MessageType.NONE,
                    I18n.text("Export to GURPS Calculator was successful."));
        } else {
            String key = Settings.getInstance().getGURPSCalculatorKey();
            String message;
            if (key == null || !UUID_PATTERN.matcher(key).matches()) {
                message = I18n.text("You must first set a valid GURPS Calculator Key in General Settings.");
            } else {
                message = I18n.text("There was an error exporting to GURPS Calculator.\nPlease try again later.");
            }
            StdDialog.showError(Command.getFocusOwner(), message);
        }
    }

    public static String get(String path) throws IOException {
        URLConnection connection = prepare(path);
        return retrieveResponse(connection);
    }

    public static String post(String path, String body) throws IOException {
        URLConnection connection = prepare(path);
        connection.setDoOutput(true);
        connection.setRequestProperty("Content-Type", "application/json");
        try (OutputStreamWriter writer = new OutputStreamWriter(connection.getOutputStream(), StandardCharsets.UTF_8)) {
            writer.write(body);
        }
        return retrieveResponse(connection);
    }

    public static String post(String path, byte[] body) throws IOException {
        URLConnection connection = prepare(path);
        connection.setDoOutput(true);
        connection.setRequestProperty("Content-Type", "application/json");
        try (DataOutputStream writer = new DataOutputStream(connection.getOutputStream())) {
            writer.write(body);
        }
        return retrieveResponse(connection);
    }

    private static URLConnection prepare(String path) throws IOException {
        URLConnection connection = new URL(BASE_GURPS_CALCULATOR_URL + "/" + path + "/").openConnection();
        connection.setRequestProperty("Accept-Charset", StandardCharsets.UTF_8.toString());
        return connection;
    }

    private static String retrieveResponse(URLConnection connection) throws IOException {
        try (InputStream stream = connection.getInputStream()) {
            try (Scanner s = new Scanner(stream, StandardCharsets.UTF_8)) {
                s.useDelimiter("\\A");
                return s.hasNext() ? s.next() : "";
            }
        }
    }
}
