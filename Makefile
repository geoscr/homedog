npm-import:
	docker cp homedog-webpack:/app/package.json $(JS_PACKAGEJSON)

clean-db:
	docker exec -ti db.homedog  psql -U postgres postgres -c "delete from users; delete from posts;"
