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

package bundler;

import java.io.BufferedReader;
import java.io.File;
import java.io.IOException;
import java.io.InputStream;
import java.io.InputStreamReader;
import java.io.OutputStream;
import java.io.PrintStream;
import java.io.PrintWriter;
import java.lang.ProcessBuilder.Redirect;
import java.nio.charset.StandardCharsets;
import java.nio.file.FileVisitOption;
import java.nio.file.FileVisitResult;
import java.nio.file.FileVisitor;
import java.nio.file.Files;
import java.nio.file.Path;
import java.nio.file.Paths;
import java.nio.file.attribute.BasicFileAttributes;
import java.time.Instant;
import java.time.ZoneId;
import java.time.format.DateTimeFormatter;
import java.util.ArrayList;
import java.util.Date;
import java.util.EnumSet;
import java.util.HashSet;
import java.util.List;
import java.util.Set;

public class Bundler {
    private static final String GCS_VERSION       = "4.20.0";
    private static       String JDK_MAJOR_VERSION = "14";
    private static final String ITEXT_VERSION     = "2.1.7";
    private static final String LOGGING_VERSION   = "1.2.0";
    private static final String FONTBOX_VERSION   = "2.0.20";
    private static final String PDFBOX_VERSION    = "2.0.20";
    private static final String LINUX             = "linux";
    private static final String MACOS             = "macos";
    private static final String WINDOWS           = "windows";
    private static final Path   DIST_DIR          = Paths.get("out", "dist");
    private static final Path   BUILD_DIR         = DIST_DIR.resolve("build");
    private static final Path   MODULE_DIR        = DIST_DIR.resolve("modules");
    private static final Path   EXTRA_DIR         = DIST_DIR.resolve("extra");
    private static final Path   I18N_DIR          = EXTRA_DIR.resolve("i18n");
    private static final Path   MANIFEST          = BUILD_DIR.resolve("com.trollworks.gcs.manifest");
    private static final Path   JRE               = BUILD_DIR.resolve("jre");
    private static final String YEARS             = "1998-" + DateTimeFormatter.ofPattern("yyyy").format(Instant.now().atZone(ZoneId.systemDefault()));
    private static final char[] HEX_DIGITS        = {'0', '1', '2', '3', '4', '5', '6', '7', '8', '9', 'a', 'b', 'c', 'd', 'e', 'f'};
    private static       String OS;
    private static       Path   PKG;
    private static       Path   JPACKAGE_15;
    private static       String ICON_TYPE;

    /**
     * The main entry point for bundling GCS.
     *
     * @param args Arguments to the program.
     */
    public static void main(String[] args) {
        checkPlatform();

        boolean sign     = false;
        boolean notarize = false;
        for (String arg : args) {
            if (MACOS.equals(OS)) {
                if ("-s".equals(arg) || "--sign".equals(arg)) {
                    if (!sign) {
                        sign = true;
                        System.out.println("Signing enabled");
                    }
                    continue;
                }
                if ("-n".equals(arg) || "--notarize".equals(arg)) {
                    if (!notarize) {
                        notarize = true;
                        System.out.println("Notarization enabled");
                    }
                    continue;
                }
            }
            if ("-h".equals(arg) || "--help".equals(arg)) {
                System.out.println("-h, --help      This help");
                System.out.println("-n, --notarize  Enable notarization of the application (macOS only)");
                System.out.println("-s, --sign      Enable signing of the application (macOS only)");
                System.exit(0);
            }
            System.out.println("Ignoring argument: " + arg);
        }

        checkJDK();
        prepareDirs();
        compile();
        copyResources();
        createModules();
        extractLocalizationTemplate();
        packageApp(sign);

        if (notarize) {
            notarizeApp();
        }

        System.out.println("Finished!");
        System.out.println();
        System.out.println("Package can be found at:");
        System.out.println(PKG.toAbsolutePath().toString());
    }

    private static void checkPlatform() {
        String osName = System.getProperty("os.name");
        if (osName.startsWith("Mac")) {
            OS = MACOS;
            PKG = Paths.get("GCS-" + GCS_VERSION + ".dmg");
            ICON_TYPE = "icns";
        } else if (osName.startsWith("Win")) {
            OS = WINDOWS;
            PKG = Paths.get("GCS-" + GCS_VERSION + ".msi");
            ICON_TYPE = "ico";
        } else if (osName.startsWith("Linux")) {
            OS = LINUX;
            PKG = Paths.get("gcs-" + GCS_VERSION + "-1_amd64.deb");
            ICON_TYPE = "png";
        } else {
            System.err.println("Unsupported platform: " + osName);
            System.exit(1);
        }
    }

