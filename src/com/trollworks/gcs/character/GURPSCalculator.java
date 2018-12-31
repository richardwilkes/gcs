/*
 * Copyright (c) 1998-2018 by Richard A. Wilkes. All rights reserved.
 *
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, version 2.0. If a copy of the MPL was not distributed with
 * this file, You can obtain one at http://mozilla.org/MPL/2.0/.
 *
 * This Source Code Form is "Incompatible With Secondary Licenses", as
 * defined by the Mozilla Public License, version 2.0.
 */

package com.trollworks.gcs.character;

import com.trollworks.gcs.common.TemporaryFile;
import com.trollworks.gcs.preferences.OutputPreferences;
import com.trollworks.gcs.services.HttpMethodType;
import com.trollworks.gcs.services.NotImplementedException;
import com.trollworks.gcs.services.WebServiceClient;
import com.trollworks.toolkit.annotation.Localize;
import com.trollworks.toolkit.io.Log;
import com.trollworks.toolkit.io.xml.XMLWriter;
import com.trollworks.toolkit.ui.image.AnnotatedImage;
import com.trollworks.toolkit.ui.menu.Command;
import com.trollworks.toolkit.ui.widget.WindowUtils;
import com.trollworks.toolkit.utility.Localization;

import java.awt.Color;
import java.awt.Desktop;
import java.awt.Font;
import java.awt.KeyboardFocusManager;
import java.io.ByteArrayOutputStream;
import java.io.FileNotFoundException;
import java.io.IOException;
import java.io.PrintWriter;
import java.net.MalformedURLException;
import java.net.URISyntaxException;
import java.net.URL;
import java.nio.file.Files;
import java.text.MessageFormat;
import java.util.Scanner;
import java.util.UUID;

import javax.swing.JEditorPane;
import javax.swing.JLabel;
import javax.swing.JOptionPane;
import javax.swing.event.HyperlinkEvent;
import javax.swing.event.HyperlinkListener;

/** Provides export to GURPS Calculator. */
public class GURPSCalculator {
    @Localize("Replace")
    private static String   OPTION_REPLACE;
    @Localize("Create New")
    private static String   OPTION_CREATE_NEW;
    @Localize("Cancel")
    private static String   OPTION_CANCEL;
    @Localize("Character Exists")
    private static String   TITLE_CHARACTER_EXISTS;
    @Localize("This character already exists in GURPS Calculator.\nWould you like to replace it?\n\nIf you choose 'Create New', you should save your\ncharacter afterwards.")
    private static String   TEXT_CHARACTER_EXISTS;
    @Localize("Success")
    private static String   SUCCESS_TITLE;
    @Localize("Export to GURPS Calculator was successful.")
    private static String   SUCCESS_MESSAGE;
    @Localize("There was an error exporting to GURPS Calculator. Please try again later.")
    private static String   ERROR_MESSAGE;
    @Localize("You need to set a valid GURPS Calculator Key in sheet preferences.<br><a href='%s'>Click here</a> for more information.")
    private static String   KEY_MISSING_MESSAGE;
    @Localize("Unable to open {0}")
    protected static String UNABLE_TO_OPEN_URL;

    static {
        Localization.initialize();
    }

