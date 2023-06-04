sudo -u postgres psql postgres -c 'DROP OWNED BY usertest'
sudo -u postgres dropdb usertest
sudo -u postgres psql postgres -c 'DROP ROLE usertest'
sudo -u postgres psql postgres -c "CREATE ROLE usertest WITH LOGIN PASSWORD '${CACHE_DB_PASSWD}' CREATEDB CREATEROLE"
sudo -u postgres createdb usertest
sudo -u postgres psql usertest -f ./tables.sql usertest