    private static void checkJDK() {
        ProcessBuilder builder = new ProcessBuilder("javac", "--version");
        builder.redirectOutput(Redirect.PIPE).redirectErrorStream(true);
        try {
            String  versionLine = "";
            Process process     = builder.start();
            try (BufferedReader in = new BufferedReader(new InputStreamReader(process.getInputStream(), StandardCharsets.UTF_8))) {
                String prefix = "javac ";
                String line;
                while ((line = in.readLine()) != null) {
                    if (line.startsWith(prefix)) {
                        versionLine = line.substring(prefix.length());
                    }
                }
            }
            if (!versionLine.startsWith(JDK_MAJOR_VERSION)) {
                System.err.println("JDK " + versionLine + " was found. JDK " + JDK_MAJOR_VERSION + " is required.");
                emitInstallJDKMessageAndExit();
            }
        } catch (IOException exception) {
            System.err.println("JDK " + JDK_MAJOR_VERSION + " is not installed!");
            emitInstallJDKMessageAndExit();
        }

        if (OS.equals(MACOS)) {
            boolean failed = false;
            Path    dir    = Paths.get(System.getProperty("user.home", "."), "jdk-15.jdk").toAbsolutePath();
            JPACKAGE_15 = dir.resolve(Paths.get("Contents", "Home", "bin", "jpackage"));
            builder = new ProcessBuilder(JPACKAGE_15.toString(), "--version");
            builder.redirectOutput(Redirect.PIPE).redirectErrorStream(true);
            try {
                String  versionLine = "";
                Process process     = builder.start();
                try (BufferedReader in = new BufferedReader(new InputStreamReader(process.getInputStream(), StandardCharsets.UTF_8))) {
                    String prefix = "15";
                    String line;
                    while ((line = in.readLine()) != null) {
                        if (line.startsWith(prefix)) {
                            versionLine = line;
                        }
                    }
                }
                if (!versionLine.startsWith("15")) {
                    failed = true;
                }
            } catch (IOException exception) {
                failed = true;
            }
            if (failed) {
                System.err.println("jpackage 15 is not available!");
                System.err.println("Unpack JDK 15 from http://jdk.java.net/15/ into " + " and try again.");
                System.exit(1);
            }
        }
    }

    private static void emitInstallJDKMessageAndExit() {
        System.err.println("Install JDK " + JDK_MAJOR_VERSION + " from http://jdk.java.net/" + JDK_MAJOR_VERSION + "/ and try again.");
        System.exit(1);
    }

    private static void prepareDirs() {
        System.out.print("Removing any previous build data... ");
        System.out.flush();
        long timing = System.nanoTime();
        try {
            if (Files.exists(DIST_DIR)) {
                Files.walkFileTree(DIST_DIR, new RecursiveDirectoryRemover());
            }
            Files.createDirectories(DIST_DIR);
            Files.createDirectories(BUILD_DIR);
            Files.createDirectories(MODULE_DIR);
            Files.createDirectories(EXTRA_DIR);
            Files.createDirectories(I18N_DIR);
            Files.deleteIfExists(PKG);
        } catch (IOException exception) {
            System.out.println();
            exception.printStackTrace(System.err);
            System.exit(1);
        }
        showTiming(timing);
    }

    private static void showTiming(long timing) {
        System.out.println(String.format("%,.3fs", Double.valueOf((System.nanoTime() - timing) / 1000000000.0)));
    }

    private static void compile() {
        System.out.print("Compiling... ");
        System.out.flush();
        long timing     = System.nanoTime();
        Path javacInput = BUILD_DIR.resolve("javac.input");
        try (PrintWriter out = new PrintWriter(Files.newBufferedWriter(javacInput))) {
            out.println("-d");
            out.println(BUILD_DIR.toString());
            out.println("--release");
            out.println(JDK_MAJOR_VERSION);
            out.println("-encoding");
            out.println("UTF8");
            out.println("--module-source-path");
            out.printf(".%1$s*%1$ssrc%2$sthird_party%1$s*%1$ssrc\n", File.separator, File.pathSeparator);
            FileScanner.walk(Paths.get("."), (path) -> {
                if (path.getFileName().toString().endsWith(".java") && !path.startsWith(Paths.get(".", "bundler"))) {
                    out.println(path.toString());
                }
            });
        } catch (IOException exception) {
            System.out.println();
            exception.printStackTrace(System.err);
            System.exit(1);
        }
        runNoOutputCmd("javac", "@" + javacInput.toString());
        showTiming(timing);
    }

