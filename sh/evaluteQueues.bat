@echo off

call C:\SAG\103oct2018_SIC\Adabas\INSTALL\adaenv.bat

FOR i IN (23,24,25) DO (
   adaopr db=$i disp=uq
   adaopr db=$i disp=com
   adaopr db=$i reset=com
)
