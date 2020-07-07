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

package com.trollworks.gcs.menu.file;

import com.trollworks.gcs.character.CharacterSheet;
import com.trollworks.gcs.character.GURPSCharacter;
import com.trollworks.gcs.character.SheetDockable;
import com.trollworks.gcs.character.TextTemplate;
import com.trollworks.gcs.menu.Command;
import com.trollworks.gcs.preferences.OutputPreferences;
import com.trollworks.gcs.preferences.Preferences;
import com.trollworks.gcs.ui.widget.WindowUtils;
import com.trollworks.gcs.utility.I18n;
import com.trollworks.gcs.utility.Log;
import com.trollworks.gcs.utility.json.JsonWriter;

import java.awt.Color;
import java.awt.Desktop;
import java.awt.Font;
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
import java.net.URISyntaxException;
import java.net.URL;
import java.net.URLConnection;
import java.nio.charset.StandardCharsets;
import java.nio.file.Files;
import java.text.MessageFormat;
import java.util.Scanner;
import java.util.UUID;
import javax.imageio.ImageIO;
import javax.swing.JEditorPane;
import javax.swing.JLabel;
import javax.swing.JOptionPane;
import javax.swing.event.HyperlinkEvent;

public class ExportToGURPSCalculatorCommand extends Command {
    /** The singleton {@link ExportToGURPSCalculatorCommand}. */
    public static final ExportToGURPSCalculatorCommand INSTANCE = new ExportToGURPSCalculatorCommand();

    private ExportToGURPSCalculatorCommand() {
        super(I18n.Text("Export to GURPS Calculator…"), "ExportToGURPSCalculator", KeyEvent.VK_L);
    }

    @Override
    public void adjust() {
    }

    @Override
    public void actionPerformed(ActionEvent event) {
        SheetDockable dockable = getTarget(SheetDockable.class);
        if (dockable != null) {
            CharacterSheet sheet     = dockable.getSheet();
            GURPSCharacter character = sheet.getCharacter();
            try {
                String key = Preferences.getInstance().getGURPSCalculatorKey();
                if ("true".equals(get(String.format("api/GetCharacterExists/%s/%s", character.getId(), key)))) {
                    String cancel = I18n.Text("Cancel");
                    switch (JOptionPane.showOptionDialog(KeyboardFocusManager.getCurrentKeyboardFocusManager().getPermanentFocusOwner(), I18n.Text("This character already exists in GURPS Calculator.\nWould you like to replace it?\n\nIf you choose 'Create New', you should save your\ncharacter afterwards."), I18n.Text("Character Exists"), JOptionPane.YES_NO_CANCEL_OPTION, JOptionPane.QUESTION_MESSAGE, null, new Object[]{I18n.Text("Replace"), I18n.Text("Create New"), cancel}, cancel)) {
                    case JOptionPane.NO_OPTION:
                        character.generateNewId();
                        character.setModified(true);
                        break;
                    case JOptionPane.CANCEL_OPTION:
                        return;
                    default:
                        break;
                    }
                }
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
                            UUID   id   = character.getId();
                            String path = String.format("api/SaveCharacter/%s/%s", id, key);
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
                                    character.save(w, true, false);
                                }
                                path = String.format("api/SaveCharacterRawFileGCS/%s/%s", id, key);
                                result = post(path, out.toByteArray());
                                if (!result.isEmpty()) {
                                    throw new IOException("Bad response from the web server for GCS file write");
                                }
                            }
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
    }

    private void showResult(boolean success) {
        String message = success ? I18n.Text("Export to GURPS Calculator was successful.") : I18n.Text("There was an error exporting to GURPS Calculator. Please try again later.");
        String key     = Preferences.getInstance().getGURPSCalculatorKey();
        if (key == null || !key.matches("[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[1-5][0-9a-fA-F]{3}-[89ab][0-9a-fA-F]{3}-[0-9a-fA-F]{12}")) {
            message = String.format(I18n.Text("You need to set a valid GURPS Calculator Key in sheet preferences.<br><a href='%s'>Click here</a> for more information."), OutputPreferences.GURPS_CALCULATOR_URL);
        }
        JLabel      styleLabel  = new JLabel();
        Font        font        = styleLabel.getFont();
        Color       color       = styleLabel.getBackground();
        JEditorPane messagePane = new JEditorPane("text/html", "<html><body style='font-family:" + font.getFamily() + ";font-weight:" + (font.isBold() ? "bold" : "normal") + ";font-size:" + font.getSize() + "pt;background-color: rgb(" + color.getRed() + "," + color.getGreen() + "," + color.getBlue() + ");'>" + message + "</body></html>");
        messagePane.setEditable(false);
        messagePane.setBorder(null);
        messagePane.addHyperlinkListener(event -> {
            if (Desktop.isDesktopSupported() && event.getEventType().equals(HyperlinkEvent.EventType.ACTIVATED)) {
                URL url = event.getURL();
                try {
                    Desktop.getDesktop().browse(url.toURI());
                } catch (IOException | URISyntaxException exception) {
                    WindowUtils.showError(null, MessageFormat.format(I18n.Text("Unable to open {0}"), url.toExternalForm()));
                }
            }
        });
        JOptionPane.showMessageDialog(Command.getFocusOwner(), messagePane, success ? I18n.Text("Success") : I18n.Text("Error"), success ? JOptionPane.INFORMATION_MESSAGE : JOptionPane.ERROR_MESSAGE);
    }

    public String get(String path) throws IOException {
        URLConnection connection = prepare(path);
        return retrieveResponse(connection);
    }

    public String post(String path, String body) throws IOException {
        URLConnection connection = prepare(path);
        connection.setDoOutput(true);
        connection.setRequestProperty("Content-Type", "application/json");
        try (OutputStreamWriter writer = new OutputStreamWriter(connection.getOutputStream(), StandardCharsets.UTF_8)) {
            writer.write(body);
        }
        return retrieveResponse(connection);
    }

    public String post(String path, byte[] body) throws IOException {
        URLConnection connection = prepare(path);
        connection.setDoOutput(true);
        connection.setRequestProperty("Content-Type", "application/json");
        try (DataOutputStream writer = new DataOutputStream(connection.getOutputStream())) {
            writer.write(body);
        }
        return retrieveResponse(connection);
    }

    private URLConnection prepare(String path) throws IOException {
        URLConnection connection = new URL(OutputPreferences.BASE_GURPS_CALCULATOR_URL + "/" + path + "/").openConnection();
        connection.setRequestProperty("Accept-Charset", StandardCharsets.UTF_8.toString());
        return connection;
    }

    private String retrieveResponse(URLConnection connection) throws IOException {
        try (InputStream stream = connection.getInputStream()) {
            try (Scanner s = new Scanner(stream, StandardCharsets.UTF_8)) {
                s.useDelimiter("\\A");
                return s.hasNext() ? s.next() : "";
            }
        }
    }
}
