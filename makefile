build:
	clean create-dist copy-non-ts webpack

clean:
	rm -rf web/dist

create-dist:
	mkdir -p web/dist

copy-non-ts:
	find web/src -type f ! -name "*.ts" -exec cp --parents {} web/dist \;

webpack:
	npx webpack