    private static void copyResources() {
        System.out.print("Copying resources... ");
        System.out.flush();
        long timing = System.nanoTime();
        copyResourceTree(Paths.get("com.trollworks.gcs", "resources"), BUILD_DIR.resolve("com.trollworks.gcs"));
        copyResourceTree(Paths.get("third_party", "org.apache.pdfbox", "resources"), BUILD_DIR.resolve("org.apache.pdfbox"));
        copyResourceTree(Paths.get("third_party", "org.apache.fontbox", "resources"), BUILD_DIR.resolve("org.apache.fontbox"));
        copyResourceTree(Paths.get("third_party", "com.lowagie.text", "resources"), BUILD_DIR.resolve("com.lowagie.text"));
        showTiming(timing);
    }

    private static void copyResourceTree(Path src, Path dst) {
        FileScanner.walk(src, (path) -> {
            Path target = dst.resolve(src.relativize(path));
            try {
                Files.createDirectories(target.getParent());
                try (InputStream in = Files.newInputStream(path)) {
                    try (OutputStream out = Files.newOutputStream(target)) {
                        byte[] data = new byte[8192];
                        int    amt;
                        while ((amt = in.read(data)) != -1) {
                            out.write(data, 0, amt);
                        }
                    }
                }
            } catch (IOException exception) {
                System.out.println();
                exception.printStackTrace(System.err);
                System.exit(1);
            }
        });
    }

    private static void createModules() {
        System.out.print("Creating modules... ");
        System.out.flush();
        long timing = System.nanoTime();
        createManifest();
        List<String> args = new ArrayList<>();
        args.add("jar");
        args.add("--create");
        args.add("--file");
        args.add(MODULE_DIR.resolve("com.trollworks.gcs-" + GCS_VERSION + ".jar").toString());
        args.add("--module-version");
        args.add(GCS_VERSION);
        args.add("--manifest");
        args.add(MANIFEST.toString());
        args.add("--main-class");
        args.add("com.trollworks.gcs.GCS");
        args.add("-C");
        args.add(BUILD_DIR.resolve("com.trollworks.gcs").toString());
        args.add(".");
        runNoOutputCmd(args);
        buildJar("com.lowagie.text", ITEXT_VERSION);
        buildJar("org.apache.commons.logging", LOGGING_VERSION);
        buildJar("org.apache.fontbox", FONTBOX_VERSION);
        buildJar("org.apache.pdfbox", PDFBOX_VERSION);
        showTiming(timing);
    }

    private static void createManifest() {
        try (PrintWriter out = new PrintWriter(Files.newBufferedWriter(MANIFEST))) {
            out.println("Manifest-Version: 1.0");
            out.println("bundle-name: GCS");
            out.println("bundle-version: " + GCS_VERSION);
            out.println("bundle-license: Mozilla Public License 2.0");
            out.println("bundle-copyright-owner: Richard A. Wilkes");
            out.println("bundle-copyright-years: " + YEARS);
        } catch (IOException exception) {
            System.out.println();
            exception.printStackTrace(System.err);
            System.exit(1);
        }
    }

    private static void buildJar(String pkg, String version) {
        List<String> args = new ArrayList<>();
        args.add("jar");
        args.add("--create");
        args.add("--file");
        args.add(MODULE_DIR.resolve(pkg + "-" + version + ".jar").toString());
        args.add("--module-version");
        args.add(version);
        args.add("-C");
        args.add(BUILD_DIR.resolve(pkg).toString());
        args.add(".");
        runNoOutputCmd(args);
    }

