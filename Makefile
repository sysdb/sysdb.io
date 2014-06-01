all: sysdb.io

sysdb.io: src/sysdb.io.go
	go build $<

install: sysdb.io
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
