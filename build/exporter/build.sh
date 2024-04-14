#! /bin/sh

go build -o progger ../../exporter/cmd/exporter.go
fyne package -os darwin --exe progger --name progger --icon Icon.png
mkdir progger.app/Contents/Frameworks
cp /opt/pdfium/lib/libpdfium.dylib progger.app/Contents/Frameworks

install_name_tool -id "@rpath/libpdfium.dylib" progger.app/Contents/Frameworks/libpdfium.dylib
install_name_tool -change ./libpdfium.dylib "@loader_path/../Frameworks/libpdfium.dylib" progger.app/Contents/MacOS/progger
install_name_tool -add_rpath "@loader_path/../Frameworks/" progger.app/Contents/MacOS/progger

create-dmg progger.dmg progger.app/
