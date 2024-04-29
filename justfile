test:
  go test ./db/./... ./scan/./...

integration:
    INTEGRATION=1 go test ./db/./... ./scan/./...

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

licenses:
  go-licenses report ./exporter/ --template build/licenses.tpl 
