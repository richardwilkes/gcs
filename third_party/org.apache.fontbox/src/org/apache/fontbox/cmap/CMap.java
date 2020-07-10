/*
 * Licensed to the Apache Software Foundation (ASF) under one or more
 * contributor license agreements.  See the NOTICE file distributed with
 * this work for additional information regarding copyright ownership.
 * The ASF licenses this file to You under the Apache License, Version 2.0
 * (the "License"); you may not use this file except in compliance with
 * the License.  You may obtain a copy of the License at
 *
 *      http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */
package org.apache.fontbox.cmap;

import java.io.IOException;
import java.io.InputStream;
import java.util.ArrayList;
import java.util.HashMap;
import java.util.List;
import java.util.Map;
import org.apache.commons.logging.Log;
import org.apache.commons.logging.LogFactory;

/**
 * This class represents a CMap file.
 *
 * @author Ben Litchfield
 */
public class CMap
{
    private static final Log LOG = LogFactory.getLog(CMap.class);

    private int wmode = 0;
    private String cmapName = null;
    private String cmapVersion = null;
    private int cmapType = -1;

    private String registry = null;
    private String ordering = null;
    private int supplement = 0;

    private int minCodeLength = 4;
    private int maxCodeLength;

    // code lengths
    private final List<CodespaceRange> codespaceRanges = new ArrayList<CodespaceRange>();

    // Unicode mappings
    private final Map<Integer,String> charToUnicode = new HashMap<Integer,String>();

    // CID mappings
    private final Map<Integer,Integer> codeToCid = new HashMap<Integer,Integer>();
    private final List<CIDRange> codeToCidRanges = new ArrayList<CIDRange>();

    private static final String SPACE = " ";
    private int spaceMapping = -1;

    /**
     * Creates a new instance of CMap.
     */
    CMap()
    {
    }

    /**
     * This will tell if this cmap has any CID mappings.
     * 
     * @return true If there are any CID mappings, false otherwise.
     */
    public boolean hasCIDMappings()
    {
        return !codeToCid.isEmpty() || !codeToCidRanges.isEmpty();
    }

    /**
     * This will tell if this cmap has any Unicode mappings.
     *
     * @return true If there are any Unicode mappings, false otherwise.
     */
    public boolean hasUnicodeMappings()
    {
        return !charToUnicode.isEmpty();
    }

    /**
     * Returns the sequence of Unicode characters for the given character code.
     *
     * @param code character code
     * @return Unicode characters (may be more than one, e.g "fi" ligature)
     */
    public String toUnicode(int code)
    {
        return charToUnicode.get(code);
    }

    /**
     * Reads a character code from a string in the content stream.
     * <p>See "CMap Mapping" and "Handling Undefined Characters" in PDF32000 for more details.
     *
     * @param in string stream
     * @return character code
     * @throws IOException if there was an error reading the stream or CMap
     */
    public int readCode(InputStream in) throws IOException
    {
        byte[] bytes = new byte[maxCodeLength];
        in.read(bytes,0,minCodeLength);
        in.mark(maxCodeLength);
        for (int i = minCodeLength-1; i < maxCodeLength; i++)
        {
            final int byteCount = i+1;
            for (CodespaceRange range : codespaceRanges)
            {
                if (range.isFullMatch(bytes, byteCount))
                {
                    return toInt(bytes, byteCount);
                }
            }
            if (byteCount < maxCodeLength)
            {
                bytes[byteCount] = (byte)in.read();
            }
        }
        String seq = "";
        for (int i = 0; i < maxCodeLength; ++i)
        {
            seq += String.format("0x%02X (%04o) ", bytes[i], bytes[i]);
        }
        LOG.warn("Invalid character code sequence " + seq + "in CMap " + cmapName);
        // PDFBOX-4811 reposition to where we were after initial read
        if (in.markSupported())
        {
            in.reset();
        }
        else
        {
            LOG.warn("mark() and reset() not supported, " + (maxCodeLength - 1) +
                     " bytes have been skipped");
        }
        return toInt(bytes, minCodeLength); // Adobe Reader behavior
    }

    /**
     * Returns an int for the given byte array
     */
    static int toInt(byte[] data, int dataLen)
    {
        int code = 0;
        for (int i = 0; i < dataLen; ++i)
        {
            code <<= 8;
            code |= (data[i] & 0xFF);
        }
        return code;
    }

    /**
     * Returns the CID for the given character code.
     *
     * @param code character code
     * @return CID
     */
    public int toCID(int code)
    {
        Integer cid = codeToCid.get(code);
        if (cid != null)
        {
            return cid;
        }
        for (CIDRange range : codeToCidRanges)
        {
            int ch = range.map((char)code);
            if (ch != -1)
            {
                return ch;
            }
        }
        return 0;
    }

    /**
     * Convert the given part of a byte array to an integer.
     * @param data the byte array
     * @param offset The offset into the byte array.
     * @param length The length of the data we are getting.
     * @return the resulting integer
     */
    private int getCodeFromArray( byte[] data, int offset, int length )
    {
        int code = 0;
        for( int i=0; i<length; i++ )
        {
            code <<= 8;
            code |= (data[offset+i]+256)%256;
        }
        return code;
    }

