mkdir ../dist/nging_${GOOS}_${GOARCH}
go build -tags "bindata sqlite${BUILDTAGS}" -o ../dist/nging_${GOOS}_${GOARCH}/nging ..
mkdir ../dist/nging_${GOOS}_${GOARCH}/data
mkdir ../dist/nging_${GOOS}_${GOARCH}/data/logs
cp -R ../data/ip2region ../dist/nging_${GOOS}_${GOARCH}/data/ip2region


mkdir ../dist/nging_${GOOS}_${GOARCH}/config
mkdir ../dist/nging_${GOOS}_${GOARCH}/config/vhosts

#cp -R ../config/config.yaml ../dist/nging_${GOOS}_${GOARCH}/config/config.yaml
cp -R ../config/config.yaml.sample ../dist/nging_${GOOS}_${GOARCH}/config/config.yaml.sample
cp -R ../config/install.sql ../dist/nging_${GOOS}_${GOARCH}/config/install.sql
cp -R ../config/ua.txt ../dist/nging_${GOOS}_${GOARCH}/config/ua.txt

if [ $GOOS = "windows" ]; then
    cp -R ../support/sqlite3_${GOARCH}.dll ../dist/nging_${GOOS}_${GOARCH}/sqlite3_${GOARCH}.dll
	export archiver_extension=tar.bz2
else
	export archiver_extension=zip
fi

cp -R ../dist/default/* ../dist/nging_${GOOS}_${GOARCH}/

archiver make ../dist/nging_${GOOS}_${GOARCH}.${archiver_extension} ../dist/nging_${GOOS}_${GOARCH}/

rm -rf ../dist/nging_${GOOS}_${GOARCH}
