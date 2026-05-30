set dotenv-required := true
set dotenv-load := true
reporter := "console"

list:
	go run download/cmd/download.go --latest

download:
	go run download/cmd/download.go --download --download-count 5

test:
  DYLD_LIBRARY_PATH=`pwd`/ go test ./scan/./... ./reader/./...

coverage:
  DYLD_LIBRARY_PATH=`pwd`/ go test -cover ./scan/./... ./reader/./...

coverageprofile:
  DYLD_LIBRARY_PATH=`pwd`/ go test -coverprofile=coverage.out ./scan/./... ./reader/./...
  go tool cover -html=coverage.out

vet:
  go vet ./reader/... ./scan/... ./exporter/...

dupcheck:
    npx jscpd --reporters {{reporter}} --min-lines 10 .

check: vet dupcheck test
  echo "Checked"

integration:
    INTEGRATION=1 DYLD_LIBRARY_PATH=`pwd`/ go test ./scan/./...

testdata:
  mkdir -p scan/test/testdata/firstscan
  mkdir -p scan/test/testdata/secondscan
  mkdir -p scan/test/testdata/creators

  cp ~/Documents/2000AD/2000AD\ 1999\ \(1977\).pdf scan/test/testdata/firstscan/
  cp ~/Documents/2000AD/2000AD\ 2300\ \(1977\).pdf scan/test/testdata/firstscan/

  cp ~/Documents/2000AD/2000AD\ 1999\ \(1977\).pdf scan/test/testdata/secondscan/
  cp ~/Documents/2000AD/2000AD\ 2300\ \(1977\).pdf scan/test/testdata/secondscan/
  cp ~/Documents/2000AD/2000AD\ 2301\ \(1977\).pdf scan/test/testdata/secondscan/

  cp ~/Documents/2000AD/2000AD\ 1999\ \(1977\).pdf scan/test/testdata/creators/
  cp ~/Documents/2000AD/2000AD\ 2183\ \(1977\).pdf scan/test/testdata/creators/
  cp ~/Documents/2000AD/2000AD\ 2300\ \(1977\).pdf scan/test/testdata/creators/
  cp ~/Documents/2000AD/2000AD\ 2215\ \(1977\).pdf scan/test/testdata/creators/
  cp ~/Documents/2000AD/2000AD\ 2272\ \(1977\).pdf scan/test/testdata/creators/
  cp ~/Documents/2000AD/2000AD\ 2317\ \(1977\).pdf scan/test/testdata/creators/
  cp ~/Documents/2000AD/2000AD\ 2337\ \(1977\).pdf scan/test/testdata/creators/
  cp ~/Documents/2000AD/2000AD\ 2348\ \(1977\).pdf scan/test/testdata/creators/

reader:
  DYLD_LIBRARY_PATH="`pwd`/" HOST=:8083 DATABASE_PATH=./reader/test.db SCAN_DIRS=./test_documents SCAN_ON_STARTUP=false gow -c run reader/cmd/main.go

fullreader:
  DYLD_LIBRARY_PATH="`pwd`/" HOST=:8083 LOG_LEVEL=info DATABASE_PATH=./reader/reader.db SCAN_DIRS=/Users/ross/Documents/2000AD SCAN_ON_STARTUP=false gow -c run reader/cmd/main.go

licenses:
    go-licenses report ./exporter/ --template build/licenses.tpl

# Build a Docker image for a given component (downloader|reader) and arch (amd64|arm64).
# Outputs a portable tarball: <component>-<arch>.tar
docker-build component arch:
    docker build --platform linux/{{arch}} \
        -t progger-{{component}}:latest \
        -f build/package/{{component}}/Dockerfile .
    docker save -o {{component}}-{{arch}}.tar progger-{{component}}:latest
