
.\bpp.cmd

copy *json deploy
cd deploy
del wheatley.zip
compress-archive -path .\* -destinationpath wheatley.zip

webjobdeploy.exe -deploy azurefunction -zipfilename .\wheatley.zip -appServiceName afken1 
cd ..

