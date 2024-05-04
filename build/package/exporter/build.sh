#! /bin/sh

if [ -f progger.dmg ]; then
  rm progger.dmg
fi 

echo "Building executable..."
go build -o progger ../../../exporter/cmd/exporter.go

echo "Building fyne package..."
fyne package -os darwin --exe progger --name progger --icon Icon.png

if [ ! -d progger.app ]; then
  echo "No progger.app directory! Exiting"
  exit
fi

if [ ! -d progger.app/Contents/Frameworks ]; then
  mkdir progger.app/Contents/Frameworks
fi

echo "Copying libpdfium into place..."
cp /opt/pdfium/lib/libpdfium.dylib progger.app/Contents/Frameworks
install_name_tool -id "@rpath/libpdfium.dylib" progger.app/Contents/Frameworks/libpdfium.dylib
install_name_tool -change ./libpdfium.dylib "@loader_path/../Frameworks/libpdfium.dylib" progger.app/Contents/MacOS/progger
install_name_tool -add_rpath "@loader_path/../Frameworks/" progger.app/Contents/MacOS/progger

echo "Creating dmg..."
create-dmg \
  --icon-size 80 \
  --icon "Progger.app" 125 175 \
  --window-size 500 320 \
  --app-drop-link 375 175 \
  progger.dmg ./progger.app/
