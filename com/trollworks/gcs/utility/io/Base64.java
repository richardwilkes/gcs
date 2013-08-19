/* ***** BEGIN LICENSE BLOCK *****
 * Version: MPL 1.1
 *
 * The contents of this file are subject to the Mozilla Public License Version
 * 1.1 (the "License"); you may not use this file except in compliance with
 * the License. You may obtain a copy of the License at
 * http://www.mozilla.org/MPL/
 *
 * Software distributed under the License is distributed on an "AS IS" basis,
 * WITHOUT WARRANTY OF ANY KIND, either express or implied. See the License
 * for the specific language governing rights and limitations under the
 * License.
 *
 * The Original Code is GURPS Character Sheet.
 *
 * The Initial Developer of the Original Code is Richard A. Wilkes.
 * Portions created by the Initial Developer are Copyright (C) 1998-2002,
 * 2005-2007 the Initial Developer. All Rights Reserved.
 *
 * Contributor(s):
 *
 * ***** END LICENSE BLOCK ***** */

package com.trollworks.gcs.utility.io;

import java.io.UnsupportedEncodingException;

/** Encodes and decodes to and from Base64 notation. */
public class Base64 {
	private static final String	PREFERRED_ENCODING	= "UTF-8";																																																																																																																																																																																																																																																																																		//$NON-NLS-1$
	private static final int	MAX_LINE_LENGTH		= 76;
	private static final byte	EQUALS_SIGN			= (byte) '=';
	private static final byte	NEW_LINE			= (byte) '\n';
	private static final byte	WHITESPACE_ENCODING	= -5;
	private static final byte	EQUALS_ENCODING		= -1;
	private static final byte[]	ALPHABET;
	private static final byte[]	NATIVE_ALPHABET		= { (byte) 'A', (byte) 'B', (byte) 'C', (byte) 'D', (byte) 'E', (byte) 'F', (byte) 'G', (byte) 'H', (byte) 'I', (byte) 'J', (byte) 'K', (byte) 'L', (byte) 'M', (byte) 'N', (byte) 'O', (byte) 'P', (byte) 'Q', (byte) 'R', (byte) 'S', (byte) 'T', (byte) 'U', (byte) 'V', (byte) 'W', (byte) 'X', (byte) 'Y', (byte) 'Z', (byte) 'a', (byte) 'b', (byte) 'c', (byte) 'd', (byte) 'e', (byte) 'f', (byte) 'g', (byte) 'h', (byte) 'i', (byte) 'j', (byte) 'k', (byte) 'l', (byte) 'm', (byte) 'n', (byte) 'o', (byte) 'p', (byte) 'q', (byte) 'r', (byte) 's', (byte) 't', (byte) 'u', (byte) 'v', (byte) 'w', (byte) 'x', (byte) 'y', (byte) 'z', (byte) '0', (byte) '1', (byte) '2', (byte) '3', (byte) '4', (byte) '5', (byte) '6', (byte) '7', (byte) '8', (byte) '9', (byte) '+', (byte) '/' };
	private static final byte[]	DECODE_ALPHABET		= { -9, -9, -9, -9, -9, -9, -9, -9, -9, WHITESPACE_ENCODING, WHITESPACE_ENCODING, -9, -9, WHITESPACE_ENCODING, -9, -9, -9, -9, -9, -9, -9, -9, -9, -9, -9, -9, -9, -9, -9, -9, -9, -9, WHITESPACE_ENCODING, -9, -9, -9, -9, -9, -9, -9, -9, -9, -9, 62, -9, -9, -9, 63, 52, 53, 54, 55, 56, 57, 58, 59, 60, 61, -9, -9, -9, EQUALS_ENCODING, -9, -9, -9, 0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, -9, -9, -9, -9, -9, -9, 26, 27, 28, 29, 30, 31, 32, 33, 34, 35, 36, 37, 38, 39, 40, 41, 42, 43, 44, 45, 46, 47, 48, 49, 50, 51, -9, -9, -9, -9, -9, -9, -9, -9, -9, -9, -9, -9, -9, -9, -9, -9, -9, -9, -9, -9, -9, -9, -9, -9, -9, -9, -9, -9, -9, -9, -9, -9, -9, -9, -9, -9, -9, -9, -9, -9, -9, -9, -9, -9, -9, -9, -9, -9, -9, -9, -9, -9, -9, -9, -9, -9, -9, -9, -9, -9, -9, -9, -9, -9, -9, -9, -9, -9, -9, -9, -9, -9, -9, -9, -9, -9, -9, -9, -9, -9, -9, -9, -9, -9, -9, -9, -9, -9, -9, -9, -9, -9, -9, -9, -9, -9, -9, -9, -9, -9, -9, -9, -9, -9, -9, -9, -9, -9, -9, -9, -9, -9, -9, -9, -9, -9, -9, -9, -9, -9, -9, -9, -9, -9, -9, -9, -9, -9, -9, -9, -9, -9, -9 };

	static {
		byte[] bytes;

		try {
			bytes = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789+/".getBytes(PREFERRED_ENCODING); //$NON-NLS-1$
		} catch (java.io.UnsupportedEncodingException use) {
			bytes = NATIVE_ALPHABET;
		}
		ALPHABET = bytes;
	}

	/**
	 * Encodes a byte array into Base64 notation.
	 * 
	 * @param source The data to convert.
	 * @return The encoded string.
	 */
	public static String encode(byte[] source) {
		return encode(source, 0, source.length, true);
	}

	/**
	 * Encodes a byte array into Base64 notation.
	 * 
	 * @param source The data to convert.
	 * @param breakLines Whether or not the lines should be broken into {@link #MAX_LINE_LENGTH}
	 *            chunks.
	 * @return The encoded string.
	 */
	public static String encode(byte[] source, boolean breakLines) {
		return encode(source, 0, source.length, breakLines);
	}

