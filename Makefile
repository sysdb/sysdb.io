.SUFFIXES: .txt .html

NEWS = \
		news/news.html \
		news/release-0.1.0.html \
		news/release-0.2.0.html \
		news/release-0.3.0.html \
		news/release-0.4.0.html \
		news/sysdb.go.html

all: sysdb.io $(NEWS)

all-install: install man-install

sysdb.io: src/sysdb.io.go
	GOPATH="$(CURDIR)" go build $<

newsconv: src/newsconv.go
	GOPATH="$(CURDIR)" go build $<

news/news.html: news/*.txt
news/release-0.1.0.html: news/release-0.1.0.txt
news/release-0.2.0.html: news/release-0.2.0.txt
news/release-0.3.0.html: news/release-0.3.0.txt
news/release-0.4.0.html: news/release-0.4.0.txt
news/sysdb.go.html: news/sysdb.go.txt

install: sysdb.io $(NEWS)
	./sysdb.io --force --output /var/www/sysdb.io
	rsync -vax static/ /var/www/sysdb.io

man-head:
	# TODO: bootstrap sysdb build system if needed
	rm -f src/sysdb/doc/*.html
	make -C src/sysdb/doc/ ADOCFLAGS="-f $(CURDIR)/asciidoc.conf"
	mkdir -p templates/manpages/head
	cp src/sysdb/doc/*.html templates/manpages/head

man-install: man-head
	./sysdb.io --configfile manpages-head.conf --force --output /var/www/sysdb.io

.txt.html:
	asciidoc -b html5 -d article --no-header-footer $<