    private static void extractLocalizationTemplate() {
        System.out.print("Extracting localization template... ");
        System.out.flush();
        long        timing = System.nanoTime();
        Set<String> keys   = new HashSet<>();
        try {
            Files.walk(Paths.get("com.trollworks.gcs", "src")).filter(path -> {
                String lower = path.getFileName().toString().toLowerCase();
                return lower.endsWith(".java") && !lower.endsWith("i18n.java") && Files.isRegularFile(path) && Files.isReadable(path);
            }).distinct().forEach(path -> {
                try {
                    Files.lines(path).forEachOrdered(line -> {
                        while (!line.isEmpty()) {
                            int i = line.indexOf("I18n.Text(");
                            if (i < 0) {
                                break;
                            }
                            int max = line.length();
                            i += 10;
                            while (i < max) {
                                char ch = line.charAt(i);
                                if (ch != ' ' && ch != '\t') {
                                    break;
                                }
                                i++;
                            }
                            if (i >= max || line.charAt(i) != '"') {
                                break;
                            }
                            i++;
                            line = processLine(keys, line.substring(i));
                        }
                    });
                } catch (IOException ioe) {
                    System.out.println();
                    ioe.printStackTrace(System.err);
                    System.exit(1);
                }
            });
            try (PrintStream out = new PrintStream(Files.newOutputStream(I18N_DIR.resolve("template.i18n")), true, StandardCharsets.UTF_8)) {
                out.println("# Generated on " + new Date());
                out.println("#");
                out.println("# This file consists of UTF-8 text. Do not save it as anything else.");
                out.println("#");
                out.println("# Key-value pairs are defined as one or more lines prefixed with 'k:' for the");
                out.println("# key, followed by one or more lines prefixed with 'v:' for the value. These");
                out.println("# prefixes are then followed by a quoted string, as generated by Text.quote().");
                out.println("# When two or more lines are present in a row, they will be concatenated");
                out.println("# together with an intervening \\n character.");
                out.println("#");
                out.println("# Do NOT modify the 'k' values. They are the values as seen in the code.");
                out.println("#");
                out.println("# Replace the 'v' values with the appropriate translation.");
                keys.stream().sorted().forEachOrdered(key -> {
                    out.println();
                    String quoted = quote(key);
                    if (quoted.length() < 77) {
                        out.println("k:" + quoted);
                        out.println("v:" + quoted);
                    } else {
                        String[] parts = key.split("\n", -1);
                        for (String part : parts) {
                            out.println("k:" + quote(part));
                        }
                        for (String part : parts) {
                            out.println("v:" + quote(part));
                        }
                    }
                });
            }
        } catch (Exception ex) {
            System.out.println();
            ex.printStackTrace(System.err);
            System.exit(1);
        }
        showTiming(timing);
    }

    private static String processLine(Set<String> keys, String in) {
        StringBuilder buffer       = new StringBuilder();
        int           len          = in.length();
        int           state        = 0;
        int           unicodeValue = 0;
        for (int i = 0; i < len; i++) {
            char ch = in.charAt(i);
            switch (state) {
            case 0: // Looking for end quote
                if (ch == '"') {
                    keys.add(buffer.toString());
                    return in.substring(i + 1);
                }
                if (ch == '\\') {
                    state = 1;
                    continue;
                }
                buffer.append(ch);
                break;
            case 1: // Processing escape sequence
                switch (ch) {
                case 't':
                    buffer.append('\t');
                    state = 0;
                    break;
                case 'b':
                    buffer.append('\b');
                    state = 0;
                    break;
                case 'n':
                    buffer.append('\n');
                    state = 0;
                    break;
                case 'r':
                    buffer.append('\r');
                    state = 0;
                    break;
                case '"':
                    buffer.append('"');
                    state = 0;
                    break;
                case '\\':
                    buffer.append('\\');
                    state = 0;
                    break;
                case 'u':
                    state = 2;
                    unicodeValue = 0;
                    break;
                default:
                    System.out.println();
                    new RuntimeException("invalid escape sequence").printStackTrace(System.err);
                    System.exit(1);
                }
                break;
            case 2: // Processing first digit of unicode escape sequence
            case 3: // Processing second digit of unicode escape sequence
            case 4: // Processing third digit of unicode escape sequence
            case 5: // Processing fourth digit of unicode escape sequence
                if (!isHexDigit(ch)) {
                    System.out.println();
                    new RuntimeException("invalid unicode escape sequence").printStackTrace(System.err);
                    System.exit(1);
                }
                unicodeValue *= 16;
                unicodeValue += hexDigitValue(ch);
                if (state == 5) {
                    state = 0;
                    buffer.append((char) unicodeValue);
                } else {
                    state++;
                }
                break;
            default:
                System.out.println();
                new RuntimeException("invalid state").printStackTrace(System.err);
                System.exit(1);
                break;
            }
        }
        return "";
    }

