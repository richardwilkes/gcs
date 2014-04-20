@echo off

REM Copyright (c) 1998-2014 by Richard A. Wilkes. All rights reserved.
REM
REM This Source Code Form is subject to the terms of the Mozilla Public License,
REM version 2.0. If a copy of the MPL was not distributed with this file, You
REM can obtain one at http://mozilla.org/MPL/2.0/.
REM
REM This Source Code Form is "Incompatible With Secondary Licenses", as defined
REM by the Mozilla Public License, version 2.0.

start javaw -Xmx256M -jar "%~p0GURPS Character Sheet.app\Contents\Java\gcs-APP_VERSION.jar" %*
