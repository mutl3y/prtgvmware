go build -o prtgvmware.exe -ldflags="-s -w" .
upx -v prtgvmware.exe
prtgvmware.exe genDocs
cp prtgvmware.exe "c:\Program Files (x86)\PRTG Network Monitor\Custom Sensors\EXEXML"
prtgvmware.exe dynamicTemplates -U https://192.168.0.201/sdk -u prtg@heynes.local -p .l3tm31n -t PRTG
cp prtgvmware.odt "c:\Program Files (x86)\PRTG Network Monitor\devicetemplates"
del prtgvmware.odt

env GOOS=linux GOARCH=amd64 go build -o prtgvmware .
scp prtgvmware prtg@192.168.0.29:/home/prtg