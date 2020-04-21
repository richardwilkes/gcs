/*
 * Copyright (c) 1998-2020 by Richard A. Wilkes. All rights reserved.
 *
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, version 2.0. If a copy of the MPL was not distributed with
 * this file, You can obtain one at http://mozilla.org/MPL/2.0/.
 *
 * This Source Code Form is "Incompatible With Secondary Licenses", as
 * defined by the Mozilla Public License, version 2.0.
 */

package com.trollworks.gcs.character;

import com.trollworks.gcs.io.Log;
import com.trollworks.gcs.io.TemporaryFile;
import com.trollworks.gcs.io.xml.XMLWriter;
import com.trollworks.gcs.menu.Command;
import com.trollworks.gcs.preferences.OutputPreferences;
import com.trollworks.gcs.services.HttpMethodType;
import com.trollworks.gcs.services.NotImplementedException;
import com.trollworks.gcs.services.WebServiceClient;
import com.trollworks.gcs.ui.widget.WindowUtils;
import com.trollworks.gcs.utility.I18n;

import java.awt.Color;
import java.awt.Desktop;
import java.awt.Font;
import java.awt.KeyboardFocusManager;
import java.io.ByteArrayOutputStream;
import java.io.FileNotFoundException;
import java.io.IOException;
import java.io.PrintWriter;
import java.net.URISyntaxException;
import java.net.URL;
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

/** Provides export to GURPS Calculator. */
public class GURPSCalculator {
    public static void export(CharacterSheet sheet) {
        if (sheet != null) {
            GURPSCharacter   character = sheet.getCharacter();
            WebServiceClient client    = new WebServiceClient(OutputPreferences.BASE_GURPS_CALCULATOR_URL);
            try {
                if (showExistsDialogIfNecessary(client, character)) {
                    try (TemporaryFile templateFile = new TemporaryFile("gcalcTemplate", ".html")) {
                        try (PrintWriter out = new PrintWriter(templateFile, StandardCharsets.UTF_8)) {
                            out.print(client.sendRequest(HttpMethodType.GET, "api/GetOutputTemplate"));
                        }
                        try (TemporaryFile outputFile = new TemporaryFile("gcalcOutput", ".html")) {
                            if (new TextTemplate(sheet).export(outputFile, templateFile)) {
                                String result = null;
                                try (Scanner scanner = new Scanner(outputFile, StandardCharsets.UTF_8)) {
                                    result = scanner.useDelimiter("\\A").next();
                                } catch (FileNotFoundException exception) {
                                    Log.error(exception);
                                }
                                UUID   id   = character.getId();
                                String key  = OutputPreferences.getGurpsCalculatorKey();
                                String path = String.format("api/SaveCharacter/%s/%s", id, key);
                                result = client.sendRequest(HttpMethodType.POST, path, null, result);
                                if (!result.isEmpty()) {
                                    throw new IOException("Bad response from the web server for template write");
                                }
                                try (TemporaryFile image = new TemporaryFile("gcalcImage", ".png")) {
                                    ImageIO.write(character.getDescription().getPortrait().getRetina(), "png", image);
                                    path = String.format("api/SaveCharacterImage/%s/%s", id, key);
                                    result = client.sendRequest(HttpMethodType.POST, path, Files.readAllBytes(image.toPath()));
                                    if (!result.isEmpty()) {
                                        throw new IOException("Bad response from the web server for image write");
                                    }
                                }
                                try (ByteArrayOutputStream out = new ByteArrayOutputStream()) {
                                    try (XMLWriter w = new XMLWriter(out)) {
                                        character.save(w, true, false);
                                    }
                                    path = String.format("api/SaveCharacterRawFileGCS/%s/%s", id, key);
                                    result = client.sendRequest(HttpMethodType.POST, path, out.toByteArray());
                                    if (!result.isEmpty()) {
                                        throw new IOException("Bad response from the web server for GCS file write");
                                    }
                                }
                                showResult(true);
                            } else {
                                showResult(false);
                            }
                        }
                    }
                }
            } catch (Exception exception) {
                Log.error(exception);
                showResult(false);
            }
        }
    }

    private static boolean showExistsDialogIfNecessary(WebServiceClient client, GURPSCharacter character) throws IOException, NotImplementedException {
        if ("true".equals(client.sendRequest(HttpMethodType.GET, String.format("api/GetCharacterExists/%s/%s", character.getId(), OutputPreferences.getGurpsCalculatorKey())))) {
            switch (JOptionPane.showOptionDialog(KeyboardFocusManager.getCurrentKeyboardFocusManager().getPermanentFocusOwner(), I18n.Text("This character already exists in GURPS Calculator.\nWould you like to replace it?\n\nIf you choose 'Create New', you should save your\ncharacter afterwards."), I18n.Text("Character Exists"), JOptionPane.YES_NO_CANCEL_OPTION, JOptionPane.QUESTION_MESSAGE, null, new Object[]{I18n.Text("Replace"), I18n.Text("Create New"), I18n.Text("Cancel")}, I18n.Text("Cancel"))) {
            case 1:
                character.generateNewId();
                character.setModified(true);
                break;
            case 2:
                return false;
            default:
                break;

            }
        }
        return true;
    }

    private static void showResult(boolean success) {
        String message = success ? I18n.Text("Export to GURPS Calculator was successful.") : I18n.Text("There was an error exporting to GURPS Calculator. Please try again later.");
        String key     = OutputPreferences.getGurpsCalculatorKey();
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
}
