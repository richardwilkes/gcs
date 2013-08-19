@echo off
REM ***** BEGIN LICENSE BLOCK *****
REM Version: MPL 1.1
REM
REM The contents of this file are subject to the Mozilla Public License Version
REM 1.1 (the "License"); you may not use this file except in compliance with
REM the License. You may obtain a copy of the License at
REM http://www.mozilla.org/MPL/
REM
REM Software distributed under the License is distributed on an "AS IS" basis,
REM WITHOUT WARRANTY OF ANY KIND, either express or implied. See the License
REM for the specific language governing rights and limitations under the
REM License.
REM
REM The Original Code is GURPS Character Sheet.
REM
REM The Initial Developer of the Original Code is Richard A. Wilkes.
REM Portions created by the Initial Developer are Copyright (C) 1998-2002,
REM 2005-2011 the Initial Developer. All Rights Reserved.
REM
REM Contributor(s):
REM
REM ***** END LICENSE BLOCK *****

start javaw -Xmx256M -jar "%~p0GURPS Character Sheet.app\Contents\Resources\Java\GCS.jar" %*
