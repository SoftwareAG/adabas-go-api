@echo off
rem 
rem Copyright © 2018-2022 Software AG, Darmstadt, Germany and/or its licensors
rem
rem SPDX-License-Identifier: Apache-2.0
rem
rem   Licensed under the Apache License, Version 2.0 (the "License");
rem   you may not use this file except in compliance with the License.
rem   You may obtain a copy of the License at
rem
rem       http://www.apache.org/licenses/LICENSE-2.0
rem
rem   Unless required by applicable law or agreed to in writing, software
rem   distributed under the License is distributed on an "AS IS" BASIS,
rem   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
rem   See the License for the specific language governing permissions and
rem   limitations under the License.
rem 

echo "Cleaning ..."

set DIR=%~dp0\..

echo "Deleting directory %DIR%\bin"
if exist %DIR%\bin rmdir %DIR%\bin /s /q

echo "Deleting directory %DIR%\test"
if exist %DIR%\test rmdir %DIR%\test /s /q

rem go clean