    private static boolean isHexDigit(char ch) {
        return ch >= '0' && ch <= '9' || ch >= 'a' && ch <= 'f' || ch >= 'A' && ch <= 'F';
    }

    private static int hexDigitValue(char ch) {
        if (ch >= '0' && ch <= '9') {
            return ch - '0';
        }
        if (ch >= 'a' && ch <= 'f') {
            return 10 + ch - 'a';
        }
        if (ch >= 'A' && ch <= 'F') {
            return 10 + ch - 'A';
        }
        return 0;
    }

    private static String quote(String in) {
        StringBuilder buffer = new StringBuilder();
        int           length = in.length();
        buffer.append('"');
        for (int i = 0; i < length; i++) {
            char ch = in.charAt(i);
            if (ch == '"' || ch == '\\') {
                buffer.append('\\');
                buffer.append(ch);
            } else if (isPrintableChar(ch)) {
                buffer.append(ch);
            } else {
                switch (ch) {
                case '\b':
                    buffer.append("\\b");
                    break;
                case '\f':
                    buffer.append("\\f");
                    break;
                case '\n':
                    buffer.append("\\n");
                    break;
                case '\r':
                    buffer.append("\\r");
                    break;
                case '\t':
                    buffer.append("\\t");
                    break;
                default:
                    buffer.append("\\u");
                    buffer.append(HEX_DIGITS[ch >> 12 & 0xF]);
                    buffer.append(HEX_DIGITS[ch >> 8 & 0xF]);
                    buffer.append(HEX_DIGITS[ch >> 4 & 0xF]);
                    buffer.append(HEX_DIGITS[ch & 0xF]);
                    break;
                }
            }
        }
        buffer.append('"');
        return buffer.toString();
    }

    private static boolean isPrintableChar(char ch) {
        if (!Character.isISOControl(ch) && Character.isDefined(ch)) {
            try {
                Character.UnicodeBlock block = Character.UnicodeBlock.of(ch);
                return block != null && block != Character.UnicodeBlock.SPECIALS;
            } catch (Exception ex) {
                return false;
            }
        }
        return false;
    }

    private static void packageApp(boolean sign) {
        System.out.print("Packaging the application... ");
        System.out.flush();
        long         timing = System.nanoTime();
        List<String> args   = new ArrayList<>();
        args.add("jlink");
        args.add("--module-path");
        args.add(MODULE_DIR.toString());
        args.add("--output");
        args.add(JRE.toString());
        args.add("--compress=2");
        args.add("--no-header-files");
        args.add("--no-man-pages");
        args.add("--strip-debug");
        args.add("--strip-native-commands");
        args.add("--add-modules");
        args.add("com.trollworks.gcs");
        runNoOutputCmd("jlink", "--module-path", MODULE_DIR.toString(), "--output", JRE.toString(), "--compress=2", "--no-header-files", "--no-man-pages", "--strip-debug", "--strip-native-commands", "--add-modules", "com.trollworks.gcs");
        args.clear();
        if (OS.equals(MACOS)) {
            args.add(JPACKAGE_15.toString());
        } else {
            args.add("jpackage");
        }
        args.add("--module");
        args.add("com.trollworks.gcs/com.trollworks.gcs.GCS");
        args.add("--app-version");
        args.add(GCS_VERSION);
        args.add("--copyright");
        args.add("©" + YEARS + " by Richard A. Wilkes");
        args.add("--vendor");
        args.add("Richard A. Wilkes");
        args.add("--description");
        args.add("GCS (GURPS Character Sheet) is a stand-alone, interactive, character sheet editor that allows you to build characters for the GURPS 4th Edition roleplaying game system.");
        args.add("--license-file");
        args.add("LICENSE");
        args.add("--icon");
        args.add(Paths.get("artifacts", ICON_TYPE, "app." + ICON_TYPE).toString());
        for (String ext : new String[]{"adm", "adq", "eqm", "eqp", "gcs", "gct", "not", "skl", "spl"}) {
            args.add("--file-associations");
            args.add(Paths.get("artifacts", "file_associations", OS, ext + "_ext.properties").toString());
        }
        args.add("--input");
        args.add(EXTRA_DIR.toString());
        args.add("--runtime-image");
        args.add(JRE.toString());
        args.add("--java-options");
        args.add("-Dhttps.protocols=TLSv1.2,TLSv1.1,TLSv1");
        switch (OS) {
        case MACOS:
            args.add("--mac-package-name");
            args.add("GCS");
            args.add("--mac-package-identifier");
            args.add("com.trollworks.gcs");
            if (sign) {
                args.add("--mac-sign");
                args.add("--mac-signing-key-user-name");
                args.add("Richard Wilkes");
            }
            break;
        case LINUX:
            args.add("--linux-package-name");
            args.add("gcs");
            args.add("--linux-deb-maintainer");
            args.add("wilkes@me.com");
            args.add("--linux-menu-group");
            args.add("Roleplaying");
            args.add("--linux-app-category");
            args.add("Roleplaying");
            args.add("--linux-rpm-license-type");
            args.add("MPLv2.0");
            args.add("--linux-shortcut");
            args.add("--linux-app-release");
            args.add("1");
            args.add("--linux-package-deps");
            args.add("");
            break;
        case WINDOWS:
            args.add("--java-options");
            args.add("-Dsun.java2d.dpiaware=false");
            args.add("--win-menu");
            args.add("--win-menu-group");
            args.add("Roleplaying");
            args.add("--win-shortcut");
            args.add("--type");
            args.add("msi");
            args.add("--win-dir-chooser");
            args.add("--win-upgrade-uuid");
            args.add("E71F99DA-AD84-4E6E-9bE7-4E65421752E1");
            Path propsFile = BUILD_DIR.resolve("console.properties");
            try (PrintWriter out = new PrintWriter(Files.newBufferedWriter(propsFile))) {
                out.println("win-console=true");
            } catch (IOException exception) {
                System.out.println();
                exception.printStackTrace(System.err);
                System.exit(1);
            }
            args.add("--add-launcher");
            args.add("GCScmdline=" + propsFile.toString());
            break;
        }
        runNoOutputCmd(args);
        showTiming(timing);
    }

