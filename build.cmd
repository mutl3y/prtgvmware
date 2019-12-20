go build -o prtgvmware.exe .
cp prtgvmware.exe "c:\Program Files (x86)\PRTG Network Monitor\Custom Sensors\EXEXML"
prtgvmware.exe dynamicTemplates -U https://192.168.0.201/sdk -u prtg@heynes.local -p .l3tm31n -t PRTG
cp prtgvmware.odt "c:\Program Files (x86)\PRTG Network Monitor\devicetemplates"
del prtgvmware.odt