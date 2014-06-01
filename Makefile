.SUFFIXES: .txt .html

NEWS = \
		adoc/news/release-0.1.0.html

all: sysdb.io $(REL_NOTES)

sysdb.io: src/sysdb.io.go
	go build $<

adoc/news/release-0.1.0.html: adoc/news/release-0.1.0.txt

install: sysdb.io $(NEWS)
	./sysdb.io --force --output /var/www/sysdb.io
	cp -r static/* /var/www/sysdb.io

man-head:
	# TODO: bootstrap sysdb build system if needed
	rm -f src/sysdb/doc/*.html
	make -C src/sysdb/doc/ ADOCFLAGS="-f $(CURDIR)/asciidoc.conf"
	mkdir -p templates/manpages/head
	cp src/sysdb/doc/*.html templates/manpages/head

man-install: man-head
	./sysdb.io --configfile manpages-head.conf --force --output /var/www/sysdb.io

.txt.html:
	asciidoc -b xhtml11 -d article --no-header-footer $<