    private static void notarizeApp() {
        System.out.print("Notarizing the application... ");
        System.out.flush();
        long         timing = System.nanoTime();
        List<String> args   = new ArrayList<>();
        args.add("xcrun");
        args.add("altool");
        args.add("--notarize-app");
        args.add("--type");
        args.add("osx");
        args.add("--file");
        args.add(PKG.toAbsolutePath().toString());
        args.add("--primary-bundle-id");
        args.add("com.trollworks.gcs");
        args.add("--password");
        args.add("@keychain:gcs_app_pw");
        List<String> lines     = runCmd(args);
        String       requestID = null;
        boolean      noErrors  = false;
        for (String line : lines) {
            line = line.trim();
            if (line.startsWith("No errors uploading ")) {
                noErrors = true;
            } else if (line.startsWith("RequestUUID = ")) {
                String[] parts = line.split("=", 2);
                if (parts.length == 2) {
                    requestID = parts[1].trim();
                }
                if (noErrors) {
                    break;
                }
            }
        }
        if (!noErrors || requestID == null) {
            failWithLines("Unable to locate request ID from response. Response follows:", lines);
        }

        args.clear();
        args.add("xcrun");
        args.add("altool");
        args.add("--notarization-info");
        args.add(requestID);
        args.add("--password");
        args.add("@keychain:gcs_app_pw");
        boolean success = false;
        while (!success) {
            try {
                Thread.sleep(10000); // 10 seconds
            } catch (InterruptedException exception) {
                exception.printStackTrace();
            }
            lines = runCmd(args);
            for (String line : lines) {
                line = line.trim();
                if ("Status: invalid".equals(line)) {
                    failWithLines("Notarization failed. Response follows:", lines);
                }
                if ("Status: success".equals(line)) {
                    success = true;
                }
            }
            System.out.print(".");
            System.out.flush();
        }

        args.clear();
        args.add("xcrun");
        args.add("stapler");
        args.add("staple");
        args.add(PKG.toAbsolutePath().toString());
        success = false;
        for (String line : runCmd(args)) {
            line = line.trim();
            if ("The staple and validate action worked!".equals(line)) {
                success = true;
            }
        }
        if (!success) {
            failWithLines("Stapling failed. Response follows:", lines);
        }
        showTiming(timing);
    }

    private static void failWithLines(String msg, List<String> lines) {
        System.out.println();
        System.err.println(msg);
        System.err.println();
        for (String line : lines) {
            System.err.println(line);
        }
        System.exit(1);
    }

