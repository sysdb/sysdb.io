all: sysdb.io

sysdb.io: src/sysdb.io.go
	go build $<

install: sysdb.io
	./sysdb.io --force --output /var/www/sysdb.io
	cp -r static/* /var/www/sysdb.io

man-latest:
	# TODO: bootstrap sysdb build system if needed
	rm -f src/sysdb/doc/*.html
	make -C src/sysdb/doc/ ADOCFLAGS="--no-header-footer"
	mkdir -p templates/manpages/latest
	cp src/sysdb/doc/*.html templates/manpages/latest

man-install: man-latest
	./sysdb.io --configfile manpages-latest.conf --force --output /var/www/sysdb.io