	/**
	 * Encodes a byte array into Base64 notation.
	 * 
	 * @param source The data to convert.
	 * @param off Offset in array where conversion should begin.
	 * @param len Length of data to convert.
	 * @param breakLines Whether or not the lines should be broken into {@link #MAX_LINE_LENGTH}
	 *            chunks.
	 * @return The encoded string.
	 */
	public static String encode(byte[] source, int off, int len, boolean breakLines) {
		int len43 = len * 4 / 3;
		byte[] outBuff = new byte[len43 + (len % 3 > 0 ? 4 : 0) + (breakLines ? len43 / MAX_LINE_LENGTH : 0)];
		int len2 = len - 2;
		int lineLength = 0;
		int i;
		int end;

		for (i = end = 0; i < len2; i += 3, end += 4) {
			encode3to4(source, i + off, 3, outBuff, end);
			lineLength += 4;
			if (breakLines && lineLength == MAX_LINE_LENGTH) {
				outBuff[end++ + 4] = NEW_LINE;
				lineLength = 0;
			}
		}

		if (i < len) {
			encode3to4(source, i + off, len - i, outBuff, end);
			end += 4;
		}

		try {
			return new String(outBuff, 0, end, PREFERRED_ENCODING);
		} catch (UnsupportedEncodingException uue) {
			return new String(outBuff, 0, end);
		}
	}

	private static void encode3to4(byte[] source, int srcOffset, int numSigBytes, byte[] destination, int destOffset) {
		int inBuff = (numSigBytes > 0 ? source[srcOffset] << 24 >>> 8 : 0) | (numSigBytes > 1 ? source[srcOffset + 1] << 24 >>> 16 : 0) | (numSigBytes > 2 ? source[srcOffset + 2] << 24 >>> 24 : 0);

		switch (numSigBytes) {
			case 3:
				destination[destOffset++] = ALPHABET[(inBuff >>> 18)];
				destination[destOffset++] = ALPHABET[inBuff >>> 12 & 0x3f];
				destination[destOffset++] = ALPHABET[inBuff >>> 6 & 0x3f];
				destination[destOffset] = ALPHABET[inBuff & 0x3f];
				break;
			case 2:
				destination[destOffset++] = ALPHABET[(inBuff >>> 18)];
				destination[destOffset++] = ALPHABET[inBuff >>> 12 & 0x3f];
				destination[destOffset++] = ALPHABET[inBuff >>> 6 & 0x3f];
				destination[destOffset] = EQUALS_SIGN;
				break;
			case 1:
				destination[destOffset++] = ALPHABET[(inBuff >>> 18)];
				destination[destOffset++] = ALPHABET[inBuff >>> 12 & 0x3f];
				destination[destOffset++] = EQUALS_SIGN;
				destination[destOffset] = EQUALS_SIGN;
				break;
		}
	}

	/**
	 * Decodes data from Base64 notation.
	 * 
	 * @param buffer The string to decode.
	 * @return The decoded data.
	 */
	public static byte[] decode(String buffer) {
		int outBuffPosn = 0;
		byte[] b4 = new byte[4];
		int b4Posn = 0;
		int i = 0;
		byte sbiCrop = 0;
		byte sbiDecode = 0;
		byte[] out;
		byte[] source;
		byte[] outBuff;
		int len;

		try {
			source = buffer.getBytes(PREFERRED_ENCODING);
		} catch (UnsupportedEncodingException uee) {
			source = buffer.getBytes();
		}
		len = source.length;
		outBuff = new byte[len * 3 / 4];

		for (i = 0; i < len; i++) {
			sbiCrop = (byte) (source[i] & 0x7f);
			sbiDecode = DECODE_ALPHABET[sbiCrop];
			if (sbiDecode >= WHITESPACE_ENCODING) {
				if (sbiDecode >= EQUALS_ENCODING) {
					b4[b4Posn++] = sbiCrop;
					if (b4Posn > 3) {
						int chunk;

						if (b4[2] == EQUALS_SIGN) {
							chunk = (DECODE_ALPHABET[b4[0]] & 0xFF) << 18 | (DECODE_ALPHABET[b4[1]] & 0xFF) << 12;
							outBuff[outBuffPosn++] = (byte) (chunk >>> 16);
						} else if (b4[3] == EQUALS_SIGN) {
							chunk = (DECODE_ALPHABET[b4[0]] & 0xFF) << 18 | (DECODE_ALPHABET[b4[1]] & 0xFF) << 12 | (DECODE_ALPHABET[b4[2]] & 0xFF) << 6;
							outBuff[outBuffPosn++] = (byte) (chunk >>> 16);
							outBuff[outBuffPosn++] = (byte) (chunk >>> 8);
						} else {
							chunk = (DECODE_ALPHABET[b4[0]] & 0xFF) << 18 | (DECODE_ALPHABET[b4[1]] & 0xFF) << 12 | (DECODE_ALPHABET[b4[2]] & 0xFF) << 6 | DECODE_ALPHABET[b4[3]] & 0xFF;
							outBuff[outBuffPosn++] = (byte) (chunk >> 16);
							outBuff[outBuffPosn++] = (byte) (chunk >> 8);
							outBuff[outBuffPosn++] = (byte) chunk;
						}
						b4Posn = 0;
						if (sbiCrop == EQUALS_SIGN) {
							break;
						}
					}
				}
			} else {
				return null;
			}
		}

		out = new byte[outBuffPosn];
		System.arraycopy(outBuff, 0, out, 0, outBuffPosn);
		return out;
	}
}
