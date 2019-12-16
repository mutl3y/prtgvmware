go build -o prtgvmware.exe .
cp prtgvmware.exe "c:\Program Files (x86)\PRTG Network Monitor\Custom Sensors\EXEXML"
prtgvmware.exe template -t PRTG
cp prtgvmware.odt "c:\Program Files (x86)\PRTG Network Monitor\devicetemplates"
