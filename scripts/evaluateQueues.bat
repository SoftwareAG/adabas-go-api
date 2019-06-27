@echo off

echo "Adabas status:"
call C:\SAG\103oct2018_SIC\Adabas\INSTALL\adaenv.bat

FOR %%D IN (23,24) DO (
   adaopr db=%%D disp=uq
   adaopr db=%%D disp=com
   adaopr db=%%D reset=com
   adaopr db=%%D stop=1-10000000
)