    private static List<String> runCmd(List<String> args) {
        List<String>   lines   = new ArrayList<>();
        ProcessBuilder builder = new ProcessBuilder(args);
        builder.redirectOutput(Redirect.PIPE).redirectErrorStream(true);
        try {
            boolean hadMsg  = false;
            Process process = builder.start();
            try (BufferedReader in = new BufferedReader(new InputStreamReader(process.getInputStream(), StandardCharsets.UTF_8))) {
                String line = in.readLine();
                while (line != null) {
                    lines.add(line);
                    line = in.readLine();
                }
            }
        } catch (IOException exception) {
            System.out.println();
            exception.printStackTrace(System.err);
            System.exit(1);
        }
        return lines;
    }

    private static void runNoOutputCmd(List<String> args) {
        runNoOutputCmd(args.toArray(new String[0]));
    }

    private static void runNoOutputCmd(String... args) {
        ProcessBuilder builder = new ProcessBuilder(args);
        builder.redirectOutput(Redirect.PIPE).redirectErrorStream(true);
        try {
            boolean hadMsg  = false;
            Process process = builder.start();
            try (BufferedReader in = new BufferedReader(new InputStreamReader(process.getInputStream(), StandardCharsets.UTF_8))) {
                String line = in.readLine();
                while (line != null) {
                    if (!line.startsWith("WARNING: Using incubator modules: jdk.incubator.jpackage")) {
                        if (!hadMsg) {
                            System.out.println();
                        }
                        System.err.println(line);
                        hadMsg = true;
                    }
                    line = in.readLine();
                }
            }
            if (hadMsg) {
                System.exit(1);
            }
        } catch (IOException exception) {
            System.out.println();
            exception.printStackTrace(System.err);
            System.exit(1);
        }
    }

    static class RecursiveDirectoryRemover implements FileVisitor<Path> {
        @Override
        public FileVisitResult preVisitDirectory(Path dir, BasicFileAttributes attrs) throws IOException {
            return FileVisitResult.CONTINUE;
        }

        @Override
        public FileVisitResult visitFile(Path file, BasicFileAttributes attrs) throws IOException {
            Files.delete(file);
            return FileVisitResult.CONTINUE;
        }

        @Override
        public FileVisitResult visitFileFailed(Path file, IOException exception) throws IOException {
            System.out.println();
            exception.printStackTrace(System.err);
            System.exit(1);
            return FileVisitResult.CONTINUE;
        }

        @Override
        public FileVisitResult postVisitDirectory(Path dir, IOException exception) throws IOException {
            if (exception != null) {
                System.out.println();
                exception.printStackTrace(System.err);
                System.exit(1);
            }
            if (!dir.equals(DIST_DIR)) {
                Files.delete(dir);
            }
            return FileVisitResult.CONTINUE;
        }
    }

    public interface Handler {
        void processFile(Path path) throws IOException;
    }

    static class FileScanner implements FileVisitor<Path> {
        private Path    mPath;
        private Handler mHandler;

        public static final void walk(Path path, Handler handler) {
            try {
                Files.walkFileTree(path, EnumSet.of(FileVisitOption.FOLLOW_LINKS), Integer.MAX_VALUE, new FileScanner(path, handler));
            } catch (Exception exception) {
                System.out.println();
                exception.printStackTrace(System.err);
                System.exit(1);
            }
        }

        private FileScanner(Path path, Handler handler) {
            mPath = path;
            mHandler = handler;
        }

        private boolean shouldSkip(Path path) {
            return !mPath.equals(path) && path.getFileName().toString().startsWith(".");
        }

        @Override
        public FileVisitResult preVisitDirectory(Path path, BasicFileAttributes attrs) throws IOException {
            if (shouldSkip(path)) {
                return FileVisitResult.SKIP_SUBTREE;
            }
            return FileVisitResult.CONTINUE;
        }

        @Override
        public FileVisitResult visitFile(Path path, BasicFileAttributes attrs) throws IOException {
            if (!shouldSkip(path)) {
                mHandler.processFile(path);
            }
            return FileVisitResult.CONTINUE;
        }

        @Override
        public FileVisitResult visitFileFailed(Path path, IOException exception) throws IOException {
            System.out.println();
            exception.printStackTrace(System.err);
            System.exit(1);
            return FileVisitResult.CONTINUE;
        }

        @Override
        public FileVisitResult postVisitDirectory(Path path, IOException exception) throws IOException {
            if (exception != null) {
                System.out.println();
                exception.printStackTrace(System.err);
                System.exit(1);
            }
            return FileVisitResult.CONTINUE;
        }
    }
}