    /**
     * This will add a character code to Unicode character sequence mapping.
     *
     * @param codes The character codes to map from.
     * @param unicode The Unicode characters to map to.
     */
    void addCharMapping(byte[] codes, String unicode)
    {
        int code = getCodeFromArray(codes, 0, codes.length);
        charToUnicode.put(code, unicode);

        // fixme: ugly little hack
        if (SPACE.equals(unicode))
        {
            spaceMapping = code;
        }
    }

    /**
     * This will add a CID mapping.
     *
     * @param code character code
     * @param cid CID
     */
    void addCIDMapping(int code, int cid)
    {
        codeToCid.put(cid, code);
    }

    /**
     * This will add a CID Range.
     *
     * @param from starting character of the CID range.
     * @param to ending character of the CID range.
     * @param cid the cid to be started with.
     *
     */
    void addCIDRange(char from, char to, int cid)
    {
        CIDRange lastRange = null;
        if (!codeToCidRanges.isEmpty())
        {
            lastRange = codeToCidRanges.get(codeToCidRanges.size() - 1);
        }
        if (lastRange == null || !lastRange.extend(from, to, cid))
        {
            codeToCidRanges.add(new CIDRange(from, to, cid));
        }
    }

    /**
     * This will add a codespace range.
     *
     * @param range A single codespace range.
     */
    void addCodespaceRange( CodespaceRange range )
    {
        codespaceRanges.add(range);
        maxCodeLength = Math.max(maxCodeLength, range.getCodeLength());
        minCodeLength = Math.min(minCodeLength, range.getCodeLength());
    }
    
    /**
     * Implementation of the usecmap operator.  This will
     * copy all of the mappings from one cmap to another.
     * 
     * @param cmap The cmap to load mappings from.
     */
    void useCmap( CMap cmap )
    {
        for (CodespaceRange codespaceRange : cmap.codespaceRanges)
        {
            addCodespaceRange(codespaceRange);
        }
        charToUnicode.putAll(cmap.charToUnicode);
        codeToCid.putAll(cmap.codeToCid);
        codeToCidRanges.addAll(cmap.codeToCidRanges);
    }

    /**
     * Returns the WMode of a CMap.
     *
     * 0 represents a horizontal and 1 represents a vertical orientation.
     * 
     * @return the wmode
     */
    public int getWMode() 
    {
        return wmode;
    }

    /**
     * Sets the WMode of a CMap.
     * 
     * @param newWMode the new WMode.
     */
    public void setWMode(int newWMode) 
    {
        wmode = newWMode;
    }

    /**
     * Returns the name of the CMap.
     * 
     * @return the CMap name.
     */
    public String getName() 
    {
        return cmapName;
    }

    /**
     * Sets the name of the CMap.
     * 
     * @param name the CMap name.
     */
    public void setName(String name) 
    {
        cmapName = name;
    }

    /**
     * Returns the version of the CMap.
     * 
     * @return the CMap version.
     */
    public String getVersion() 
    {
        return cmapVersion;
    }

    /**
     * Sets the version of the CMap.
     * 
     * @param version the CMap version.
     */
    public void setVersion(String version) 
    {
        cmapVersion = version;
    }

    /**
     * Returns the type of the CMap.
     * 
     * @return the CMap type.
     */
    public int getType() 
    {
        return cmapType;
    }

    /**
     * Sets the type of the CMap.
     * 
     * @param type the CMap type.
     */
    public void setType(int type) 
    {
        cmapType = type;
    }

    /**
     * Returns the registry of the CIDSystemInfo.
     * 
     * @return the registry.
     */
    public String getRegistry() 
    {
        return registry;
    }

    /**
     * Sets the registry of the CIDSystemInfo.
     * 
     * @param newRegistry the registry.
     */
    public void setRegistry(String newRegistry) 
    {
        registry = newRegistry;
    }

    /**
     * Returns the ordering of the CIDSystemInfo.
     * 
     * @return the ordering.
     */
    public String getOrdering() 
    {
        return ordering;
    }

    /**
     * Sets the ordering of the CIDSystemInfo.
     * 
     * @param newOrdering the ordering.
     */
    public void setOrdering(String newOrdering) 
    {
        ordering = newOrdering;
    }

    /**
     * Returns the supplement of the CIDSystemInfo.
     * 
     * @return the supplement.
     */
    public int getSupplement() 
    {
        return supplement;
    }

    /**
     * Sets the supplement of the CIDSystemInfo.
     * 
     * @param newSupplement the supplement.
     */
    public void setSupplement(int newSupplement) 
    {
        supplement = newSupplement;
    }
    
    /** 
     * Returns the mapping for the space character.
     * 
     * @return the mapped code for the space character
     */
    public int getSpaceMapping()
    {
        return spaceMapping;
    }

    @Override
    public String toString()
    {
        return cmapName;
    }
}
