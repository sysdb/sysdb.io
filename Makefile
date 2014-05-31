all: sysdb.io

sysdb.io: src/sysdb.io.go
	go build $<

install: all
	./sysdb.io --force --output /var/www/sysdb.io
	cp -r static/* /var/www/sysdb.io
