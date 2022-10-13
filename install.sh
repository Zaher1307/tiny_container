wget https://busybox.net/downloads/binaries/1.35.0-x86_64-linux-musl/busybox
chmod a+x busybox
mkdir -p root-fs/bin root-fs/proc root-fs/dev
./busybox --install root-fs/bin
rm busybox
go build -o container container.go