    public static void export(CharacterSheet sheet) {
        if (sheet != null) {
            GURPSCharacter   character = sheet.getCharacter();
            WebServiceClient client    = new WebServiceClient(OutputPreferences.BASE_GURPS_CALCULATOR_URL);
            try {
                if (showExistsDialogIfNecessary(client, character)) {
                    try (TemporaryFile templateFile = new TemporaryFile("gcalcTemplate", ".html")) { //$NON-NLS-1$ //$NON-NLS-2$
                        try (PrintWriter out = new PrintWriter(templateFile)) {
                            out.print(client.sendRequest(HttpMethodType.GET, "api/GetOutputTemplate")); //$NON-NLS-1$
                        }
                        try (TemporaryFile outputFile = new TemporaryFile("gcalcOutput", ".html")) { //$NON-NLS-1$ //$NON-NLS-2$
                            if (new TextTemplate(sheet).export(outputFile, templateFile)) {
                                String result = null;
                                try (Scanner scanner = new Scanner(outputFile)) {
                                    result = scanner.useDelimiter("\\A").next(); //$NON-NLS-1$
                                } catch (FileNotFoundException exception) {
                                    Log.error(exception);
                                }
                                UUID   id   = character.getId();
                                String key  = OutputPreferences.getGurpsCalculatorKey();
                                String path = String.format("api/SaveCharacter/%s/%s", id, key); //$NON-NLS-1$
                                result = client.sendRequest(HttpMethodType.POST, path, null, result);
                                if (!result.isEmpty()) {
                                    throw new IOException("Bad response from the web server for template write"); //$NON-NLS-1$
                                }
                                try (TemporaryFile image = new TemporaryFile("gcalcImage", ".png")) { //$NON-NLS-1$ //$NON-NLS-2$
                                    AnnotatedImage.writePNG(image, character.getDescription().getPortrait().getRetina(), 150, null);
                                    path   = String.format("api/SaveCharacterImage/%s/%s", id, key); //$NON-NLS-1$
                                    result = client.sendRequest(HttpMethodType.POST, path, Files.readAllBytes(image.toPath()));
                                    if (!result.isEmpty()) {
                                        throw new IOException("Bad response from the web server for image write"); //$NON-NLS-1$
                                    }
                                }
                                try (ByteArrayOutputStream out = new ByteArrayOutputStream()) {
                                    try (XMLWriter w = new XMLWriter(out)) {
                                        character.save(w, true, false);
                                    }
                                    path   = String.format("api/SaveCharacterRawFileGCS/%s/%s", id, key); //$NON-NLS-1$
                                    result = client.sendRequest(HttpMethodType.POST, path, out.toByteArray());
                                    if (!result.isEmpty()) {
                                        throw new IOException("Bad response from the web server for GCS file write"); //$NON-NLS-1$
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

    private static boolean showExistsDialogIfNecessary(WebServiceClient client, GURPSCharacter character) throws MalformedURLException, IOException, NotImplementedException {
        if (client.sendRequest(HttpMethodType.GET, String.format("api/GetCharacterExists/%s/%s", character.getId(), OutputPreferences.getGurpsCalculatorKey())).equals("true")) { //$NON-NLS-1$ //$NON-NLS-2$
            switch (JOptionPane.showOptionDialog(KeyboardFocusManager.getCurrentKeyboardFocusManager().getPermanentFocusOwner(), TEXT_CHARACTER_EXISTS, TITLE_CHARACTER_EXISTS, JOptionPane.YES_NO_CANCEL_OPTION, JOptionPane.QUESTION_MESSAGE, null, new Object[] { OPTION_REPLACE, OPTION_CREATE_NEW, OPTION_CANCEL }, OPTION_CANCEL)) {
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
        String message = success ? SUCCESS_MESSAGE : ERROR_MESSAGE;
        String key     = OutputPreferences.getGurpsCalculatorKey();
        if (key == null || !key.matches("[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[1-5][0-9a-fA-F]{3}-[89ab][0-9a-fA-F]{3}-[0-9a-fA-F]{12}")) { //$NON-NLS-1$
            message = String.format(KEY_MISSING_MESSAGE, OutputPreferences.GURPS_CALCULATOR_URL);
        }
        JLabel        styleLabel = new JLabel();
        Font          font       = styleLabel.getFont();
        Color         color      = styleLabel.getBackground();
        StringBuilder buffer     = new StringBuilder();
        buffer.append("<html><body style='font-family:"); //$NON-NLS-1$
        buffer.append(font.getFamily());
        buffer.append(";font-weight:"); //$NON-NLS-1$
        buffer.append(font.isBold() ? "bold" : "normal"); //$NON-NLS-1$ //$NON-NLS-2$
        buffer.append(";font-size:"); //$NON-NLS-1$
        buffer.append(font.getSize());
        buffer.append("pt;background-color: rgb("); //$NON-NLS-1$
        buffer.append(color.getRed());
        buffer.append(","); //$NON-NLS-1$
        buffer.append(color.getGreen());
        buffer.append(","); //$NON-NLS-1$
        buffer.append(color.getBlue());
        buffer.append(");'>"); //$NON-NLS-1$
        buffer.append(message);
        buffer.append("</body></html>"); //$NON-NLS-1$
        JEditorPane messagePane = new JEditorPane("text/html", buffer.toString()); //$NON-NLS-1$
        messagePane.setEditable(false);
        messagePane.setBorder(null);
        messagePane.addHyperlinkListener(new HyperlinkListener() {
            @Override
            public void hyperlinkUpdate(HyperlinkEvent event) {
                if (Desktop.isDesktopSupported() && event.getEventType().equals(HyperlinkEvent.EventType.ACTIVATED)) {
                    URL url = event.getURL();
                    try {
                        Desktop.getDesktop().browse(url.toURI());
                    } catch (IOException | URISyntaxException exception) {
                        WindowUtils.showError(null, MessageFormat.format(UNABLE_TO_OPEN_URL, url.toExternalForm()));
                    }
                }
            }
        });
        JOptionPane.showMessageDialog(Command.getFocusOwner(), messagePane, success ? SUCCESS_TITLE : WindowUtils.ERROR, success ? JOptionPane.INFORMATION_MESSAGE : JOptionPane.ERROR_MESSAGE);
    }
}
